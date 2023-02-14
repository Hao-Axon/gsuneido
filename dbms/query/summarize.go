// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import (
	"fmt"
	"log"
	"sort"
	"strings"

	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/generic/hmap"
	"github.com/apmckinlay/gsuneido/util/generic/ord"
	"github.com/apmckinlay/gsuneido/util/generic/set"
	"github.com/apmckinlay/gsuneido/util/generic/slc"
	"github.com/apmckinlay/gsuneido/util/str"
)

type Summarize struct {
	Query1
	by []string
	// cols, ops, and ons are parallel
	cols     []string
	ops      []string
	ons      []string
	wholeRow bool
	summarizeApproach
	rewound bool
	srcHdr  *Header
	get     func(th *Thread, su *Summarize, dir Dir) Row
	t       QueryTran
	st      *SuTran
}

type summarizeApproach struct {
	strategy sumStrategy
	index    []string
	frac     float64
}

type sumStrategy int

const (
	// sumSeq reads in order of 'by', and groups consecutive, like projSeq
	sumSeq sumStrategy = iota + 1
	// sumMap uses a map to accumulate results. It is not incremental -
	// it must process all the source before producing any results.
	sumMap
	// sumIdx uses an index to get an overall min or max
	sumIdx
	// sumTbl optimizes <table> summarize count
	sumTbl
)

func NewSummarize(src Query, by, cols, ops, ons []string) *Summarize {
	if !set.Subset(src.Columns(), by) {
		panic("summarize: nonexistent columns: " +
			str.Join(", ", set.Difference(by, src.Columns())))
	}
	check(by)
	check(ons)
	for i := 0; i < len(cols); i++ {
		if cols[i] == "" {
			if ons[i] == "" {
				cols[i] = "count"
			} else {
				cols[i] = ops[i] + "_" + ons[i]
			}
		}
	}
	su := &Summarize{Query1: Query1{source: src},
		by: by, cols: cols, ops: ops, ons: ons}
	sort.Stable(su)
	// if single min or max, and on is a key, then we can give the whole row
	su.wholeRow = su.minmax1() && slc.ContainsFn(src.Keys(), ons, set.Equal[string])
	return su
}

// Len, Less, Swap implement sort.Interface
func (su *Summarize) Len() int {
	return len(su.cols)
}
func (su *Summarize) Less(i, j int) bool {
	return su.ons[i] < su.ons[j]
}
func (su *Summarize) Swap(i, j int) {
	su.ons[i], su.ons[j] = su.ons[j], su.ons[i]
	su.cols[i], su.cols[j] = su.cols[j], su.cols[i]
	su.ops[i], su.ops[j] = su.ops[j], su.ops[i]
}

func check(cols []string) {
	for _, c := range cols {
		if strings.HasSuffix(c, "_lower!") {
			panic("can't summarize _lower! fields")
		}
	}
}

func (su *Summarize) minmax1() bool {
	return len(su.by) == 0 &&
		len(su.ops) == 1 && (su.ops[0] == "min" || su.ops[0] == "max")
}

func (su *Summarize) SetTran(t QueryTran) {
	su.t = t
	su.st = MakeSuTran(su.t)
	su.source.SetTran(t)
}

func (su *Summarize) String() string {
	return parenQ2(su.source) + " " + su.stringOp()
}

func (su *Summarize) stringOp() string {
	s := "SUMMARIZE"
	switch su.strategy {
	case sumSeq:
		s += "-SEQ"
	case sumMap:
		s += "-MAP"
	case sumIdx:
		s += "-IDX"
	case sumTbl:
		s += "-TBL"
	}
	if su.wholeRow {
		s += "*"
	}
	if len(su.by) > 0 {
		s += " " + str.Join(", ", su.by) + ","
	}
	sep := " "
	for i := range su.cols {
		s += sep
		sep = ", "
		if su.cols[i] != "" {
			s += su.cols[i] + " = "
		}
		s += su.ops[i]
		if su.ops[i] != "count" {
			s += " " + su.ons[i]
		}
	}
	return s
}

