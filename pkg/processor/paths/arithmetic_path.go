package paths

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"go/token"
	"strings"
)

type Path interface {
	ReadFrom(rootNamespace namespaces.Namespace) (namespaces.Namespace, rookoutErrors.RookoutError)
	WriteTo(rootNamespace namespaces.Namespace, value namespaces.Namespace) rookoutErrors.RookoutError
}

type ArithmeticPath struct {
	operations      node
	writeOperations []pathOperation
	negation        bool
}

func NewArithmeticPath(configuration interface{}) (*ArithmeticPath, rookoutErrors.RookoutError) {
	var configStr string
	switch configuration.(type) {
	case string:
		configStr = configuration.(string)
	case map[string]interface{}:
		configMap := configuration.(map[string]interface{})
		path, ok := configMap["path"]
		if !ok || path == nil {
			return nil, rookoutErrors.NewRookInvalidArithmeticPathException(configuration, nil)
		}

		configStr, ok = path.(string)
		if !ok {
			return nil, rookoutErrors.NewRookInvalidArithmeticPathException(configuration, nil)
		}
	default:
		return nil, rookoutErrors.NewRookInvalidArithmeticPathException(configuration, nil)
	}

	arithmeticPath := &ArithmeticPath{
		negation: false,
	}

	if strings.HasPrefix(configStr, "NOT(") && strings.HasSuffix(configStr, ")") {
		arithmeticPath.negation = true
		configStr = configStr[len("NOT(") : len(configStr)-len(")")]
	}

	mapsActions, err := newMapsActions()
	if err != nil {
		return nil, err
	}
	arithmeticPath.operations, err = mapsActions.Parse(configStr)
	if err != nil {
		return nil, err
	}
	arithmeticPath.writeOperations = mapsActions.GetWriteOperations()

	return arithmeticPath, nil
}

func (a ArithmeticPath) ReadFrom(rootNamespace namespaces.Namespace) (namespaces.Namespace, rookoutErrors.RookoutError) {
	var res valueNode
	var err rookoutErrors.RookoutError
	if operations, ok := a.operations.(executableNode); ok {
		res, err = operations.Execute(rootNamespace)
		if err != nil {
			return nil, err
		}
	} else {
		res = a.operations.(valueNode)
	}

	switch res.(type) {
	case *namespaceResultNode:
		return res.(*namespaceResultNode).namespace, nil
	default:
		val := res.Interface()
		if _, ok := val.(bool); ok && a.negation {
			val = !val.(bool)
		}
		return namespaces.NewGoObjectNamespace(val), nil
	}
}

func (a ArithmeticPath) WriteTo(namespace namespaces.Namespace, value namespaces.Namespace) rookoutErrors.RookoutError {
	if operations, ok := a.operations.(executableNode); ok {
		_, _ = operations.Execute(namespace)
	}

	var rookErr rookoutErrors.RookoutError
	for _, op := range a.writeOperations[:len(a.writeOperations)-1] {
		namespace, rookErr = op.Read(namespace, true)
		if rookErr != nil {
			return rookErr
		}
	}

	lastOp := a.writeOperations[len(a.writeOperations)-1]
	lastWriteOp, ok := lastOp.(writeOperation)
	if !ok {
		msg := fmt.Sprintf("last operation is not writable: %v (%#v)", lastOp, lastOp)
		return rookoutErrors.NewRookInvalidArithmeticPathException(msg, nil)
	}
	rookErr = lastWriteOp.Write(namespace, value)
	if rookErr != nil {
		return rookErr
	}

	return nil
}

type optLevel uint32


const (
	MULDIV        optLevel = 0
	ADDSUB        optLevel = 1
	COMP          optLevel = 2
	IN            optLevel = 3
	AND           optLevel = 4
	OR            optLevel = 5
	NUM_OF_LEVELS optLevel = 6
)

var optLevelToStr = map[optLevel][]string{
	MULDIV: {"*", "/"},
	ADDSUB: {"+", "-"},
	COMP:   {"<=", ">=", "!=", "=", "==", ">", "<", "LT", "GT", "LE", "GE", "EQ", "NE", "lt", "gt", "le", "ge", "eq", "ne"},
	IN:     {"IN", "in"},
	AND:    {"&&"},
	OR:     {"||"},
}

var ArithmeticExpressions = map[string]string{
	"NE": "!=",
	"=":  "==",
	"EQ": "==",
	"LT": "<",
	"GT": ">",
	"GE": ">=",
	"LE": "<=",
	"IN": "in",

	"AND": "&&",
	"OR":  "||",
}

var strToToken = map[string]token.Token{
	"+":  token.ADD,
	"-":  token.SUB,
	"*":  token.MUL,
	"/":  token.QUO,
	"&&": token.LAND,
	"||": token.LOR,
	"<=": token.LEQ,
	">=": token.GEQ,
	"!=": token.NEQ,
	"=":  token.EQL,
	"==": token.EQL,
	">":  token.GTR,
	"<":  token.LSS,
}
