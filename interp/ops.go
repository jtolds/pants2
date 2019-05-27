package interp

import (
	"fmt"

	"github.com/jtolds/pants2/ast"
	"github.com/jtolds/pants2/lib/big"
)

func equalityTest(left, right Value) bool {
	if typename(left) != typename(right) {
		return false
	}
	switch left.(type) {
	case ValNumber:
		x, y := left.(ValNumber).Val, right.(ValNumber).Val
		return x.Cmp(&y) == 0
	case ValString:
		return left.(ValString).Val == right.(ValString).Val
	case ValBool:
		return left.(ValBool).Val == right.(ValBool).Val
	default:
		return false // TODO: throw an error about comparing funcs or procs?
	}
}

type typesym int

var (
	typesymNum typesym = 0
	typesymStr typesym = 1
)

func (t typesym) String() string {
	switch t {
	case typesymNum:
		return "number"
	case typesymStr:
		return "string"
	default:
		return "unknown"
	}
}

func typename(val Value) typesym {
	switch val.(type) {
	case ValNumber:
		return typesymNum
	case ValString:
		return typesymStr
	default:
		panic(fmt.Sprintf("type unimplemented: %#v", val))
	}
}

var zero big.Rat

func unsupportedOp(t *ast.Token, op string, left, right Value) error {
	return NewRuntimeError(t,
		"unsupported operation: %s %s %s",
		typename(left), op, typename(right))
}

func Add(t *ast.Token, left, right Value) (v Value, err error) {
	switch x := left.(type) {
	case ValNumber:
		y, ok := right.(ValNumber)
		if !ok {
			return nil, unsupportedOp(t, "+", left, right)
		}
		var rv ValNumber
		rv.Val.Add(&x.Val, &y.Val)
		return rv, nil
	case ValString:
		y, ok := right.(ValString)
		if !ok {
			return nil, unsupportedOp(t, "+", left, right)
		}
		return ValString{Val: x.Val + y.Val}, nil
	default:
		return nil, unsupportedOp(t, "+", left, right)
	}
}

func Subtract(t *ast.Token, left, right Value) (v Value, err error) {
	x, ok1 := left.(ValNumber)
	y, ok2 := right.(ValNumber)
	if !ok1 || !ok2 {
		return nil, unsupportedOp(t, "-", left, right)
	}
	var rv ValNumber
	rv.Val.Sub(&x.Val, &y.Val)
	return rv, nil
}

func Multiply(t *ast.Token, left, right Value) (v Value, err error) {
	x, ok1 := left.(ValNumber)
	y, ok2 := right.(ValNumber)
	if !ok1 || !ok2 {
		return nil, unsupportedOp(t, "*", left, right)
	}
	var rv ValNumber
	rv.Val.Mul(&x.Val, &y.Val)
	return rv, nil
}

func Divide(t *ast.Token, left, right Value) (v Value, err error) {
	x, ok1 := left.(ValNumber)
	y, ok2 := right.(ValNumber)
	if !ok1 || !ok2 {
		return nil, unsupportedOp(t, "/", left, right)
	}
	if zero.Cmp(&y.Val) == 0 {
		return nil, NewRuntimeError(t, "Division by zero")
	}
	var rv ValNumber
	var im big.Rat
	im.Inv(&y.Val)
	rv.Val.Mul(&x.Val, &im)
	return rv, nil
}

func Modulo(t *ast.Token, left, right Value) (v Value, err error) {
	x, ok1 := left.(ValNumber)
	y, ok2 := right.(ValNumber)
	if !ok1 || !ok2 {
		return nil, unsupportedOp(t, "%", left, right)
	}
	if zero.Cmp(&y.Val) == 0 {
		return nil, NewRuntimeError(t, "Division by zero")
	}
	leftdenom := x.Val.Denom()
	rightdenom := y.Val.Denom()
	if !leftdenom.IsInt64() || leftdenom.Int64() != 1 ||
		!rightdenom.IsInt64() || rightdenom.Int64() != 1 {
		return nil, NewRuntimeError(t, "Modulo only works on integers")
	}
	var rv ValNumber
	var im big.Int
	im.Mod(x.Val.Num(), y.Val.Num())
	rv.Val.SetInt(&im)
	return rv, nil
}

func LessThan(t *ast.Token, left, right Value) (v Value, err error) {
	switch x := left.(type) {
	case ValNumber:
		y, ok := right.(ValNumber)
		if !ok {
			return nil, unsupportedOp(t, "<", left, right)
		}
		return ValBool{Val: x.Val.Cmp(&y.Val) < 0}, nil
	case ValString:
		y, ok := right.(ValString)
		if !ok {
			return nil, unsupportedOp(t, "<", left, right)
		}
		return ValBool{Val: x.Val < y.Val}, nil
	default:
		return nil, unsupportedOp(t, "<", left, right)
	}
}

func LessThanEqual(t *ast.Token, left, right Value) (v Value, err error) {
	switch x := left.(type) {
	case ValNumber:
		y, ok := right.(ValNumber)
		if !ok {
			return nil, unsupportedOp(t, "<=", left, right)
		}
		return ValBool{Val: x.Val.Cmp(&y.Val) <= 0}, nil
	case ValString:
		y, ok := right.(ValString)
		if !ok {
			return nil, unsupportedOp(t, "<=", left, right)
		}
		return ValBool{Val: x.Val <= y.Val}, nil
	default:
		return nil, unsupportedOp(t, "<=", left, right)
	}
}

func GreaterThan(t *ast.Token, left, right Value) (v Value, err error) {
	switch x := left.(type) {
	case ValNumber:
		y, ok := right.(ValNumber)
		if !ok {
			return nil, unsupportedOp(t, ">", left, right)
		}
		return ValBool{Val: x.Val.Cmp(&y.Val) > 0}, nil
	case ValString:
		y, ok := right.(ValString)
		if !ok {
			return nil, unsupportedOp(t, ">", left, right)
		}
		return ValBool{Val: x.Val > y.Val}, nil
	default:
		return nil, unsupportedOp(t, ">", left, right)
	}
}

func GreaterThanEqual(t *ast.Token, left, right Value) (v Value, err error) {
	switch x := left.(type) {
	case ValNumber:
		y, ok := right.(ValNumber)
		if !ok {
			return nil, unsupportedOp(t, ">=", left, right)
		}
		return ValBool{Val: x.Val.Cmp(&y.Val) >= 0}, nil
	case ValString:
		y, ok := right.(ValString)
		if !ok {
			return nil, unsupportedOp(t, ">=", left, right)
		}
		return ValBool{Val: x.Val >= y.Val}, nil
	default:
		return nil, unsupportedOp(t, ">=", left, right)
	}
}

func Equal(t *ast.Token, left, right Value) (v Value, err error) {
	return ValBool{Val: equalityTest(left, right)}, nil
}

func NotEqual(t *ast.Token, left, right Value) (v Value, err error) {
	return ValBool{Val: !equalityTest(left, right)}, nil
}

var operations = map[string]func(t *ast.Token, left, right ast.Value) (ast.Value, error){
	"+":  Add,
	"-":  Subtract,
	"*":  Multiply,
	"/":  Divide,
	"%":  Modulo,
	"<":  LessThan,
	"<=": LessThanEqual,
	">":  GreaterThan,
	">=": GreaterThanEqual,
	"==": Equal,
	"!=": NotEqual,
}
