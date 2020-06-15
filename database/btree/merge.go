// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package btree

// Merge combines an fbtree with an mbtree to produce a new fbtree.
// It does not modify the original fbtree or mbtree.
// TODO deletes
func Merge(fb *fbtree, mb *mbtree) *fbtree {
	return fb.Update(func(up *fbupdate) {
		mb.ForEach(func(key string, off uint64) {
			up.Insert(key, off)
		})
	})
}