func (su *Summarize) Columns() []string {
	if su.wholeRow {
		return set.Union(su.source.Columns(), su.cols)
	}
	return set.Union(su.by, su.cols)
}

func (su *Summarize) Keys() [][]string {
	if len(su.by) == 0 {
		// singleton
		return [][]string{{}} // intentionally {} not nil
	}
	return projectKeys(su.source.Keys(), su.by)
}

func (su *Summarize) Indexes() [][]string {
	if len(su.by) == 0 || containsKey(su.by, su.source.Keys()) {
		return su.source.Indexes()
	}
	var idxs [][]string
	for _, src := range su.source.Indexes() {
		if set.StartsWithSet(src, su.by) {
			idxs = append(idxs, src)
		}
	}
	return idxs
}

// containsKey returns true if a set of columns contain one of the keys
func containsKey(cols []string, keys [][]string) bool {
	for _, key := range keys {
		if set.Subset(cols, key) {
			return true
		}
	}
	return false
}

func (su *Summarize) Nrows() (int, int) {
	nr, pop := su.source.Nrows()
	if len(su.by) == 0 {
		nr = 1
	} else if !containsKey(su.by, su.source.Keys()) {
		nr /= 2 // ???
	}
	return nr, pop
}

func (su *Summarize) rowSize() int {
	return su.source.rowSize() + len(su.cols)*8
}

func (su *Summarize) Updateable() string {
	return "" // override Query1 source.Updateable
}

func (su *Summarize) SingleTable() bool {
	return false
}

func (*Summarize) Output(*Thread, Record) {
	panic("can't output to this query")
}

func (su *Summarize) Transform() Query {
	if p, ok := su.source.(*Project); ok && p.unique {
		// remove project-copy
		su = NewSummarize(p.source, su.by, su.cols, su.ops, su.ons)
	}
	src := su.source.Transform()
	if _, ok := src.(*Nothing); ok {
		return NewNothing(su.Columns())
	}
	if src != su.source {
		return NewSummarize(src, su.by, su.cols, su.ops, su.ons)
	}
	return su
}

func (su *Summarize) optimize(mode Mode, index []string, frac float64) (Cost, Cost, any) {
	if _, ok := su.source.(*Table); ok &&
		len(su.by) == 0 && len(su.ops) == 1 && su.ops[0] == "count" {
		Optimize(su.source, mode, nil, 0)
		return 0, 1, &summarizeApproach{strategy: sumTbl}
	}
	seqFixCost, seqVarCost, seqApp := su.seqCost(mode, index, frac)
	idxFixCost, idxVarCost, idxApp := su.idxCost(mode)
	mapFixCost, mapVarCost, mapApp := su.mapCost(mode, index, frac)
	// trace.Println("summarize seq", seqFixCost, "+", seqVarCost, "=", seqFixCost+seqVarCost)
	// trace.Println("summarize idx", idxFixCost, "+", idxVarCost, "=", idxFixCost+idxVarCost)
	// trace.Println("summarize map", mapFixCost, "+", mapVarCost, "=", mapFixCost+mapVarCost)
	return min3(seqFixCost, seqVarCost, seqApp, idxFixCost, idxVarCost, idxApp,
		mapFixCost, mapVarCost, mapApp)
}

func (su *Summarize) seqCost(mode Mode, index []string, frac float64) (Cost, Cost, any) {
	if len(su.by) == 0 {
		frac = ord.Min(1, frac)
	}
	approach := &summarizeApproach{strategy: sumSeq, frac: frac}
	if len(su.by) == 0 || containsKey(su.by, su.source.Keys()) {
		if len(su.by) != 0 {
			approach.index = index
		}
		fixcost, varcost := Optimize(su.source, mode, approach.index, frac)
		return fixcost, varcost, approach
	}
	best := bestGrouped(su.source, mode, index, frac, su.by)
	if best.index == nil {
		return impossible, impossible, nil
	}
	approach.index = best.index
	return best.fixcost, best.varcost, approach
}

