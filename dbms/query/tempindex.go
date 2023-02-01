// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import (
	"fmt"

	"github.com/apmckinlay/gsuneido/db19/index/ixkey"
	"github.com/apmckinlay/gsuneido/db19/stor"
	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/sortlist"
	"github.com/apmckinlay/gsuneido/util/str"
	"golang.org/x/exp/slices"
)

// TempIndex is inserted by SetApproach as required.
// It builds a sortlist of record addresses.
// (More complex for multiple sources and derived.)
// Keys are not constructed for the index or Lookup/Select
// so there are no size limits.
type TempIndex struct {
	Query1
	order   []string
	tran    QueryTran
	st      *SuTran
	th      *Thread
	hdr     *Header
	iter    rowIter
	rewound bool
	selOrg  []string
	selEnd  []string
}

var selMin = []string{ixkey.Min}
var selMax = []string{ixkey.Max}

func NewTempIndex(src Query, order []string, tran QueryTran) *TempIndex {
	order = withoutFixed(order, src.Fixed())
	return &TempIndex{Query1: Query1{source: src}, order: order, tran: tran,
		selOrg: selMin, selEnd: selMax}
}

func (ti *TempIndex) String() string {
	return parenQ2(ti.source) + " " + ti.stringOp()
}

func (ti *TempIndex) stringOp() string {
	return "TEMPINDEX" + str.Join("(,)", ti.order)
}

func (ti *TempIndex) Transform() Query {
	return ti
}

func (ti *TempIndex) lookupCost() Cost {
	panic(assert.ShouldNotReachHere())
}

func (q1 *Query1) setApproach([]string, float64, any, QueryTran) {
	assert.ShouldNotReachHere()
}

// execution --------------------------------------------------------

func (ti *TempIndex) Rewind() {
	if ti.iter != nil {
		ti.iter.Rewind()
	}
	ti.rewound = true
}

func (ti *TempIndex) Select(cols, vals []string) {
	// similar to Where Select
	ti.Rewind()
	if cols == nil && vals == nil { // clear select
		ti.selOrg, ti.selEnd = selMin, selMax
		return
	}
	fixed := ti.source.Fixed()
	satisfied, conflict := selectFixed(cols, vals, fixed)
	if conflict {
		ti.selOrg, ti.selEnd = selMax, selMin
		return
	}
	if satisfied {
		ti.selOrg, ti.selEnd = selMin, selMax
		return
	}
	ti.selOrg = ti.makeKey(cols, vals, false)
	ti.selEnd = append(ti.selOrg, ixkey.Max)
}

func (ti *TempIndex) Lookup(th *Thread, cols, vals []string) Row {
	ti.th = th
	defer func() { ti.th = nil }()
	if ti.iter == nil {
		ti.iter = ti.makeIndex()
	}
	key := ti.makeKey(cols, vals, true)
	row := ti.iter.Seek(key)
	if row == nil || !ti.matches(row, key) {
		return nil
	}
	return row
}

func (ti *TempIndex) makeKey(cols, vals []string, full bool) []string {
	key := make([]string, 0, len(ti.order))
	for _, col := range ti.order {
		j := slices.Index(cols, col)
		if j == -1 {
			if full {
				fmt.Println("TempIndex makeKey order:", ti.order, "cols:", cols)
				// fmt.Println(Format(ti))
				panic("TempIndex makeKey not full")
			}
			break
		}
		key = append(key, vals[j])
	}
	return key
}

func (ti *TempIndex) matches(row Row, key []string) bool {
	for i, col := range ti.order {
		x := row.GetRawVal(ti.hdr, col, nil, ti.st)
		y := key[i]
		if x != y {
			return false
		}
	}
	return true
}

func (ti *TempIndex) Get(th *Thread, dir Dir) Row {
	ti.th = th
	defer func() { ti.th = nil }()
	if ti.iter == nil {
		ti.iter = ti.makeIndex()
		ti.rewound = true
	}
	if ti.conflict() {
		return nil
	}
	var row Row
	if ti.rewound {
		if dir == Next {
			row = ti.iter.Seek(ti.selOrg)
		} else { // Prev
			row = ti.iter.Seek(ti.selEnd)
			if row == nil {
				ti.iter.Rewind()
			}
			row = ti.iter.Get(dir)
		}
		ti.rewound = false
	} else {
		row = ti.iter.Get(dir)
	}
	if row == nil || !ti.selected(row) {
		return nil
	}
	return row
}

