// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import (
	"github.com/apmckinlay/gsuneido/compile"
	"github.com/apmckinlay/gsuneido/compile/ast"
	"github.com/apmckinlay/gsuneido/compile/lexer"
	tok "github.com/apmckinlay/gsuneido/compile/tokens"
)

type qparser struct {
	compile.Parser
}

func NewQueryParser(src string) *qparser {
	lxr := lexer.NewQueryLexer(src)
	p := &qparser{compile.Parser{ParserBase: compile.ParserBase{Lxr: lxr}}}
	p.Init()
	p.Next()
	return p
}

func ParseQuery(src string) Query {
	p := NewQueryParser(src)
	result := p.query()
	if p.Token != tok.Eof {
		p.Error("did not parse all input")
	}
	return result
}

func (p *qparser) query() Query {
	switch {
	case p.MatchIf(tok.Insert):
		return nil //p.insert() //TODO
	case p.MatchIf(tok.Update):
		return nil //p.update() //TODO
	case p.MatchIf(tok.Delete):
		return nil //p.delete() //TODO
	default:
		return p.sort()
	}
}

func (p *qparser) sort() Query {
	q := p.baseQuery()
	if p.MatchIf(tok.Sort) {
		reverse := p.MatchIf(tok.Reverse)
		q = &Sort{Query1: Query1{source: q},
			reverse: reverse, columns: p.commaList()}
	}
	return q
}

func (p *qparser) baseQuery() Query {
	q := p.source()
	for p.operation(&q) {
	}
	return q
}

func (p *qparser) source() Query {
	if p.MatchIf(tok.LParen) {
		q := p.baseQuery()
		p.Match(tok.RParen)
		return q
	}
	return p.table()
}

func (p *qparser) table() Query {
	return &Table{name: p.MatchIdent()}
}

func (p *qparser) operation(pq *Query) bool {
	switch {
	case p.MatchIf(tok.Extend):
		*pq = p.extend(*pq)
	case p.MatchIf(tok.Intersect):
		*pq = p.intersect(*pq)
	case p.MatchIf(tok.Join):
		*pq = p.join(*pq)
	case p.MatchIf(tok.Leftjoin):
		*pq = p.leftjoin(*pq)
	case p.MatchIf(tok.Minus):
		*pq = p.minus(*pq)
	case p.MatchIf(tok.Project):
		*pq = p.project(*pq)
	case p.MatchIf(tok.Remove):
		*pq = p.remove(*pq)
	case p.MatchIf(tok.Rename):
		*pq = p.rename(*pq)
	case p.MatchIf(tok.Summarize):
		*pq = p.summarize(*pq)
	case p.MatchIf(tok.Times):
		*pq = p.times(*pq)
	case p.MatchIf(tok.Union):
		*pq = p.union(*pq)
	case p.MatchIf(tok.Where):
		*pq = p.where(*pq)
	default:
		return false
	}
	return true
}

func (p *qparser) extend(q Query) Query {
	cols := make([]string, 0, 4)
	exprs := make([]ast.Expr, 0, 4)
	for {
		cols = append(cols, p.MatchIdent())
		var expr ast.Expr
		if p.MatchIf(tok.Eq) {
			expr = p.Expression()
		}
		exprs = append(exprs, expr)
		if !p.MatchIf(tok.Comma) {
			break
		}
	}
	return &Extend{Query1: Query1{source: q}, cols: cols, exprs: exprs}
}

func (p *qparser) intersect(q Query) Query {
	return &Intersect{Compatible: Compatible{
		Query2: Query2{Query1: Query1{source: q}, source2: p.source()}}}
}

func (p *qparser) join(q Query) Query {
	by := p.joinBy()
	return &Join{Query2: Query2{Query1: Query1{source: q},
		source2: p.source()}, by: by}
}

func (p *qparser) leftjoin(q Query) Query {
	by := p.joinBy()
	return &LeftJoin{Join: Join{Query2: Query2{Query1: Query1{source: q},
		source2: p.source()}, by: by}}
}

func (p *qparser) joinBy() []string {
	if p.MatchIf(tok.By) {
		by := p.parenList()
		if len(by) == 0 {
			p.Error("invalid empty join by")
		}
		return by
	}
	return nil
}

func (p *qparser) minus(q Query) Query {
	return &Minus{Compatible: Compatible{
		Query2: Query2{Query1: Query1{source: q}, source2: p.source()}}}
}

func (p *qparser) project(q Query) Query {
	return &Project{Query1: Query1{source: q}, columns: p.commaList()}
}

func (p *qparser) remove(q Query) Query {
	return &Remove{Query1: Query1{source: q}, columns: p.commaList()}
}

func (p *qparser) rename(q Query) Query {
	var from, to []string
	for {
		from = append(from, p.MatchIdent())
		p.Match(tok.To)
		to = append(to, p.MatchIdent())
		if !p.MatchIf(tok.Comma) {
			break
		}
	}
	return &Rename{Query1: Query1{source: q}, from: from, to: to}
}

func (p *qparser) summarize(q Query) Query {
	su := &Summarize{Query1: Query1{source: q}}
	su.by = p.sumBy()
	p.sumOps(su)
	return su
}

func (p *qparser) sumBy() []string {
	var by []string
	for p.Token == tok.Identifier &&
		!isSumOp(p.Token) &&
		p.Lxr.Ahead(1).Token != tok.Eq {
		by = append(by, p.MatchIdent())
		p.Match(tok.Comma)
	}
	return by
}

func (p *qparser) sumOps(su *Summarize) {
	for {
		var col, op, on string
		if p.Lxr.Ahead(1).Token == tok.Eq {
			col = p.MatchIdent()
			p.Match(tok.Eq)
		}
		if !isSumOp(p.Token) {
			p.Error("expected count, total, min, max, or list")
		}
		op = p.MatchIdent()
		if op != "count" {
			on = p.MatchIdent()
		}
		su.cols = append(su.cols, col)
		su.ops = append(su.ops, op)
		su.ons = append(su.ons, on)
		if !p.MatchIf(tok.Comma) {
			break
		}
	}
}

func isSumOp(t tok.Token) bool {
	return tok.SummarizeStart < t && t < tok.SummarizeEnd
}

func (p *qparser) times(q Query) Query {
	return &Times{Query2: Query2{Query1: Query1{source: q},
		source2: p.source()}}
}

func (p *qparser) union(q Query) Query {
	return &Union{Compatible: Compatible{
		Query2: Query2{Query1: Query1{source: q}, source2: p.source()}}}
}

func (p *qparser) where(q Query) Query {
	expr := p.Expression()
	if nary, ok := expr.(*ast.Nary); !ok || nary.Tok != tok.And {
		expr = &ast.Nary{Tok: tok.And, Exprs: []ast.Expr{expr}}
	}
	return &Where{Query1: Query1{source: q}, expr: expr.(*ast.Nary)}
}

func (p *qparser) parenList() []string {
	p.Match(tok.LParen)
	if p.MatchIf(tok.RParen) {
		return nil
	}
	list := p.commaList()
	p.Match(tok.RParen)
	return list
}

func (p *qparser) commaList() []string {
	list := make([]string, 0, 4)
	for {
		list = append(list, p.MatchIdent())
		if !p.MatchIf(tok.Comma) {
			break
		}
	}
	return list
}