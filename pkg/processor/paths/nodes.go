package paths

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg/processor/namespaces"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
	"go/constant"
	"go/token"
	"go/types"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

type Kind = reflect.Kind

const (
	None = iota + reflect.UnsafePointer + 1
	List
	NamespaceResult
)

type node interface {
}

type executableNode interface {
	node
	Execute(namespace namespaces.Namespace) (valueNode, rookoutErrors.RookoutError)
}

type valueNode interface {
	node
	Kind() Kind
	Interface() interface{}
}

type optNode struct {
	optStr string
	level  optLevel
	token  token.Token
}

func newOpt(optStr string) (node, rookoutErrors.RookoutError) {
	o := &optNode{}
	o.optStr = optStr

	optUpper := strings.ToUpper(o.optStr)
	for key, value := range ArithmeticExpressions {
		if key == optUpper {
			o.optStr = value
			break
		}
	}

	found := false
	for i := optLevel(0); i < NUM_OF_LEVELS; i++ {
		for j := 0; j < len(optLevelToStr[i]); j++ {
			if o.optStr == optLevelToStr[i][j] {
				o.level = i
				found = true
				break
			}
		}
	}
	if !found {
		return nil, rookoutErrors.NewRookInvalidArithmeticPathException("condition could not be resolved: "+optStr, nil)
	}

	o.token, _ = strToToken[o.optStr]

	return o, nil
}

func newValueNode(val interface{}) valueNode {
	switch val.(type) {
	case nil:
		return newNoneNode()
	case *big.Rat:
		val, _ = val.(*big.Rat).Float64()
		return reflect.ValueOf(val)
	case int64:
		return reflect.ValueOf(int(val.(int64)))
	default:
		return reflect.ValueOf(val)
	}
}

func valueNodeToConstantValue(val valueNode) (constant.Value, bool) {
	switch val.Kind() {
	case NamespaceResult:
		return valueNodeToConstantValue(newValueNode(val.Interface()))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return constant.MakeInt64(val.(reflect.Value).Int()), true
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return constant.MakeUint64(val.(reflect.Value).Uint()), true
	case reflect.String:
		return constant.MakeString(val.(reflect.Value).String()), true
	case reflect.Float64, reflect.Float32:
		return constant.MakeFloat64(val.(reflect.Value).Float()), true
	case reflect.Bool:
		return constant.MakeBool(val.(reflect.Value).Bool()), true
	default:
		return nil, false
	}
}

func isOperable(a valueNode, b valueNode) (res bool) {
	defer func() {
		if v := recover(); v != nil {
			res = false
		}
	}()

	aValue := reflect.ValueOf(a.Interface())
	bValue := reflect.ValueOf(b.Interface())
	aValue.Convert(bValue.Type())
	bValue.Convert(aValue.Type())
	return true
}

func (o *optNode) evalExpression(a valueNode, b valueNode) (v valueNode, err rookoutErrors.RookoutError) {
	defer func() {
		r := recover()
		if r != nil {
			msg := fmt.Sprintf("unable to evaluate %v %s %v (%v)", a, o.optStr, b, r)
			v = nil
			err = rookoutErrors.NewRookInvalidArithmeticPathException(msg, nil)
		}
	}()

	aValue, aIsPrimitive := valueNodeToConstantValue(a)
	bValue, bIsPrimitive := valueNodeToConstantValue(b)

	if !aIsPrimitive || !bIsPrimitive || !isOperable(a, b) {
		aObj := a.Interface()
		bObj := b.Interface()

		if (aObj == namespaces.ReferenceTypeInstance && bObj != nil) || (bObj == namespaces.ReferenceTypeInstance && aObj != nil) {
			msg := "Comparison of reference types to something other than nil is not supported"
			return nil, rookoutErrors.NewRookInvalidArithmeticPathException(msg, nil)
		}

		if aObj == namespaces.StructTypeInstance || bObj == namespaces.StructTypeInstance {
			msg := "Comparison of structs is not supported"
			return nil, rookoutErrors.NewRookInvalidArithmeticPathException(msg, nil)
		}

		eq := reflect.DeepEqual(aObj, bObj)
		if (o.optStr == "==" && eq) ||
			(o.optStr == "!=" && !eq) {
			return newValueNode(true), nil
		} else if (o.optStr == "!=" && eq) ||
			(o.optStr == "==" && !eq) {
			return newValueNode(false), nil
		}

		evaluation := fmt.Sprintf("%v%s%v", a.Interface(), o.optStr, b.Interface())
		res, evalErr := types.Eval(token.NewFileSet(), nil, token.NoPos, evaluation)
		if evalErr != nil {
			msg := fmt.Sprintf("unable to evaluate %s", evaluation)
			return nil, rookoutErrors.NewRookInvalidArithmeticPathException(msg, evalErr)
		}
		return newValueNode(constant.Val(res.Value)), nil
	}

	if o.level == COMP {
		res := constant.Compare(aValue, o.token, bValue)
		return newValueNode(res), nil
	} else {
		resVal := constant.BinaryOp(aValue, o.token, bValue)
		return newValueNode(constant.Val(resVal)), nil
	}
}