type rowIter interface {
	Get(Dir) Row
	Rewind()
	Seek(key []string) Row
}

func (ti *TempIndex) makeIndex() rowIter {
	ti.st = MakeSuTran(ti.tran)
	// need to copy header to avoid data race from concurrent sortlist
	ti.hdr = ti.source.Header().Dup()
	if ti.source.SingleTable() {
		return ti.single()
	}
	return ti.multi()
}

func (ti *TempIndex) selected(row Row) bool {
	if ti.satisfied() {
		return true
	}
	for i, sel := range ti.selOrg {
		col := ti.order[i]
		x := row.GetRawVal(ti.hdr, col, ti.th, ti.st)
		if x != sel {
			return false
		}
	}
	return true
}

func (ti *TempIndex) satisfied() bool {
	return len(ti.selOrg) == 1 && ti.selOrg[0] == ixkey.Min &&
		len(ti.selEnd) == 1 && ti.selEnd[0] == ixkey.Max
}

func (ti *TempIndex) conflict() bool {
	return len(ti.selOrg) == 1 && ti.selOrg[0] == ixkey.Max &&
		len(ti.selEnd) == 1 && ti.selEnd[0] == ixkey.Min
}

//-------------------------------------------------------------------
// single is for ti.source.SingleTable.
// Since there is only one table,
// we can store the record offset directly in the sortlist.

type singleIter struct {
	tran QueryTran
	iter *sortlist.Iter
}

func (ti *TempIndex) single() rowIter {
	// sortlist uses a goroutine
	// so UIThread must be false
	// so interp doesn't call Interrupt
	// leading to "illegal UI call from background thread"
	defer func(prev bool) {
		ti.th.UIThread = prev
	}(ti.th.UIThread)
	ti.th.UIThread = false
	var th2 Thread // separate thread because sortlist runs in the background
	b := sortlist.NewSorting(func(x, y uint64) bool {
		return ti.less(&th2, ti.singleGet(x), ti.singleGet(y))
	})
	for {
		row := ti.source.Get(ti.th, Next)
		if row == nil {
			break
		}
		b.Add(row[0].Off)
	}
	// lt must be consistent with singleLess
	lt := func(off uint64, key []string) bool {
		return ti.less2(ti.th, ti.singleGet(off), key)
	}
	return &singleIter{tran: ti.tran, iter: b.Finish().Iter(lt)}
}

func (ti *TempIndex) singleGet(off uint64) Row {
	rec := ti.tran.GetRecord(off)
	return Row{{Record: rec}}
}

// less is used by singleLess and multiLess
func (ti *TempIndex) less(th *Thread, xrow, yrow Row) bool {
	for _, col := range ti.order {
		x := xrow.GetRawVal(ti.hdr, col, th, ti.st)
		y := yrow.GetRawVal(ti.hdr, col, th, ti.st)
		if x != y {
			return x < y
		}
	}
	return false
}

func (ti *TempIndex) less2(th *Thread, row Row, key []string) bool {
	for i, col := range ti.order {
		if i >= len(key) {
			return false
		}
		x := row.GetRawVal(ti.hdr, col, th, ti.st)
		y := key[i]
		if x != y {
			return x < y
		}
	}
	return false
}

func (it *singleIter) Get(dir Dir) Row {
	if dir == Next {
		it.iter.Next()
	} else {
		it.iter.Prev()
	}
	if it.iter.Eof() {
		return nil
	}
	return it.get()
}

func (it *singleIter) get() Row {
	if it.iter.Eof() {
		return nil
	}
	off := it.iter.Cur()
	dbrec := DbRec{Record: it.tran.GetRecord(off), Off: off}
	return Row{dbrec}
}

func (it *singleIter) Seek(key []string) Row {
	it.iter.Seek(key)
	if it.iter.Eof() {
		return nil
	}
	return it.get()
}

func (it *singleIter) Rewind() {
	it.iter.Rewind()
}