func (su *Summarize) idxCost(mode Mode) (Cost, Cost, any) {
	if !su.minmax1() {
		return impossible, impossible, nil
	}
	nrows, _ := su.source.Nrows()
	frac := 1.
	if nrows > 0 {
		frac = 1 / float64(nrows)
	}
	fixcost, varcost := Optimize(su.source, mode, su.ons, frac)
	return fixcost, varcost,
		&summarizeApproach{strategy: sumIdx, index: su.ons, frac: frac}
}

func (su *Summarize) mapCost(mode Mode, index []string, _ float64) (Cost, Cost, any) {
	//FIXME technically, map should only be allowed in ReadMode
	nrows, _ := su.Nrows()
	if index != nil || nrows > mapLimit-mapLimit/3 {
		return impossible, impossible, nil
	}
	fixcost, varcost := Optimize(su.source, mode, nil, 1)
	fixcost += nrows * 20 // ???
	return fixcost + varcost, 0, &summarizeApproach{strategy: sumMap, frac: 1}
}

func (su *Summarize) setApproach(_ []string, frac float64, approach any, tran QueryTran) {
	su.summarizeApproach = *approach.(*summarizeApproach)
	switch su.strategy {
	case sumTbl:
		su.get = getTbl
	case sumIdx:
		su.get = getIdx
	case sumMap:
		t := sumMapT{}
		su.get = t.getMap
	case sumSeq:
		t := sumSeqT{}
		su.get = t.getSeq
	}
	su.source = SetApproach(su.source, su.index, su.frac, tran)
	su.rewound = true
	su.srcHdr = su.source.Header()
}

// execution --------------------------------------------------------

func (su *Summarize) Header() *Header {
	if su.wholeRow {
		flds := su.source.Header().Fields
		flds = slc.With(flds, su.cols)
		return NewHeader(flds, su.Columns())
	}
	flds := append(su.by, su.cols...)
	return NewHeader([][]string{flds}, flds)
}

func (su *Summarize) Rewind() {
	su.source.Rewind()
	su.rewound = true
}

func (su *Summarize) Get(th *Thread, dir Dir) Row {
	defer func() { su.rewound = false }()
	return su.get(th, su, dir)
}

func getTbl(_ *Thread, su *Summarize, _ Dir) Row {
	if !su.rewound {
		return nil
	}
	tbl := su.source.(*Table)
	var rb RecordBuilder
	nr, _ := tbl.Nrows()
	rb.Add(IntVal(nr))
	return Row{DbRec{Record: rb.Build()}}
}

func getIdx(th *Thread, su *Summarize, _ Dir) Row {
	if !su.rewound {
		return nil
	}
	dir := Prev // max
	if str.EqualCI(su.ops[0], "min") {
		dir = Next
	}
	row := su.source.Get(th, dir)
	if row == nil {
		return nil
	}
	var rb RecordBuilder
	rb.AddRaw(row.GetRawVal(su.srcHdr, su.ons[0], th, su.st))
	rec := rb.Build()
	if su.wholeRow {
		return append(row, DbRec{Record: rec})
	}
	return Row{DbRec{Record: rec}}
}

//-------------------------------------------------------------------

type sumMapT struct {
	mapList []mapPair
	mapPos  int
}

type mapPair struct {
	row Row
	ops []sumOp
}

func (t *sumMapT) getMap(th *Thread, su *Summarize, dir Dir) Row {
	if su.rewound {
		assert.That(!su.wholeRow)
		t.mapList = su.buildMap(th)
		if dir == Next {
			t.mapPos = -1
		} else { // Prev
			t.mapPos = len(t.mapList)
		}
	}
	if dir == Next {
		t.mapPos++
	} else { // Prev
		t.mapPos--
	}
	if t.mapPos < 0 || len(t.mapList) <= t.mapPos {
		return nil
	}
	row := t.mapList[t.mapPos].row
	var rb RecordBuilder
	for _, col := range su.by {
		rb.AddRaw(row.GetRawVal(su.srcHdr, col, th, su.st))
	}
	ops := t.mapList[t.mapPos].ops
	for i := range ops {
		val, _ := ops[i].result()
		rb.Add(val.(Packable))
	}
	return Row{DbRec{Record: rb.Build()}}
}

