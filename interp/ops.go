package interp

import (
	"fmt"
	"math/big"

	"github.com/jtolds/pants2/ast"
)

func equalityTest(expr ast.Expr, left, right Value) bool {
	if typename(left) != typename(right) {
		return false
	}
	switch left.(type) {
	case ValNumber:
		return left.(ValNumber).Val.Cmp(right.(ValNumber).Val) == 0
	case ValString:
		return left.(ValString).Val == right.(ValString).Val
	case ValBool:
		return left.(ValBool).Val == right.(ValBool).Val
	default:
		return false // TODO: throw an error about comparing funcs or procs?
	}
}

type opkey struct{ op, left, right string }

func typename(val interface{}) string { return fmt.Sprintf("%T", val) }

var operations = map[opkey]func(t *ast.Token, left, right Value) (
	Value, error){
	opkey{"+", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValNumber{
			Val: new(big.Rat).Add(left.(ValNumber).Val, right.(ValNumber).Val)}, nil
	},
	opkey{"-", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValNumber{
			Val: new(big.Rat).Sub(left.(ValNumber).Val, right.(ValNumber).Val)}, nil
	},
	opkey{"*", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValNumber{
			Val: new(big.Rat).Mul(left.(ValNumber).Val, right.(ValNumber).Val)}, nil
	},
	opkey{"/", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		if new(big.Rat).Cmp(right.(ValNumber).Val) == 0 {
			return nil, NewRuntimeError(t, "Division by zero")
		}
		return ValNumber{
			Val: new(big.Rat).Mul(left.(ValNumber).Val,
				new(big.Rat).Inv(right.(ValNumber).Val))}, nil
	},
	opkey{"%", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		if new(big.Rat).Cmp(right.(ValNumber).Val) == 0 {
			return nil, NewRuntimeError(t, "Division by zero")
		}
		leftdenom := left.(ValNumber).Val.Denom()
		rightdenom := right.(ValNumber).Val.Denom()
		if !leftdenom.IsInt64() || leftdenom.Int64() != 1 ||
			!rightdenom.IsInt64() || rightdenom.Int64() != 1 {
			return nil, NewRuntimeError(t, "Modulo only works on integers")
		}
		return ValNumber{
			Val: new(big.Rat).SetInt(
				new(big.Int).Mod(
					left.(ValNumber).Val.Num(),
					right.(ValNumber).Val.Num()))}, nil
	},
	opkey{"+", typename(ValString{}), typename(ValString{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValString{Val: left.(ValString).Val + right.(ValString).Val}, nil
	},
	opkey{"<", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValNumber).Val.Cmp(right.(ValNumber).Val) < 0}, nil
	},
	opkey{"<=", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValNumber).Val.Cmp(right.(ValNumber).Val) <= 0}, nil
	},
	opkey{">", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValNumber).Val.Cmp(right.(ValNumber).Val) > 0}, nil
	},
	opkey{">=", typename(ValNumber{}), typename(ValNumber{})}: func(
		t *ast.Token, left, right Value) (Value, error) {
		return ValBool{Val: left.(ValNumber).Val.Cmp(right.(ValNumber).Val) >= 0}, nil
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
