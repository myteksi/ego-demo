// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

import "C"
import (
	"container/heap"
	"sync"
	"sync/atomic"
)

// TL;DR: clutch is mostly an optimised, thread safe map[uint64]interface{}.
//
// clutch is a self-compacting three-level registry with low-latency, lock-free
// O(1) lookup / insertion / removal, and BOUNDED CAPACITY of 16M entries per
// thread (adjustable at the expense of the minimum per-thread memory overhead)
// The name derives from the objective of coupling the envoy filter and the go
// filter with minimal friction. In fact, the problem it solves is very similar
// to a memory allocator that wants to return unused pages to the operating
// system. Since discovering that every C->Go call requires a mutex interaction
// anyway (see src/runtime/cgo/gcc_libinit.c:_cgo_wait_runtime_init_done), this
// may be considered a slight case of overengineering :)
//
// It is PURPOSE-BUILT as part of a C-to-Golang shim for envoy and maintains
// references to Go objects conceptually owned by envoy filter instances (C++
// objects). It makes various assumptions that may not hold in a more general
// setting. For example, lookup is free from data races only if all Get(),
// Tag(), and Remove() calls are coming from the same thread as the associated
// call to AcquireSlot(). This is guaranteed by envoy's threading model (see
// https://blog.envoyproxy.io/envoy-threading-model-a8d44b922310) where filter
// callbacks are always executed in the same thread in order to remove the need
// for locking). However, Get() is also thread safe as long as the registered
// item is guaranteed to exist, which is the case when accessing filter factory
// data upon filter creation.
//
// Registered items receive a tag (an "address") that is unique as long as the
// item is not removed from the registry. Tags are also unlikely to be reused
// with high probability, which provides a reasonable amount of protection
// against damage in the case of accidental double removals.
//
// Tags have an internal structure that helps with the retrieval of the item:
//
//   tag
//   +------+------+------+-------+
//   | mark | slot | hunk | entry |
//   +------+------+------+-------+
//    28bits 12bits 12bits  12bits
//
// These fields refer to these clutch data structures
//
//   clutch (64KB)
//   +-------+-------+-------+-------+
//   | *slot | *slot |  ...  | *slot | slots: 4K slot entries
//   +-------+-------+-------+-------+
//   | int   | int   |       | int   | free-entry-index-stack
//   +-------+-------+-------+-------+
//
//   slot (48KB)
//   +-------+-------+-------+-------+
//   | *hunk | *hunk |  ...  | *hunk | hunks: ~4K hunk pointers
//   +-------+-------+-------+-------+
//   | int   | int   |       | int   | hunk-usage-max-heap
//   +-------+-------+-------+-------+
//
//   hunk (80KB)
//   +------+------+-------+------+
//   | item | item |  ...  | item | items: ~4K items pairs
//   +------+------+-------+------+
//   | int  | int  |       | int  | free-entry-index-stack / item marks
//   +------+------+-------+------+
//
//
// hence, the item for a given tag can be retrieved using three array lookups:
//
//   entry = clutch.slots[tag.slot].hunks[tag.hunk].entries[tag.entry]
//
// As a safety, it is checked if entry.mark equals tag.mark.
//
// The first level of indirection ("slots") is used to shard the registry and
// avoid lock contention.
//
// The second level of indirection ("hunks") is used to ensure the management
// data structures ("hunk") are released quickly when they become sparsely
// populated. This is achieved by always preferring the fullest partially
// filled hunk for registering new items (this is what the hunk-index-heap is
// used for). Besides reducing memory overhead, this is also should improve
// cache locality. Each hunk requires approximately 80KB of RAM.
//
// The third level of indirection ("items") is a simple array with a free entry
// index. The LIFO nature of the free elements index is also hoped to improve
// cache locality.
//
// Since the first allocated hunk of any slot is never released, a realistic
// memory estimate is about 5MB of memory per instance when running 32 threads.
//
type clutch struct {
	sync.RWMutex // lock for slot 0

	slots [slotsLen]*slot
	next  [slotsLen]uint64
	head  uint64 // first free thread slot minus 1
}

