package interp

import (
	"fmt"

	"github.com/jtolds/pants2/ast"
	"github.com/jtolds/pants2/lib/big"
)

func equalityTest(expr ast.Expr, left, right Value) bool {
	if typename(left) != typename(right) {
		return false
	}
	switch left.(type) {
	case *ValNumber:
		return left.(*ValNumber).Val.Cmp(&right.(*ValNumber).Val) == 0
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
	case *ValNumber:
		return typesymNum
	case ValString:
		return typesymStr
	default:
		panic(fmt.Sprintf("type unimplemented: %#v", val))
	}
}

type opkey struct {
	op          string
	left, right typesym
}

var zero big.Rat

var operations = map[opkey]func(t *ast.Token, left, right Value) (
	Value, error){
	opkey{"+", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		var rv ValNumber
		rv.Val.Add(&left.(*ValNumber).Val, &right.(*ValNumber).Val)
		return &rv, nil
	},
	opkey{"-", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		var rv ValNumber
		rv.Val.Sub(&left.(*ValNumber).Val, &right.(*ValNumber).Val)
		return &rv, nil
	},
	opkey{"*", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		var rv ValNumber
		rv.Val.Mul(&left.(*ValNumber).Val, &right.(*ValNumber).Val)
		return &rv, nil
	},
	opkey{"/", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		if zero.Cmp(&right.(*ValNumber).Val) == 0 {
			return nil, NewRuntimeError(t, "Division by zero")
		}
		var rv ValNumber
		var im big.Rat
		im.Inv(&right.(*ValNumber).Val)
		rv.Val.Mul(&left.(*ValNumber).Val, &im)
		return &rv, nil
	},
	opkey{"%", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		if zero.Cmp(&right.(*ValNumber).Val) == 0 {
			return nil, NewRuntimeError(t, "Division by zero")
		}
		leftdenom := left.(*ValNumber).Val.Denom()
		rightdenom := right.(*ValNumber).Val.Denom()
		if !leftdenom.IsInt64() || leftdenom.Int64() != 1 ||
			!rightdenom.IsInt64() || rightdenom.Int64() != 1 {
			return nil, NewRuntimeError(t, "Modulo only works on integers")
		}
		var rv ValNumber
		var im big.Int
		im.Mod(left.(*ValNumber).Val.Num(), right.(*ValNumber).Val.Num())
		rv.Val.SetInt(&im)
		return &rv, nil
	},
	opkey{"+", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValString{Val: left.(ValString).Val + right.(ValString).Val}, nil
	},
	opkey{"<", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(*ValNumber).Val.Cmp(&right.(*ValNumber).Val) < 0}, nil
	},
	opkey{"<=", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(*ValNumber).Val.Cmp(&right.(*ValNumber).Val) <= 0}, nil
	},
	opkey{">", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(*ValNumber).Val.Cmp(&right.(*ValNumber).Val) > 0}, nil
	},
	opkey{">=", typename(&ValNumber{}), typename(&ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(*ValNumber).Val.Cmp(&right.(*ValNumber).Val) >= 0}, nil
	},
	opkey{"<", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValString).Val < right.(ValString).Val}, nil
	},
	opkey{"<=", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValString).Val <= right.(ValString).Val}, nil
	},
	opkey{">", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValString).Val >= right.(ValString).Val}, nil
	},
	opkey{">=", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValString).Val >= right.(ValString).Val}, nil
	},
}
