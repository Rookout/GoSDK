package paths

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"github.com/Rookout/GoSDK/pkg/utils"
	"github.com/yhirose/go-peg"
	"strconv"
	"strings"
)

type mapsActions struct {
	operations []pathOperation
	parser     *peg.Parser
}

func newMapsActions() (m *mapsActions, err rookoutErrors.RookoutError) {
	m = &mapsActions{}
	m.parser, err = m.buildParser()
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *mapsActions) buildParser() (*peg.Parser, rookoutErrors.RookoutError) {
	
	parser, err := peg.NewParser(`
		COMPARISON_EXPRESSION   <-  EXPRESSION (COMPARISON EXPRESSION)*
		EXPRESSION   			<-  ATOM (COMPARISON ATOM)*
		
		#For some reason the whitespaces are ignored - causing parsing errors
		LIST		 <-  [ "["] ([ \t]*ATOM ([ \t]*[,][ \t]* ATOM)*)? ["\]" ][ \t]*
        ATOM         <-  NULL / FLOAT / NUMBER / '(' EXPRESSION ')' / '(' COMPARISON_EXPRESSION ')' / CHAR / STRING / APOSTROPHE_STRING / BOOL / LIST / NAMESPACE
		COMPARISON   <-  <'+'/'-'/'/'/'*'/'<='/'>='/'!='/'=='/"="/'>'/'<'/"LT"/"GT"/"LE"/"GE"/"EQ"/"NE"/"lt"/"gt"/"le"/"ge"/"eq"/"ne"/"in"/"IN"/"or"/"OR"/"||"/"and"/"AND"/"&&">
		NUMBER       <-  < [-]?[0-9]+ >
        FLOAT        <-  < [-]?[0-9]+([.][0-9]+)+ >
		#The only value we are not capturing here is '"' (")
		STRING       <-  < [\"] [ !#-~]* [\"] >
		APOSTROPHE_STRING <- < ['] [ !-&(-~]* ['] >
		BOOL		 <-  'True' / 'False' / 'true' / 'false'
		CHAR       	 <-  < ['] [ !-~] ['] >
		NULL		 <-  'None' / 'nil' / 'null' / 'undefined'

		METHOD_ACCESS <- < ([a-zA-Z0-9_]+) > '(' < (ATOM ([,] ATOM)*)? > ')'
		VARIABLE_ACCESS <- < ([a-zA-Z0-9_]+) >
		MAP_ACCESS <- '[' < (ATOM) > ']'
		NAMESPACE    <-   (METHOD_ACCESS / VARIABLE_ACCESS) (MAP_ACCESS / '.' METHOD_ACCESS / '.' VARIABLE_ACCESS)*

		%whitespace  <-  [ \t]*
		---
        # Expression parsing 1 is the least important
		%expr  = EXPRESSION
		%comparison = L + -  # level 1
		%comparison = L * /  # level 2
		%comparison = L > < # level 3
    `)
	if nil != err {
		return nil, rookoutErrors.NewArithmeticPathException(err)
	}
	grammar := parser.Grammar

	grammar["COMPARISON_EXPRESSION"].Action = m.makeComparisonExpression
	grammar["EXPRESSION"].Action = m.makeComparisonExpression
	grammar["COMPARISON"].Action = m.makeComparison
	grammar["LIST"].Action = m.makeList
	grammar["BOOL"].Action = m.makeBoolean
	grammar["NAMESPACE"].Action = m.makeNamespace
	grammar["VARIABLE_ACCESS"].Action = m.makeAttributeOperation
	grammar["METHOD_ACCESS"].Action = m.makeMethodOperation
	grammar["MAP_ACCESS"].Action = m.makeLookupOperation
	grammar["FLOAT"].Action = m.makeFloat
	grammar["NUMBER"].Action = m.makeNumber
	grammar["STRING"].Action = m.makeString
	grammar["APOSTROPHE_STRING"].Action = m.makeString
	grammar["CHAR"].Action = m.makeString
	grammar["NULL"].Action = m.makeNull

	return parser, nil
}

func (m *mapsActions) Parse(path string) (node, rookoutErrors.RookoutError) {
	operations, err := m.parser.ParseAndGetValue(path, nil)
	if err != nil {
		return nil, rookoutErrors.NewRookInvalidArithmeticPathException(path, err)
	}
	return operations.(node), nil
}

func (m *mapsActions) GetWriteOperations() []pathOperation {
	return m.operations
}

