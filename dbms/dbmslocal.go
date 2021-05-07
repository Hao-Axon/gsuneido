// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package dbms

import (
	"fmt"
	"log"
	"strings"

	"github.com/apmckinlay/gsuneido/db19"
	"github.com/apmckinlay/gsuneido/db19/index/ixkey"
	"github.com/apmckinlay/gsuneido/db19/tools"
	qry "github.com/apmckinlay/gsuneido/dbms/query"
	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/str"
)

// DbmsLocal implements the Dbms interface using a local database
// i.e. standalone
type DbmsLocal struct {
	db        *db19.Database
	libraries []string //TODO concurrency
}

func NewDbmsLocal(db *db19.Database) IDbms {
	return &DbmsLocal{db: db}
}

// Dbms interface

var _ IDbms = (*DbmsLocal)(nil)

func (dbms DbmsLocal) Admin(req string) {
	qry.DoRequest(dbms.db, req)
}

func (DbmsLocal) Auth(string) bool {
	panic("Auth only allowed on clients")
}

func (dbms DbmsLocal) Check() string {
	if err := dbms.db.Check(); err != nil {
		return fmt.Sprint(err)
	}
	return ""
}

func (DbmsLocal) Connections() Value {
	return EmptyObject
}

func (DbmsLocal) Cursor(string) ICursor {
	panic("DbmsLocal Cursor not implemented")
}

func (DbmsLocal) Cursors() int {
	panic("DbmsLocal Cursors not implemented")
}

func (dbms DbmsLocal) Dump(table string) string {
	var err error
	if table == "" {
		_, err = tools.Dump(dbms.db, "database.su")
	} else {
		_, err = tools.DumpDbTable(dbms.db, table, table+".su")
	}
	if err != nil {
		return fmt.Sprint(err)
	}
	return ""
}

func (DbmsLocal) Exec(t *Thread, v Value) Value {
	fname := ToStr(ToContainer(v).ListGet(0))
	if i := strings.IndexByte(fname, '.'); i != -1 {
		ob := Global.GetName(t, fname[:i])
		m := fname[i+1:]
		return t.CallLookupEach1(ob, m, v)
	}
	fn := Global.GetName(t, fname)
	return t.CallEach1(fn, v)
}

func (DbmsLocal) Final() int {
	panic("DbmsLocal Final not implemented")
}

func (dbms DbmsLocal) Get(query string, dir Dir) (Row, *Header) {
	tran := dbms.db.NewReadTran()
	defer tran.Complete()
	return get(tran, query, dir)
}

func get(tran qry.QueryTran, query string, dir Dir) (Row, *Header) {
	q := qry.ParseQuery(query)
	qry.Setup(q, qry.ReadMode, tran)
	only := false
	if dir == Only {
		only = true
		dir = Next
	}
	row := q.Get(dir)
	if row == nil {
		return nil, nil
	}
	if only && q.Get(dir) != nil {
		panic("Query1 not unique: " + query)
	}
	return row, q.Header()
}

func (DbmsLocal) Info() Value {
	panic("DbmsLocal Info not implemented")
}

func (DbmsLocal) Kill(string) int {
	panic("DbmsLocal Kill not implemented")
}

func (DbmsLocal) Load(string) int {
	panic("DbmsLocal Load not implemented")
}

func (dbms DbmsLocal) LibGet(name string) (result []string) {
	defer func() {
		if e := recover(); e != nil {
			// debug.PrintStack()
			panic("error loading " + name + " " + fmt.Sprint(e))
		}
	}()

	// TODO
	rt := dbms.db.NewReadTran()
	ix := rt.GetIndex("stdlib", []string{"name", "group"})
	var rb ixkey.Encoder
	rb.Add(Pack(SuStr(name)))
	rb.Add(Pack(SuInt(-1))) // group
	key := rb.String()
	off := ix.Lookup(key)
	if off == 0 {
		if !strings.HasPrefix(name, "Rule_") {
			fmt.Println("LibGet", name, "NOT FOUND")
		}
		return nil
	}
	rec := rt.GetRecord(off)
	s := rec.GetStr(rt.ColToFld("stdlib", "text"))

	// fmt.Println("LOAD", name, "SUCCEEDED")
	return []string{"stdlib", string(s)}
}

func (DbmsLocal) Libraries() *SuObject {
	return &SuObject{}
}

func (DbmsLocal) Log(s string) {
	log.Println(s)
}

func (DbmsLocal) Nonce() string {
	panic("nonce only allowed on clients")
}

func (DbmsLocal) Run(string) Value {
	panic("DbmsLocal Run not implemented")
}

var sessionId string

func (DbmsLocal) SessionId(id string) string {
	if id != "" {
		sessionId = id
	}
	return sessionId
}

func (DbmsLocal) Size() int64 {
	panic("DbmsLocal Size not implemented")
}

func (DbmsLocal) Token() string {
	panic("DbmsLocal Token not implemented")
}

func (dbms DbmsLocal) Transaction(update bool) ITran {
	if update {
		return &UpdateTranLocal{dbms.db.NewUpdateTran()}
	}
	return &ReadTranLocal{dbms.db.NewReadTran()}
}

var prevTimestamp SuDate

func (DbmsLocal) Timestamp() SuDate {
	t := Now()
	if t.Equal(prevTimestamp) {
		t = t.Plus(0, 0, 0, 0, 0, 0, 1)
	}
	prevTimestamp = t
	return t
}

func (DbmsLocal) Transactions() *SuObject {
	panic("DbmsLocal Transactions not implemented")
}

func (dbms DbmsLocal) Unuse(lib string) bool {
	if lib == "stdlib" || !str.List(dbms.libraries).Has(lib) {
		return false
	}
	dbms.libraries = str.List(dbms.libraries).Without(lib)
	return true
}

func (dbms DbmsLocal) Use(lib string) bool {
	if str.List(dbms.libraries).Has(lib) {
		return false
	}
	//TODO check schema
	dbms.libraries = append(dbms.libraries, lib)
	return true
}

func (DbmsLocal) Close() {
}

// ReadTranLocal --------------------------------------------------------

type ReadTranLocal struct {
	*db19.ReadTran
}

func (t ReadTranLocal) Get(query string, dir Dir) (Row, *Header) {
	return get(t.ReadTran, query, dir)
}

func (t ReadTranLocal) Query(query string) IQuery {
	panic("ReadTranLocal Query not implemented") //TODO
}

func (t ReadTranLocal) Request(request string) int {
	panic("ReadTranLocal Request not implemented") //TODO
}

// UpdateTranLocal --------------------------------------------------------

type UpdateTranLocal struct {
	*db19.UpdateTran
}

func (t UpdateTranLocal) Get(query string, dir Dir) (Row, *Header) {
	return get(t.UpdateTran, query, dir)
}

func (t UpdateTranLocal) Query(query string) IQuery {
	panic("UpdateTranLocal Query not implemented") //TODO
}

func (t UpdateTranLocal) Request(request string) int {
	panic("UpdateTranLocal Request not implemented") //TODO
}