func (su *Summarize) buildMap(th *Thread) []mapPair {
	hdr := su.source.Header()
	hfn := func(k rowHash) uint32 { return k.hash }
	eqfn := func(x, y rowHash) bool {
		return x.hash == y.hash &&
			equalCols(x.row, y.row, hdr, su.by, th, su.st)
	}
	sumMap := hmap.NewHmapFuncs[rowHash, []sumOp](hfn, eqfn)
	for {
		row := su.source.Get(th, Next)
		if row == nil {
			break
		}
		rh := rowHash{hash: hashCols(row, hdr, su.by, th, su.st), row: row}
		sums := sumMap.Get(rh)
		if sums == nil {
			sums = su.newSums()
			sumMap.Put(rh, sums)
			if sumMap.Size() > mapLimit {
				log.Panicf("summarize-map too large (> %d)", mapLimit)
			}
		}
		su.addToSums(sums, row, th, su.st)
	}
	i := 0
	list := make([]mapPair, sumMap.Size())
	iter := sumMap.Iter()
	for rh, ops := iter(); rh.row != nil; rh, ops = iter() {
		list[i] = mapPair{row: rh.row, ops: ops}
		i++
	}
	return list
}

//-------------------------------------------------------------------

type sumSeqT struct {
	curDir  Dir
	curRow  Row
	nextRow Row
	sums    []sumOp
}

func (t *sumSeqT) getSeq(th *Thread, su *Summarize, dir Dir) Row {
	if su.rewound {
		t.sums = su.newSums()
		t.curDir = dir
		t.curRow = nil
		t.nextRow = su.source.Get(th, dir)
	}

	// if direction changes, have to skip over previous result
	if dir != t.curDir {
		if t.nextRow == nil {
			su.source.Rewind()
		}
		for {
			t.nextRow = su.source.Get(th, dir)
			if t.nextRow == nil || !su.sameBy(th, su.st, t.curRow, t.nextRow) {
				break
			}
		}
		t.curDir = dir
	}

	if t.nextRow == nil {
		return nil
	}
	t.curRow = t.nextRow
	for i := range t.sums {
		t.sums[i].reset()
	}
	for {
		su.addToSums(t.sums, t.nextRow, th, su.st)
		t.nextRow = su.source.Get(th, dir)
		if t.nextRow == nil || !su.sameBy(th, su.st, t.curRow, t.nextRow) {
			break
		}
	}
	// output after each group
	return su.seqRow(th, t.curRow, t.sums)
}

func (su *Summarize) addToSums(sums []sumOp, row Row, th *Thread, st *SuTran) {
	for i := 0; i < len(su.ons); {
		raw := "*uninit*"
		var val Value
		for col := su.ons[i]; i < len(su.ons) && su.ons[i] == col; i++ {
			switch su.ops[i] {
			case "count":
				sums[i].add("", nil, row)
			case "list", "min", "max":
				if raw == "*uninit*" {
					raw = row.GetRawVal(su.srcHdr, col, th, st)
				}
				sums[i].add(raw, nil, row)
			default: // total, average
				if val == nil {
					val = row.GetVal(su.srcHdr, col, th, st)
				}
				sums[i].add("", val, row)
			}
		}
	}
}

func (su *Summarize) sumRow(row Row) Row {
	if su.wholeRow {
		return row
	}
	return nil
}

func (su *Summarize) sameBy(th *Thread, st *SuTran, row1, row2 Row) bool {
	for _, f := range su.by {
		if row1.GetRawVal(su.srcHdr, f, th, st) !=
			row2.GetRawVal(su.srcHdr, f, th, st) {
			return false
		}
	}
	return true
}