//-------------------------------------------------------------------
// multi is when we have multiple records (not ti.source.SingleTable).
// So we store the record offsets and/or the in memory records in a separate heap
// that the sortlist points to.

type multiIter struct {
	ti    *TempIndex
	nrecs int
	heap  *stor.Stor
	iter  *sortlist.Iter
}

const heapChunkSize = 16 * 1024
const derivedMaxSize = 8 * 1024

func (ti *TempIndex) multi() rowIter {
	// sortlist uses a goroutine
	// so UIThread must be false
	// so interp doesn't call Interrupt
	// leading to "illegal UI call from background thread"
	defer func(prev bool) {
		ti.th.UIThread = prev
	}(ti.th.UIThread)
	ti.th.UIThread = false
	it := multiIter{ti: ti, nrecs: len(ti.hdr.Fields),
		heap: stor.HeapStor(heapChunkSize)}
	it.heap.Alloc(1) // avoid offset 0
	var th2 Thread   // separate thread because sortlist runs in the background
	b := sortlist.NewSorting(func(x, y uint64) bool {
		xrow := make(Row, it.nrecs)
		yrow := make(Row, it.nrecs)
		xbuf := it.heap.Data(x)
		ybuf := it.heap.Data(y)
		for i := 0; i < it.nrecs; i++ {
			xbuf, xrow[i].Record, _ = it.getRec(xbuf)
			ybuf, yrow[i].Record, _ = it.getRec(ybuf)
		}
		return it.ti.less(&th2, xrow, yrow)
	})
	for {
		row := ti.source.Get(ti.th, Next)
		if row == nil {
			break
		}
		assert.That(len(row) == it.nrecs)
		n := it.nrecs * stor.SmallOffsetLen
		for _, dbrec := range row {
			if dbrec.Off == 0 { // derived record e.g. from extend or summarize
				n += len(dbrec.Record)
			}
		}
		if n > derivedMaxSize {
			panic(fmt.Sprintf("temp index: derived too large (%d > %d)", n,
				derivedMaxSize))
		}
		off, buf := it.heap.Alloc(n)
		for _, dbrec := range row {
			if dbrec.Off > 0 {
				// database records are stored as their offset
				stor.WriteSmallOffset(buf, dbrec.Off)
				buf = buf[stor.SmallOffsetLen:]
			} else {
				// derived records are stored as their size|multiMask
				// followed by their actual contents
				size := len(dbrec.Record)
				stor.WriteSmallOffset(buf, multiMask|uint64(size))
				buf = buf[stor.SmallOffsetLen:]
				copy(buf, dbrec.Record)
				buf = buf[size:]
			}
		}
		b.Add(off)
	}
	// lt must be consistent with multiLess
	lt := func(off uint64, key []string) bool {
		return ti.less2(ti.th, it.get(off), key)
	}
	it.iter = b.Finish().Iter(lt)
	return &it
}

const multiMask = 0xffff000000

func (it *multiIter) Seek(key []string) Row {
	it.iter.Seek(key)
	if it.iter.Eof() {
		return nil
	}
	return it.get(it.iter.Cur())
}

func (it *multiIter) Rewind() {
	it.iter.Rewind()
}

func (it *multiIter) Get(dir Dir) Row {
	if dir == Next {
		it.iter.Next()
	} else {
		it.iter.Prev()
	}
	if it.iter.Eof() {
		return nil
	}
	return it.get(it.iter.Cur())
}

func (it *multiIter) get(off uint64) Row {
	row := make([]DbRec, it.nrecs)
	buf := it.heap.Data(off)
	var rec Record
	for i := 0; i < it.nrecs; i++ {
		buf, rec, off = it.getRec(buf)
		row[i] = DbRec{Record: rec, Off: off}
	}
	return row
}

func (it *multiIter) getRec(buf []byte) ([]byte, Record, uint64) {
	off := stor.ReadSmallOffset(buf)
	buf = buf[stor.SmallOffsetLen:]
	var rec Record
	if off < multiMask {
		rec = it.ti.tran.GetRecord(off)
	} else { // derived
		size := off &^ multiMask
		rec = Record(buf[:size])
		buf = buf[size:]
		off = 0
	}
	return buf, rec, off
}