const (
	// changing these values to something other than (hunkBase: 4096,
	// hunkCount: 4096, slotCount: 64) voids warranty. While there may be good
	// reasons for this, please do think twice.

	// spare has two purposes: for one, it reduces the struct sizes a little to
	// better fit traditional allocation boundaries, for another, it adds some
	// redundancy into the address space which is needed to make the encoding
	// more resilient. spare needs to be at least 1 and at least one of the
	// slotsLen, itemsLen or hunksLen constants need to be adjusted.
	spare = 10

	itemsBase = 1
	maxItems  = 4096 // must be < 2G
	itemsLen  = maxItems - spare

	hunkBase = maxItems * itemsBase
	maxHunks = 4096             // must be < 2G
	hunksLen = maxHunks - spare // ~16M filter instances / thread

	slotBase = maxHunks * hunkBase
	maxSlots = 4096     // 4096 threads should be enough for everyone o_0
	slotsLen = maxSlots // no spares

	markBase = maxSlots * slotBase

	// special tag to indicate out-of-space
	noAddr = ^uint32(0)
)

// AcquireSlot implements a lock-free, concurrent version of the free list
// management employed by hunk. It returns a versioned slot reference to
// avoid accidental double releases. A return value of 0 is a valid slot
// identifier, but it also indicates a problem (e.g. out of slots, or an
// implementation bug). clutch operations on slot 0 may be slower as they
// use a mutex for coordination between threads.
func (c *clutch) AcquireSlot() uint64 {

	for {
		// add 1 to reserve slot 0 for overflow and default handling
		head := atomic.LoadUint64(&c.head) + 1

		// head uses the high bits as version counter. in other words,
		// head = version * slotsLen + index. This is motivated further
		// down.
		index := head % slotsLen
		if index == 0 {
			// we have allocated slotsLen-1 slots and need to use the
			// reserved slot 0. We will fall back to Mutex locking for
			// when calling Tag() with slot 0
			return 0
		}

		next := c.next[index]
		if 0 == next {
			// not initialised --> use implied default
			next = index + 1
		}

		// c.next[index] this isn't ours, yet. Therefore, it may be modified
		// between now and the CompareAndSwapUint32 call, causing the value
		// held in next to become stale. For example:
		//
		//    Thread 1        Thread 2        Thread 3
		//                                    z := Acquire()
		//    Release(a)
		//    Release(b)
		//
		//    x := Acquire:
		//    ...head -> b
		//    ...next -> a
		//                    b := Acquire()
		//                    ...head -> b
		//                    ...next -> a
		//                    ...head <- a
		//                                    Release(z)
		//                                    ...head -> a
		//                                    ...next <- a
		//                                    ...head <- z
		//                    Release(b)
		//                    ...head -> z
		//                    ...next <- z
		//                    ...head <- b
		//
		//    head <- a *BANG*
		//
		// In this example, without further precautions, we would lose entry
		// z forever, because the CompareAndSwapUint32 would see head unchanged
		// and proceed.
		//
		// Therefore, we have embedded a version counter for the value of
		// c.next[index] into the relevant reference to it (c.head, or other
		// c.next[] entries). This version counter is updated in ReleaseSlot().
		//
		// With this guard in place, it would need significant time of thread
		// suspension until this counter completes a full cycle (likely longer
		// than needed to read this comment), perfect timing, and the need for
		// index to have arrive at the same value via a release sequence in
		// order to provoke a harmful collision.

		if atomic.CompareAndSwapUint64(&c.head, head-1, next-1) {

			// allocate a new slot and remember its version in order to
			// detect double free and possibly other problems.
			c.slots[index], c.next[index] = &slot{}, head

			return head
		}
	}
}

// ReleaseSlot() releases a thread slot given a versioned slot identifier.
func (c *clutch) ReleaseSlot(head uint64) {

	if 0 == head {
		// don't release slot 0.
		return
	}

	index := head % slotsLen // strip version
	if c.next[index] != head {
		// FIXME: log corrupt / double free
		return
	}

	c.slots[index].release()
	c.slots[index] = nil

	head = head + slotsLen // increment version count of future head
	for {
		next := atomic.LoadUint64(&c.head) + 1 // load future head.next
		c.next[index] = next                   // we still own the slot, so this is safe

		// we use a simple mod count to avoid ABA issues. This is not safe in
		// general, but we know that Release() is called only when a thread
		// terminates. Since this is a rather heavy operation, we can assume
		// that wrap-around won't occur while another thread is retrying its
		// Acquire() operation.
		if atomic.CompareAndSwapUint64(&c.head, next-1, head-1) {
			return
		}
	}
}

