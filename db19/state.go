// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package db19

import (
	"encoding/binary"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/apmckinlay/gsuneido/db19/meta"
	"github.com/apmckinlay/gsuneido/db19/stor"
	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/cksum"
)

type DbState struct {
	store *stor.Stor
	meta  *meta.Meta
}

type stateHolder struct {
	state unsafe.Pointer // *DbState
	mutex sync.Mutex
}

func (sh *stateHolder) get() *DbState {
	return (*DbState)(atomic.LoadPointer(&sh.state))
}

func (sh *stateHolder) set(newState *DbState) {
	atomic.StorePointer(&sh.state, unsafe.Pointer(newState))
}

// GetState returns a snapshot of the state as of a point in time.
// This state must be treated as read-only and must not be modified.
// To modify the state use UpdateState.
func (db *Database) GetState() *DbState {
	return db.state.get()
}

func (sh *stateHolder) updateState(fn func(*DbState)) *DbState {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	newState := *sh.get() // shallow copy
	fn(&newState)
	sh.set(&newState)
	return &newState
}

// UpdateState applies the given update function to a copy of theState
// and sets theState to the result.
// Guarded by stateMutex so only one thread can execute at a time.
// Note: the state passed to the update function is a *shallow* copy,
// it is up to the function to make copies of any nested containers.
func (db *Database) UpdateState(fn func(*DbState)) *DbState {
	return db.state.updateState(fn)
}

//-------------------------------------------------------------------

// Merge updates the base fbtree's with the overlay mbtree
// for the given transaction number (the oldest/first).
// It is called by concur.go merger.
func (db *Database) Merge(tranNum int) {
	state := db.GetState()
	updates := state.meta.Merge(tranNum) // outside UpdateState
	db.UpdateState(func(state *DbState) {
		meta := *state.meta // copy
		meta.ApplyMerge(updates)
		state.meta = &meta
	})
}

//-------------------------------------------------------------------

// Persist writes index changes (and a new state) to the database file.
// It is called by concur.go persister.
func (db *Database) Persist(flatten bool) uint64 {
	state := db.GetState()
	updates := state.meta.Persist(flatten) // outside UpdateState
	state = db.UpdateState(func(state *DbState) {
		meta := *state.meta // copy
		meta.ApplyPersist(updates)
		state.meta = &meta
	})
	return state.Write()
}

const magic1 = "\x01\x23\x45\x67\x89\xab\xcd\xef"
const magic2 = "\xfe\xdc\xba\x98\x76\x54\x32\x10"
const dateSize = 8
const stateLen = len(magic1) + dateSize + meta.Noffsets*stor.SmallOffsetLen +
	len(magic2) + cksum.Len
const magic2at = stateLen - len(magic2)

func (state *DbState) Write() uint64 {
	// NOTE: indexes should already have been saved
	offsets := state.meta.Write(state.store)
	return writeState(state.store, offsets)
}

func writeState(store *stor.Stor, offsets [meta.Noffsets]uint64) uint64 {
	stateOff, buf := store.Alloc(stateLen)
	copy(buf, magic1)
	i := len(magic1)
	t := time.Now().Unix()
	binary.BigEndian.PutUint64(buf[i:], uint64(t))
	i += dateSize
	for _, o := range offsets {
		stor.WriteSmallOffset(buf[i:], o)
		i += stor.SmallOffsetLen
	}
	i += cksum.Len
	cksum.Update(buf[:i])
	copy(buf[i:], magic2)
	i += len(magic2)
	assert.That(i == stateLen)
	return stateOff
}

func ReadState(st *stor.Stor, off uint64) (*DbState, time.Time) {
	offsets, t := readState(st, off)
	return &DbState{store: st, meta: meta.ReadOverlay(st, offsets)}, t
}

func readState(st *stor.Stor, off uint64) ([meta.Noffsets]uint64, time.Time) {
	buf := st.Data(off)[:stateLen]
	i := len(magic1)
	assert.That(string(buf[:i]) == magic1)
	cksum.MustCheck(buf[:magic2at])
	assert.That(string(buf[magic2at:magic2at+len(magic2)]) == magic2)
	t := time.Unix(int64(binary.BigEndian.Uint64(buf[i:])), 0)
	i += dateSize
	var offsets [meta.Noffsets]uint64
	for j := range offsets {
		offsets[j] = stor.ReadSmallOffset(buf[i:])
		i += stor.SmallOffsetLen
	}
	return offsets, t
}
