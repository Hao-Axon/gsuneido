// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

/*
Package stor is used to access physical storage,
normally by memory mapped file access.

Storage is chunked. Allocations may not straddle chunks.
*/
package stor

import (
	"bytes"
	"log"
	"math"
	"math/bits"
	"sync"
	"sync/atomic"

	"github.com/apmckinlay/gsuneido/util/assert"
	"github.com/apmckinlay/gsuneido/util/dbg"
	"github.com/apmckinlay/gsuneido/util/exit"
)

// Offset is an offset within storage
type Offset = uint64

// storage is the interface to different kinds of storage.
// The main implementation accesses memory mapped files.
// There is also an in memory version for testing.
type storage interface {
	// Get returns the i'th chunk of storage
	Get(chunk int) []byte
	// Close closes the storage (if necessary)
	Close(size int64, unmap bool)
}

// Stor is the externally visible storage
type Stor struct {
	impl storage
	// chunksize must be a power of two and must be initialized
	chunksize uint64
	// threshold is the offset in a chunk where we proactively get next chunk
	threshold uint64
	// shift must be initialized to match chunksize
	shift int
	// size is the currently used amount.
	size atomic.Uint64
	// chunks must be initialized up to size,
	// with at least one chunk if size is 0
	chunks atomic.Value // [][]byte
	lock   sync.Mutex
}

const closedSize = math.MaxUint64

func NewStor(impl storage, chunksize uint64, size uint64) *Stor {
	shift := bits.TrailingZeros(uint(chunksize))
	assert.That(1<<shift == chunksize) // chunksize must be power of 2
	threshold := chunksize * 3 / 4     // ???
	stor := &Stor{impl: impl, chunksize: chunksize, threshold: threshold,
		shift: shift}
	stor.size.Store(size)
	return stor
}

// Alloc allocates n bytes of storage and returns its Offset and byte slice
// Returning data here allows slicing to the correct length and capacity
// to prevent erroneously writing too far.
// If insufficient room in the current chunk, advance to next
// (allocations may not straddle chunks)
func (s *Stor) Alloc(n int) (Offset, []byte) {
	assert.That(0 < n && n <= int(s.chunksize))
	for {
		oldsize := s.size.Load()
		if oldsize == closedSize {
			log.Println("Stor: Alloc after Close")
			exit.Wait()
		}
		offset := oldsize
		newsize := offset + uint64(n)
		chunk := s.offsetToChunk(newsize)
		nchunks := s.offsetToChunk(oldsize + s.chunksize - 1)
		if chunk >= nchunks { // straddle
			chunks := s.chunks.Load().([][]byte)
			if chunk >= len(chunks) {
				s.getChunk(chunk)
			}
			offset = s.chunkToOffset(chunk)
			newsize = offset + uint64(n)
		}
		// attempt to confirm our allocation
		if s.size.CompareAndSwap(oldsize, newsize) {
			// proactively get next chunk if we passed the threshold
			i := offset & (s.chunksize - 1) // index within chunk
			if i <= s.threshold && i+uint64(n) > s.threshold {
				s.getChunk(s.offsetToChunk(offset) + 1)
			}
			return offset, s.Data(offset)[:n:n] // fast path
		}
		// another thread beat us, loop and try again
	}
}

func (s *Stor) getChunk(chunk int) {
	s.lock.Lock() // note: lock does not prevent concurrent allocations
	chunks := s.chunks.Load().([][]byte)
	if chunk >= len(chunks) {
		// no one else beat us to it
		chunks = append(chunks, s.impl.Get(chunk))
		s.chunks.Store(chunks)
	}
	s.lock.Unlock()
}

// Data returns a byte slice starting at the given offset
// and extending to the end of the chunk
// since we don't know the size of the original alloc.
func (s *Stor) Data(offset Offset) []byte {
	// The existing chunks must be mapped initially
	// since lazily mapping would require locking.
	chunk := s.offsetToChunk(offset)
	chunks := s.chunks.Load().([][]byte)
	c := chunks[chunk]
	return c[offset&(s.chunksize-1):]
}

func (s *Stor) offsetToChunk(offset Offset) int {
	return int(offset >> s.shift)
}

func (s *Stor) chunkToOffset(chunk int) Offset {
	return uint64(chunk) << s.shift
}

// Size returns the current (allocated) size of the data.
// The actual file size will be rounded up to the next chunk size.
func (s *Stor) Size() uint64 {
	size := s.size.Load()
	if size == closedSize {
		log.Println("Stor: Size after Close")
		dbg.PrintStack()
		exit.Wait()
	}
	return size
}

// FirstOffset searches forewards from a given offset for a given byte slice
// and returns the offset, or 0 if not found
func (s *Stor) FirstOffset(off uint64, str string) uint64 {
	b := []byte(str)
	chunks := s.chunks.Load().([][]byte)
	c := s.offsetToChunk(off)
	n := off & (s.chunksize - 1)
	for ; c < len(chunks); c++ {
		buf := chunks[c][n:]
		if i := bytes.Index(buf, b); i != -1 {
			return uint64(c)*s.chunksize + n + uint64(i)
		}
		n = 0
	}
	return 0
}

// LastOffset searches backwards from a given offset for a given byte slice
// and returns the offset, or 0 if not found.
// It is used by repair and by asof/history.
func (s *Stor) LastOffset(off uint64, str string) uint64 {
	b := []byte(str)
	chunks := s.chunks.Load().([][]byte)
	c := s.offsetToChunk(off)
	n := off & (s.chunksize - 1)
	for ; c >= 0; c-- {
		buf := chunks[c][:n]
		if i := bytes.LastIndex(buf, b); i != -1 {
			return uint64(c)*s.chunksize + uint64(i)
		}
		n = s.chunksize
	}
	return 0
}

type writable interface {
	Write(off uint64, data []byte)
}

func (s *Stor) Write(off uint64, data []byte) {
	if w, ok := s.impl.(writable); ok {
		w.Write(off, data)
	} else {
		copy(s.Data(off), data) // for testing with heap stor
	}
}

func (s *Stor) Close(unmap bool, callback ...func(uint64)) {
	var size uint64
	if _, ok := s.impl.(*heapStor); ok {
		size = s.size.Load() // for tests
	} else {
		size = s.size.Swap(closedSize)
	}
	if size != closedSize {
		for _, f := range callback {
			f(size)
		}
		s.impl.Close(int64(size), unmap)
	}
}
