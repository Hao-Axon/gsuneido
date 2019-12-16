// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package ast

import (
	tok "github.com/apmckinlay/gsuneido/lexer/tokens"
	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/dnum"
	"github.com/apmckinlay/gsuneido/util/regex"
)

// Folder implements constant folding for expressions.
// It is a "decorator" Factory that wraps another Factory (e.g. Builder)
// Doing the folding as the AST is built is implicitly bottom up
// without requiring an explicit tree traversal.
// It also means we only build the folded tree.
type Folder struct {
	Factory
}

func (f Folder) Unary(token tok.Token, expr Expr) Expr {
	c, ok := expr.(*Constant)
	if !ok || token == tok.Div {
		return f.Factory.Unary(token, expr)
	}
	val := c.Val
	switch token {
	case tok.Add:
		val = UnaryPlus(val)
	case tok.Sub:
		val = UnaryMinus(val)
	case tok.Not:
		val = Not(val)
	case tok.BitNot:
		val = BitNot(val)
	case tok.LParen:
		break
	default:
		panic("folder unexpected unary operator " + token.String())
	}
	return f.Constant(val)
}

func (f Folder) Binary(lhs Expr, token tok.Token, rhs Expr) Expr {
	c1, ok := lhs.(*Constant)
	if !ok {
		return f.Factory.Binary(lhs, token, rhs)
	}
	c2, ok := rhs.(*Constant)
	if !ok {
		return f.Factory.Binary(lhs, token, rhs)
	}
	val := c1.Val
	val2 := c2.Val
	switch token {
	case tok.Is:
		val = Is(val, val2)
	case tok.Isnt:
		val = Isnt(val, val2)
	case tok.Match:
		pat := regex.Compile(ToStr(val2))
		val = Match(val, pat)
	case tok.MatchNot:
		pat := regex.Compile(ToStr(val2))
		val = Not(Match(val, pat))
	case tok.Lt:
		val = Lt(val, val2)
	case tok.Lte:
		val = Lte(val, val2)
	case tok.Gt:
		val = Gt(val, val2)
	case tok.Gte:
		val = Gte(val, val2)
	case tok.Mod:
		val = Mod(val, val2)
	case tok.LShift:
		val = LeftShift(val, val2)
	case tok.RShift:
		val = RightShift(val, val2)
	default:
		panic("folder unexpected unary operator " + token.String())
	}
	return f.Constant(val)
}

func (f Folder) Trinary(cond Expr, e1 Expr, e2 Expr) Expr {
	c, ok := cond.(*Constant)
	if !ok {
		return f.Factory.Trinary(cond, e1, e2)
	}
	if c.Val == True {
		return e1
	}
	if c.Val == False {
		return e2
	}
	panic("?: requires boolean")
}

func (f Folder) In(e Expr, exprs []Expr) Expr {
	c, ok := e.(*Constant)
	if !ok {
		return f.Factory.In(e, exprs)
	}
	for _, e := range exprs {
		c2, ok := e.(*Constant)
		if !ok {
			return f.Factory.In(e, exprs)
		}
		if c.Val.Equal(c2.Val) {
			return f.Constant(True)
		}
	}
	return f.Constant(False)
}

var allones Value = SuDnum{Dnum: dnum.FromInt(0xffffffff)}

func (f Folder) Nary(token tok.Token, exprs []Expr) Expr {
	switch token {
	case tok.Add: // including Sub
		exprs = commutative(exprs, Add, nil)
	case tok.Mul: // including Div
		exprs = f.foldMul(exprs)
	case tok.BitOr:
		exprs = commutative(exprs, BitOr, allones)
	case tok.BitAnd:
		exprs = commutative(exprs, BitAnd, Zero)
	case tok.BitXor:
		exprs = commutative(exprs, BitXor, nil)
	case tok.Or:
		exprs = commutative(exprs, or, True)
	case tok.And:
		exprs = commutative(exprs, and, False)
	case tok.Cat:
		exprs = foldCat(exprs)
	default:
		// 	panic("folder unexpected n-ary operator " + token.String())
	}
	if len(exprs) == 1 {
		return exprs[0]
	}
	return f.Factory.Nary(token, exprs)
}

func or(x, y Value) Value {
	return SuBool(Bool(x) || Bool(y))
}

func and(x, y Value) Value {
	return SuBool(Bool(x) && Bool(y))
}

type bopfn func(Value, Value) Value

// commutative folds constants in a list of expressions
// fold is a short circuit value e.g. zero for multiply
func commutative(exprs []Expr, bop bopfn, fold Value) []Expr {
	var first *Constant
	dst := 0
	for _, e := range exprs {
		if c, ok := e.(*Constant); !ok {
			exprs[dst] = e
			dst++
		} else {
			if c.Val.Equal(fold) {
				exprs[0] = c
				return exprs[:1]
			}
			if first == nil {
				first = c
				exprs[dst] = e
				dst++
			} else {
				first.Val = bop(first.Val, c.Val)
			}
		}
	}
	return exprs[:dst]
}

func (f Folder) foldMul(exprs []Expr) []Expr {
	// extract and combine constants
	mul := One
	div := One
	dst := 0
	for _, e := range exprs {
		if ud := unaryDivConst(e); ud != nil {
			div = Mul(div, ud)
		} else if c, ok := e.(*Constant); ok {
			mul = Mul(mul, c.Val)
		} else {
			exprs[dst] = e
			dst++
		}
	}
	exprs = exprs[:dst]

	if !div.Equal(One) && (!mul.Equal(One) || len(exprs) == 0) {
		mul = Div(mul, div)
		div = One
	}
	if div.Equal(One) {
		if !mul.Equal(One) {
			exprs = append(exprs, f.Constant(mul))
		}
	} else {
		exprs = append(exprs, f.Unary(tok.Div, f.Constant(div)))
	}
	if len(exprs) == 1 && !unaryDivOrConstant(exprs[0]) {
		// force an operation to preserve conversion
		exprs = append(exprs, f.Constant(One))
	}
	return exprs
}

func unaryDivConst(e Expr) Value {
	if u, ok := e.(*Unary); ok {
		if c, ok := u.E.(*Constant); ok {
			return c.Val
		}
	}
	return nil
}

func unaryDivOrConstant(e Expr) bool {
	u, ok := e.(*Unary)
	if ok && u.Tok == tok.Div {
		return true
	}
	_, ok = e.(*Constant)
	return ok
}

// foldCat folds contiguous constants in a list of expressions
// cat is not commutative, so only combine contiguous constants
func foldCat(exprs []Expr) []Expr {
	var first *Constant
	dst := 0
	for _, e := range exprs {
		if c, ok := e.(*Constant); !ok {
			exprs[dst] = e
			dst++
			first = nil
		} else if first == nil {
			first = c
			exprs[dst] = e
			dst++
		} else {
			first.Val = SuStr(AsStr(first.Val) + AsStr(c.Val))
		}
	}
	return exprs[:dst]
}
