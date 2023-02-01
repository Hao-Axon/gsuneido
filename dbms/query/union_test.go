// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/assert"
)

func TestUnion_MergeSwitchDir(t *testing.T) {
	db := heapDb()
	db.adm("create one (a) key(a)")
	db.adm("create two (a) key(a)")
	db.act("insert { a: 1 } into one")
	db.act("insert { a: 2 } into two")
	db.act("insert { a: 3 } into two")
	db.act("insert { a: 4 } into one")
	db.act("insert { a: 4 } into two")
	db.act("insert { a: 5 } into two")
	db.act("insert { a: 6 } into one")
	db.act("insert { a: 7 } into one")
	db.act("insert { a: 8 } into two")
	db.act("insert { a: 9 } into two")
	tran := db.NewReadTran()
	q := ParseQuery("one union two", tran, nil)
	q, _, _ = Setup(q, ReadMode, tran)
	fmt.Println(Format(q))
	assert.That(strings.Contains(q.String(), "MERGE"))

	get := func(dir Dir) Row {
		// fmt.Println(dir)
		row := q.Get(nil, dir)
		// fmt.Println("=>", row)
		return row
	}
	hdr := q.Header()
	a := func(row Row) int {
		if row == nil {
			return 0
		}
		return ToInt(row.GetVal(hdr, "a", nil, nil))
	}

	// random walk
	cur := 0 // rewound
	for i := 0; i < 1000; i++ {
		dir := Next
		if rand.Int()%2 == 0 {
			dir = Prev
			if cur--; cur < 0 {
				cur = 9
			}
		} else {
			cur = (cur + 1) % 10
		}
		row := get(dir)
		assert.This(a(row)).Is(cur)
		if row == nil {
			q.Rewind()
		}
	}
}