func (o *optNode) inExpression(a valueNode, b valueNode) (valueNode, rookoutErrors.RookoutError) {
	switch b.Interface().(type) {
	case []interface{}:
		list := b.Interface().([]interface{})
		for _, v := range list {
			if a.Interface() == v {
				return newValueNode(true), nil
			}
		}
		return newValueNode(false), nil
	case string:
		aStr, ok := a.Interface().(string)
		if !ok {
			return newValueNode(false), nil
		}

		unquoted, err := strconv.Unquote(aStr)
		if err == nil {
			aStr = unquoted
		}

		if strings.Contains(b.Interface().(string), aStr) {
			return newValueNode(true), nil
		}
		return newValueNode(false), nil
	default:
		msg := fmt.Sprintf("can't use `in` expression on %v of type %#v", b, b)
		return nil, rookoutErrors.NewRookInvalidArithmeticPathException(msg, nil)
	}
}

func (o *optNode) Execute(a valueNode, b valueNode) (valueNode, rookoutErrors.RookoutError) {
	switch o.level {
	case MULDIV, ADDSUB, COMP, AND, OR:
		return o.evalExpression(a, b)
	case IN:
		return o.inExpression(a, b)
	}
	msg := fmt.Sprintf("invalid opt level: %d (%s)", o.level, o.optStr)
	return nil, rookoutErrors.NewRookInvalidArithmeticPathException(msg, nil)
}

type listNode struct {
	nodes []node
}

func newListNode(nodes []node) *listNode {
	l := &listNode{}
	l.nodes = nodes
	return l
}

func (l *listNode) Execute(namespace namespaces.Namespace) (valueNode, rookoutErrors.RookoutError) {
	var err rookoutErrors.RookoutError
	for i, n := range l.nodes {
		if executable, ok := n.(executableNode); ok {
			l.nodes[i], err = executable.Execute(namespace)
			if err != nil {
				return nil, err
			}
		}
	}
	return l, nil
}

func (l *listNode) Interface() interface{} {
	values := make([]interface{}, 0, len(l.nodes))
	for _, n := range l.nodes {
		if val, ok := n.(valueNode); ok {
			values = append(values, val.Interface())
		}
	}
	return values
}

func (l *listNode) Kind() Kind {
	return List
}

type comparisonNode struct {
	left  node
	opt   *optNode
	right node
}

func newComparisonNode(left node, opt *optNode, right node) node {
	return &comparisonNode{
		left:  left,
		opt:   opt,
		right: right,
	}
}

func (c *comparisonNode) Execute(namespace namespaces.Namespace) (n valueNode, err rookoutErrors.RookoutError) {
	left := c.left
	right := c.right
	if executable, ok := left.(executableNode); ok {
		left, err = executable.Execute(namespace)
		if err != nil {
			return nil, err
		}
	}
	if executable, ok := right.(executableNode); ok {
		right, err = executable.Execute(namespace)
		if err != nil {
			return nil, err
		}
	}
	return c.opt.Execute(left.(valueNode), right.(valueNode))
}

type namespaceNode struct {
	operations []pathOperation
}

func newNamespaceNode(operations []pathOperation) *namespaceNode {
	n := &namespaceNode{}
	n.operations = operations
	return n
}

func (n *namespaceNode) Execute(namespace namespaces.Namespace) (valueNode, rookoutErrors.RookoutError) {
	if namespace == nil {
		return nil, rookoutErrors.NewRookInvalidArithmeticPathException("unable to Execute namespace operations on nil", nil)
	}

	var err rookoutErrors.RookoutError
	res := namespace
	for _, path := range n.operations {
		res, err = path.Read(res, false)
		if err != nil {
			return nil, err
		}
	}

	return newNamespaceResultNode(res), nil
}

type namespaceResultNode struct {
	namespace namespaces.Namespace
}

func newNamespaceResultNode(namespace namespaces.Namespace) *namespaceResultNode {
	n := &namespaceResultNode{}
	n.namespace = namespace
	return n
}

func (n *namespaceResultNode) Kind() Kind {
	return NamespaceResult
}

func (n *namespaceResultNode) Interface() interface{} {
	return n.namespace.GetObject()
}

type noneNode struct {
}

func newNoneNode() *noneNode {
	return &noneNode{}
}

func (n *noneNode) Kind() Kind {
	return None
}

func (n *noneNode) Interface() interface{} {
	return nil
}
