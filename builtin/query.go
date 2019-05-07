package builtin

import (
	"fmt"
	"strings"

	. "github.com/apmckinlay/gsuneido/runtime"
)

var _ = builtinRaw("Query1(@args)",
	func(t *Thread, as *ArgSpec, args ...Value) Value {
		return queryOne("Query1", t, false, true, as, args...)
	})

var _ = builtinRaw("QueryFirst(@args)",
	func(t *Thread, as *ArgSpec, args ...Value) Value {
		return queryOne("QueryFirst", t, false, false, as, args...)
	})

var _ = builtinRaw("QueryLast(@args)",
	func(t *Thread, as *ArgSpec, args ...Value) Value {
		return queryOne("QueryLast", t, true, false, as, args...)
	})

const noTran = 0

func queryOne(which string, t *Thread, prev bool, single bool,
	as *ArgSpec, args ...Value) Value {
	query := buildQuery(which, as, args)
	row, hdr := t.Dbms().Get(noTran, query, prev, single)
	fmt.Println(hdr)
	fmt.Println(row)
	return SuRecordFromRow(row, hdr)
}

func buildQuery(which string, as *ArgSpec, args []Value) string {
	iter := NewArgsIter(as, args)
	k, v := iter()
	if k != nil || v == nil {
		panic("usage: " + which + "(query, [field: value, ...])")
	}
	var sb strings.Builder
	sb.WriteString(IfStr(v))
	for {
		k, v := iter()
		if v == nil {
			break
		}
		if k == nil {
			panic("usage: " + which + "(query, [field: value, ...])")
		}
		field := IfStr(k)
		if field == "block" {
			continue
		}
		sb.WriteString("\nwhere ")
		sb.WriteString(field)
		sb.WriteString(" = ")
		sb.WriteString(v.String())
	}
	return sb.String()
}