func (m *mapsActions) makeComparisonExpression(values *peg.Values, _ peg.Any) (peg.Any, error) {
	if values.Len() == 2 {
		return nil, rookoutErrors.NewRookInvalidArithmeticPathException(values.SS, nil)
	}

	elements := make([]node, 0, len(values.Vs))
	for _, n := range values.Vs {
		element, ok := n.(node)
		if !ok {
			return nil, rookoutErrors.NewRookInvalidArithmeticPathException(values.SS, nil)
		}
		elements = append(elements, element)
	}

	for len(elements) > 1 {
		elementsChanged := false

		for level := optLevel(0); level < NUM_OF_LEVELS && !elementsChanged; level++ {
			for i := 1; i < len(elements); i += 2 {
				left := elements[i-1]
				opt, ok := elements[i].(*optNode)
				if !ok {
					return nil, rookoutErrors.NewRookInvalidArithmeticPathException(values.SS, nil)
				}
				right := elements[i+1]

				if opt.level != level {
					continue
				}

				compNode := newComparisonNode(left, opt, right)
				elements[i-1] = compNode
				elements = append(elements[:i], elements[i+2:]...)

				elementsChanged = true
			}
		}
	}

	return elements[0], nil
}

func (m *mapsActions) makeNamespace(values *peg.Values, _ peg.Any) (peg.Any, error) {
	var operations []pathOperation
	for _, rawOperation := range values.Vs {
		if operation, ok := rawOperation.(pathOperation); ok {
			operations = append(operations, operation)
		} else {
			msg := fmt.Sprintf("parsing %v in (%s)", rawOperation, values.SS)
			return nil, rookoutErrors.NewRookInvalidArithmeticPathException(msg, nil)
		}
	}

	m.operations = append(m.operations, operations...)
	return newNamespaceNode(operations), nil
}

func (m *mapsActions) makeList(values *peg.Values, _ peg.Any) (peg.Any, error) {
	a := make([]node, 0, len(values.Vs))
	for _, s := range values.Vs {
		a = append(a, s.(node))
	}

	return newListNode(a), nil
}

func (m *mapsActions) makeComparison(values *peg.Values, _ peg.Any) (peg.Any, error) {
	opt := strings.Replace(values.Token(), " ", "", -1)
	return newOpt(opt)
}

func (m *mapsActions) makeBoolean(values *peg.Values, _ peg.Any) (peg.Any, error) {
	val := strings.ToLower(utils.ReplaceAll(values.Token(), " ", ""))
	if val == "true" {
		return newValueNode(true), nil
	}
	return newValueNode(false), nil
}

func (m *mapsActions) makeString(values *peg.Values, _ peg.Any) (peg.Any, error) {
	return newValueNode(values.Token()[1 : len(values.Token())-1]), nil
}

func (m *mapsActions) makeNumber(values *peg.Values, _ peg.Any) (peg.Any, error) {
	i, err := strconv.Atoi(values.Token())
	if err != nil {
		msg := "unable to parse int: " + values.Token()
		return nil, rookoutErrors.NewRookInvalidArithmeticPathException(msg, err)
	}
	return newValueNode(i), nil
}

func (m *mapsActions) makeFloat(values *peg.Values, _ peg.Any) (peg.Any, error) {
	f, err := strconv.ParseFloat(values.Token(), 64)
	if err != nil {
		msg := "unable to parse float: " + values.Token()
		return nil, rookoutErrors.NewRookInvalidArithmeticPathException(msg, err)
	}
	return newValueNode(f), nil
}

func (m *mapsActions) makeLookupOperation(values *peg.Values, _ peg.Any) (peg.Any, error) {
	return newLookupOperation(values.Token())
}

func (m *mapsActions) makeMethodOperation(values *peg.Values, _ peg.Any) (peg.Any, error) {
	args := ""
	if len(values.Ts) >= 2 {
		args = values.Ts[1].S
		if (strings.HasSuffix(args, "'") && strings.HasPrefix(args, "'")) ||
			(strings.HasSuffix(args, "\"") && strings.HasPrefix(args, "\"")) {
			args = args[1 : len(args)-1]
		}
	}
	return newMethodOperation(values.Token(), args), nil
}

func (m *mapsActions) makeAttributeOperation(values *peg.Values, _ peg.Any) (peg.Any, error) {
	return newAttributeOperation(values.Token()), nil
}

func (m *mapsActions) makeNull(_ *peg.Values, _ peg.Any) (peg.Any, error) {
	return newNoneNode(), nil
}