// TagItem stores a reference to item and returns a tag for it using which it
// can be retrieved. head must be a value returned by AcquireSlot().
// A return value of 0 indicates the item could not be registered.
func (c *clutch) TagItem(head uint64, item interface{}) uint64 {

	if nil == item {
		return 0
	}

	if head == 0 {
		// something went wrong booking a slot --> use shared one
		c.Lock()
		defer c.Unlock()
		if nil == c.slots[0] {
			c.slots[0] = &slot{}
		}
	}

	index := head % slotsLen // strip version
	slot := c.slots[index]
	if c.next[index] != head || nil == slot {
		// ouch. slot reference expired or invalid slot 0 reference, or wtf.
		// TODO: log/stat

		// anyway, please don't crash if we can help it.
		if 0 == head {
			return 0
		}
		return c.TagItem(0, item)
	}

	// generate simplistic "checksum" that fits into the free tag bits
	mark := uint32((slot.version*markBase + head*0x9e3779b9) / markBase)

	// allocate an address
	addr := slot.add(item, mark)
	if noAddr == addr {
		// everything booked. not good.
		// TODO: log/stat

		// try to buy some time...
		if 0 == head {
			return 0
		}
		return c.TagItem(0, item)
	}

	slot.version++

	// assemble & return encode tag value.
	// NOTE: if spare were zero, there is a corner case where we would return
	//       a tag value of 0.
	return uint64(mark)*markBase + index*slotBase + uint64(addr) + spare/spare
}

// GetItem returns a value previously registered with `tag`, or nil if no value
// was registered for `tag` or it was removed in the meantime (and no value has
// received the same tag -- which is possible but unlikely).
func (c *clutch) GetItem(tag uint64) interface{} {
	return c.get(tag, false)
}

// RemoveItem removes and returns a value previously registered with `tag`, or
// returns nil if no value was registered for `tag` or it was removed in the
// meantime (and no value has received the same tag -- which is possible but
// unlikely).
func (c *clutch) RemoveItem(tag uint64) interface{} {
	return c.get(tag, true)
}

// get() returns the item associated with tag, or nil if no such entry could
// be found. The entry will be removed if (and only if) `remove` is `true`.
func (c *clutch) get(tag uint64, remove bool) interface{} {

	if 0 == tag {
		return nil
	}

	tag -= spare / spare // sames as 1 but ensures 0 != spare

	mark := tag / markBase
	tag -= mark * markBase

	slot := tag / slotBase
	tag -= slot * slotBase

	if slotsLen <= slot {
		// TODO: log/stat corruption
		return nil
	}

	// shared slot?
	if 0 == slot {
		if remove {
			c.Lock()
			defer c.Unlock()
		} else {
			c.RLock()
			defer c.RUnlock()
		}
	}

	return c.slots[slot].get(uint32(tag), uint32(mark), remove)
}

// slot implements heap.Interface for hunks.free to sort hunks by use count.
// This is done in order to expedite the release of hunks that are already
// less full than others (because they won't receive new items until they
// become the fullest hunk).
type slot struct {
	hunks   [hunksLen]*hunk
	free    [hunksLen]uint32
	version uint64
}

// Len implements heap.Interface.Len
func (s *slot) Len() int {
	return hunksLen
}

// Swap implements heap.Interface.Swap
func (s *slot) Swap(i, j int) {

	// s.free is a heap with indexes to free entries (increased by 1
	// to avoid the need for init).
	if fi, fj := s.free[i], s.free[j]; fi != fj {

		if fi == 0 {
			fi = uint32(i + 1) // uninitialised --> use default
		} else if hi := s.hunks[fi-1]; hi != nil {
			hi.rank = j
		}

		if fj == 0 {
			fj = uint32(j + 1) // uninitialised --> use default
		} else if hj := s.hunks[fj-1]; hj != nil {
			hj.rank = i
		}

		s.free[i], s.free[j] = fj, fi
	}
}

// key defines our heap order
func (s *slot) key(i int) int {

	// construct a key such that the fullest hunks go to the beginning
	// but completely filled hunks go to the end

	if f := s.free[i]; f != 0 {
		// s.free is a heap with indexes to free entries (increased by 1
		// to avoid the need for init).
		index := f - 1
		if hunk := s.hunks[index]; hunk != nil {
			if used := hunk.used; used < itemsLen {
				return used + 1
			}
			// queue full hunks last
			return -1
		}
	}

	// queue absent hunks after empty hunks
	return 0
}

// Less implements heap.Interface.Less
func (s *slot) Less(i, j int) bool {
	return s.key(j) < s.key(i) // max heap!
}

// Push implements heap.Interface.Push
func (s *slot) Push(x interface{}) { panic("don't push hunks") }

// Pop implements heap.Interface.Pop
func (s *slot) Pop() interface{} { panic("don't pop hunks") }