func (su *Summarize) seqRow(th *Thread, curRow Row, sums []sumOp) Row {
	var rb RecordBuilder
	if !su.wholeRow {
		for _, fld := range su.by {
			rb.AddRaw(curRow.GetRawVal(su.srcHdr, fld, th, su.st))
		}
	}
	for _, sum := range sums {
		val, _ := sum.result()
		rb.Add(val.(Packable))
	}
	row := Row{DbRec{Record: rb.Build()}}
	if su.wholeRow {
		_, wholeRow := sums[0].result()
		row = append(wholeRow, row[0])
	}
	return row
}

func (su *Summarize) Select(cols, vals []string) {
	su.source.Select(cols, vals)
	su.rewound = true
}

// operations -------------------------------------------------------

func (su *Summarize) newSums() []sumOp {
	sums := make([]sumOp, len(su.ops))
	for i, op := range su.ops {
		sums[i] = newSumOp(op)
	}
	return sums
}

func newSumOp(op string) sumOp {
	switch op {
	case "count":
		return &sumCount{}
	case "total":
		return &sumTotal{total: Zero}
	case "average":
		return &sumAverage{total: Zero}
	case "min":
		return &sumMin{}
	case "max":
		return &sumMax{}
	case "list":
		return sumList(make(map[string]struct{}))
	}
	panic(assert.ShouldNotReachHere())
}

type sumOp interface {
	add(raw string, val Value, row Row)
	result() (Value, Row)
	reset()
}

type sumCount struct {
	count int
}

func (sum *sumCount) add(_ string, _ Value, _ Row) {
	sum.count++
}
func (sum *sumCount) result() (Value, Row) {
	return IntVal(sum.count), nil
}
func (sum *sumCount) reset() {
	sum.count = 0
}

type sumTotal struct {
	total Value
}

func (sum *sumTotal) add(_ string, val Value, _ Row) {
	defer func() { recover() }() // ignore panics
	sum.total = OpAdd(sum.total, val)
}
func (sum *sumTotal) result() (Value, Row) {
	return sum.total, nil
}
func (sum *sumTotal) reset() {
	sum.total = Zero
}

type sumAverage struct {
	count int
	total Value
}

func (sum *sumAverage) add(_ string, val Value, _ Row) {
	sum.count++
	defer func() { recover() }()
	sum.total = OpAdd(sum.total, val)
}
func (sum *sumAverage) result() (Value, Row) {
	return OpDiv(sum.total, IntVal(sum.count)), nil
}
func (sum *sumAverage) reset() {
	sum.count = 0
	sum.total = Zero
}

type sumMin struct {
	raw string
	row Row
}

func (sum *sumMin) add(raw string, _ Value, row Row) {
	if sum.row == nil || raw < sum.raw {
		sum.raw, sum.row = raw, row
	}
}
func (sum *sumMin) result() (Value, Row) {
	return Unpack(sum.raw), sum.row
}
func (sum *sumMin) reset() {
	sum.raw = ""
	sum.row = nil
}

type sumMax struct {
	raw string
	row Row
}

func (sum *sumMax) add(raw string, _ Value, row Row) {
	if sum.row == nil || raw > sum.raw {
		sum.raw, sum.row = raw, row
	}
}
func (sum *sumMax) result() (Value, Row) {
	return Unpack(sum.raw), sum.row
}
func (sum *sumMax) reset() {
	sum.raw = ""
	sum.row = nil
}

type sumList map[string]struct{}

const sumListLimit = 16384 // ???

func (sum sumList) add(raw string, _ Value, _ Row) {
	sum[raw] = struct{}{}
	if len(sum) > sumListLimit {
		panic(fmt.Sprintf("summarize list too large (> %d)", sumListLimit))
	}
}
func (sum sumList) result() (Value, Row) {
	list := make([]Value, len(sum))
	i := 0
	for raw := range sum {
		list[i] = Unpack(raw)
		i++
	}
	if len(list) <= 3 { // for tests
		sort.Slice(list,
			func(i, j int) bool { return list[i].Compare(list[j]) < 0 })
	}
	return NewSuObject(list), nil
}
func (sum sumList) reset() {
	for k := range sum {
		delete(sum, k)
	}
}