// add registers item, and associate it with the given mark. The returned value
// encodes the item's location in the slot.
func (s *slot) add(item interface{}, mark uint32) uint32 {

	// pick the available hunk with least free entries.
	// s.free is a heap with indexes to free entries, but increased by 1
	// to avoid the need for init).
	f := s.free[0]

	if 0 == f {
		// The heap was not initialised. Swap will always initialise the hunk
		// indexes it touches, so this can only happen if there are no elements
		// on the heap, yet.
		f = 1
		s.free[0] = f
		s.hunks[0] = &hunk{}

		// Block some entries in the first hunk we allocated. This is
		// intended to avoid a corner cases where after a spike in tags
		// items are removed under an even distribution across all hunks.
		//
		// In this setting, we could end up with a sparsely populated set
		// of hunks that won't go away. By blocking some of the first hunks
		// entries, it is forced to the head of the list under this
		// scenario.
		//
		// As a side effect, this also prevents the first hunk from ever being
		// released, which is useful for low-concurrency high-frequency cases
		s.hunks[0].used += itemsLen / 2
	}

	index := f - 1
	h := s.hunks[index]
	if nil == h {
		// no hunk allocated, yet (or freed).
		h = &hunk{}
		s.hunks[index] = h
	}

	// try to place item in the hunk. This will fail if the hunk is full
	addr := h.add(item, mark)
	if noAddr == addr {
		// based on our heap order (key()), if this hunk is full, all
		// hunks must be full.
		return noAddr
	}

	if h.used == itemsLen {
		// we only need to update the heap in case the hunk became unavailable
		// because in all other cases, the hunk will remain a maximal element
		heap.Fix(s, 0)
	}

	// encode the item location and return the tag suffix
	return index*hunkBase + addr
}

// get returns the item addressed by tag if its mark matches, or nil if no
// such entry could be found. The entry will be removed if (and only if)
// `remove` is `true`.
func (s *slot) get(addr, mark uint32, remove bool) (result interface{}) {

	// decode the item's hunk index from the tag suffix
	index := addr / hunkBase
	addr -= index * hunkBase

	if hunksLen <= index {
		// TODO: log/stat corruption
		return nil
	}

	if h := s.hunks[index]; nil != h {
		// retrieve the item from its hunk based on the remaining addr bits
		result = h.get(addr, mark, remove)
		if result != nil && remove {
			heap.Fix(s, int(h.rank)) // adjust free list position and h.rank
			if 0 == h.used {
				s.hunks[index] = nil // release hunk
			}
		}
	}
	return
}

// release checks if the slot is clear and makes some noise, otherwise
func (s *slot) release() {
	// TODO: double check everything is gone and log/stat otherwise
}

// hunk is a simple item lookup with free entry management. Free entries are
// indexed via next, such that the first free element is at `head`, the next
// free element is at ^next[head], the next at ^next[^next[head]] etc., except
// when the value of next[index] is 0, in which case the next free entry is
// located at `next[index + 1]``. This means, recycled entries are always
// allocated in the reverse order of their release.
//
// While allocated, `next[index]` stores a marker that helps to detect the
// use of stale tags. When retrieving or deleting an item, this marker must
// match the value used when the item was added.
//
type hunk struct {
	items [itemsLen]interface{}
	next  [itemsLen]uint32 // free list / mark storage
	head  uint32           // head of free list
	rank  int              // position in free heap of owning slot
	used  int              // number of used entries
}

// find returns the item addressed by tag if its mark matches, or nil if no
// such entry could be found. The entry will be removed if (and only if)
// `remove` is `true`.
func (h *hunk) get(addr uint32, mark uint32, remove bool) (item interface{}) {

	index := addr / itemsBase
	if itemsLen <= index {
		// TODO: log/stat corruption
		return nil
	}

	if mark == h.next[index] {

		item = h.items[index]
		if remove {
			// release item reference
			h.items[index] = nil

			// "push" free item index (flip bits to avoid the need for init)
			h.next[index], h.head = ^h.head, index
			h.used--
		}
	} else {
		// TODO: log/stat corruption / stale tag
	}
	return
}

// tag registers item, and associate it with the given mark. The returned value
// encodes the item's location in the hunk.
func (h *hunk) add(item interface{}, mark uint32) uint32 {

	// don't proceed if we're out of free entries or if we would be wasting space

	if itemsLen <= h.used || nil == item {
		return noAddr
	}

	// "pop" next free item index

	index := h.head
	next := ^h.next[index]
	if 0 == ^next {
		// not initialised --> use implied default
		next = index + 1
	}
	h.head = next
	h.used++

	// keep a reference to item and recycle the "next" entry for remembering
	// the item mark in order to detect double free and possibly other issues.
	h.items[index], h.next[index] = item, mark

	return index * itemsBase
}
