// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type clutchTestItem struct {
	tag, data uint64
}

type clutchSlotTester struct {
	*clutch
	slot  uint64
	items map[uint64]clutchTestItem
	t     *testing.T
	seq   uint64
}

func newClutchSlotTester(t *testing.T, c *clutch) *clutchSlotTester {
	s := clutchSlotTester{
		clutch: c,
		slot:   c.AcquireSlot(),
		items:  map[uint64]clutchTestItem{},
		t:      t,
		seq:    1867324,
	}
	return &s
}

func (s *clutchSlotTester) release() {
	s.clutch.ReleaseSlot(s.slot)
}

// add checks tag release for possible problems
// (provided all releases were done via del())
func (s *clutchSlotTester) add() uint64 {

	s.seq++
	tag := s.clutch.TagItem(s.slot, s.seq)
	if !assert.NotEqual(s.t, 0, tag, "Could not add tag (seq:%v)", s.seq) {
		return 0
	}

	index := tag & (markBase - 1) // insider knowledge: lower 24 bits must be unique

	old, found := s.items[index]

	if !assert.Equal(s.t, false, found, "Double tag for %v(%v): old:%v new:%v", index, index&(slotBase-1), old.tag, tag) {
		return 0
	}

	s.items[index] = clutchTestItem{tag: tag, data: s.seq}
	return tag
}

// del checks tag release for possible problems
// (provided it was obtained via add())
func (s *clutchSlotTester) del(tag uint64) {

	index := tag & (markBase - 1)
	old, found := s.items[index]

	if !assert.Equal(s.t, true, found, "No tag for %v(%v): old:%v new:%v", index, index&(slotBase-1), old.tag, tag) {
		return
	}

	if !assert.Equal(s.t, old.tag, tag, "Bad tag for %v(%v): old:%v tag:%v", index, index&(slotBase-1), old.tag, tag) {
		return
	}

	item := s.clutch.RemoveItem(tag)
	seq, ok := item.(uint64)

	if !assert.Equal(s.t, true, ok, "Tag not found for %v(%v): %v (item:%v)", index, index&(slotBase-1), tag, item) {
		return
	}

	if !assert.Equal(s.t, old.data, seq, "Bad item for %v(%v): expected:%v actual:%v", index, index&(slotBase-1), old.data, seq) {
		return
	}

	delete(s.items, index)
}

// TestReleaseFirstTag is derived from a bug encountered in an early version
// of clutch: the sequence (+a,+b,+c,-a,-b,+a,+b) would lead to tag c being
// issued instead of b because releasing tag at index 0 a would break the
// internal free list and lead to reassignment of slot c.
func TestReleaseFirstTag(t *testing.T) {
	s := newClutchSlotTester(t, &clutch{})
	defer s.release()

	// +1, +2, +3 (add three tags, indexes should be 1,2,3 based on inside know)
	a := s.add()
	if t.Failed() {
		return
	}

	b := s.add()
	if t.Failed() {
		return
	}

	s.add()
	if t.Failed() {
		return
	}

	// -1 (release first tag)
	s.del(a)
	if t.Failed() {
		return
	}

	// -2 (release second tag)
	s.del(b)
	if t.Failed() {
		return
	}

	// +2 (should be index 2 based on inside knowledge)
	s.add()
	if t.Failed() {
		return
	}

	// +1 (should be index 1 based on inside knowledge)
	s.add()
	if t.Failed() {
		return
	}
}

// TestReleaseAll allocatate two hunks worth of tags (which will spread across
// three hunks, actually), and releases those using different patterns.
// 123456789ABCDEF, 2468ACE13579BDF, 369CF147AD258BE, 48C159D26AE37BF,
// 5AF16B26B38D49E, 6C17D28E39F4A5B, 7E18F293A4B5C6D, and the same in reverse.
func TestReleaseAll(t *testing.T) {
	s := newClutchSlotTester(t, &clutch{})
	defer s.release()

	for direction := -1; direction < 2; direction += 2 {
		for stride := 1; stride < 8; stride++ {
			for n := 0; n < 2; n++ {
				tags := [2 * itemsLen]uint64{}
				for i := range tags {
					tags[i] = s.add()
					if t.Failed() {
						return
					}
				}

				for m := 0; m < stride; m++ {
					for i := range tags {
						if m == i%stride {
							if direction < 0 {
								i = len(tags) - 1 - i
							}
							s.del(tags[i])
							if t.Failed() {
								return
							}
						}
					}
				}
			}
			if !assert.Equal(s.t, 0, len(s.items), "Broken test case") {
				return
			}
		}
	}
}

// TestAllocMax allocatate the maximum number of items for a slot and verifies
// that overflow to slot 0 is happening as planned
func TestAllocMax(t *testing.T) {
	s := newClutchSlotTester(t, &clutch{})
	defer s.release()

	const maxSlotItems = itemsLen*hunksLen - itemsLen/2 // insider knowledge

	// fill up slot 1
	for i := 0; i < maxSlotItems; i++ {
		// don't use s.add() to avoid bookkeeping overhead
		tag := s.clutch.TagItem(s.slot, struct{}{})
		index := tag & (markBase - 1)
		slot := index / slotBase
		if !assert.Equal(s.t, s.slot, slot) {
			return
		}
	}

	// this should go the shared slot
	tag := s.clutch.TagItem(s.slot, s.seq)
	if !assert.NotEqual(s.t, uint64(0), tag) {
		return
	}

	// validate slot is 0
	index := tag & (markBase - 1)
	slot := index / slotBase
	if !assert.Equal(s.t, uint64(0), slot) {
		return
	}
}

func TestAcquireSlotSingleThread(t *testing.T) {
	tcs := []struct {
		name           string
		numberAcquire  int
		startExpected  uint64
		endExpected    uint64
		appendExpected []uint64
	}{
		{
			name:          "should start with one",
			numberAcquire: 1,
			startExpected: 1,
			endExpected:   1,
		},
		{
			name:          "should contains two values [1, 2]",
			numberAcquire: 2,
			startExpected: 1,
			endExpected:   2,
		},
		{
			name:          "should contains three values [1, 2, 3]",
			numberAcquire: 3,
			startExpected: 1,
			endExpected:   3,
		},
		{
			name:          "should contains 1000 values from [1 -> 1000]",
			numberAcquire: 1000,
			startExpected: 1,
			endExpected:   1000,
		},
		{
			name:           "should contains 4096 values from [1 -> 4095, 0]",
			numberAcquire:  4096,
			startExpected:  1,
			endExpected:    4095,
			appendExpected: []uint64{0},
		},
		{
			name:           "should contains 5000 values from [1 -> 4095, 0, 0, 0, 0, 0]",
			numberAcquire:  5000,
			startExpected:  1,
			endExpected:    4095,
			appendExpected: []uint64{0, 0, 0, 0, 0},
		},
	}
	for _, v := range tcs {
		tc := v
		t.Run(tc.name, func(t *testing.T) {
			clutch := &clutch{}
			// Build expected
			expected := make([]uint64, tc.numberAcquire)
			expectedIndex := 0
			for i := tc.startExpected; i <= tc.endExpected; i, expectedIndex = i+1, expectedIndex+1 {
				expected[expectedIndex] = i
			}
			for i := 0; i < len(tc.appendExpected); i, expectedIndex = i+1, expectedIndex+1 {
				expected[expectedIndex] = tc.appendExpected[i]
			}
			// Build results
			results := make([]uint64, tc.numberAcquire)
			for i := 0; i < tc.numberAcquire; i++ {
				results[i] = clutch.AcquireSlot()
			}
			// Check results
			assert.Equal(t, expected, results)
		})
	}
}

func TestReleaseSlotSingleThread(t *testing.T) {
	type Op struct {
		accquire    bool
		releaseSlot uint64
	}
	tcs := []struct {
		name     string
		ops      []Op
		expected []uint64
	}{
		{
			name: "should not increase version if release slot 0",
			ops: []Op{
				Op{accquire: true},
				Op{accquire: true},
				Op{releaseSlot: 0},
				Op{accquire: true},
			},
			expected: []uint64{1, 2, 0, 3},
		},
		{
			name: "should increase version after release slot 1",
			ops: []Op{
				Op{accquire: true},
				Op{accquire: true},
				Op{releaseSlot: 1},
				Op{accquire: true},
			},
			expected: []uint64{1, 2, 0, 4097},
		},
		{
			name: "should do not thing with double free slot-2",
			ops: []Op{
				Op{accquire: true},
				Op{accquire: true},
				Op{accquire: true},
				Op{releaseSlot: 2},
				Op{releaseSlot: 2},
				Op{accquire: true},
			},
			expected: []uint64{1, 2, 3, 0, 0, 4098},
		},
		{
			name: "should only increase version on slot-1 two times",
			ops: []Op{
				Op{accquire: true},
				Op{accquire: true},
				Op{releaseSlot: 1},
				Op{accquire: true},
				Op{accquire: true},
				Op{releaseSlot: 4097},
				Op{accquire: true},
				Op{accquire: true},
				Op{accquire: true},
			},
			expected: []uint64{1, 2, 0, 4097, 3, 0, 8193, 4, 5},
		},
		{
			name: "should only increase version of last slot",
			ops: []Op{
				Op{accquire: true},
				Op{accquire: true},
				Op{accquire: true},
				Op{accquire: true},
				Op{accquire: true},
				Op{releaseSlot: 5},
				Op{accquire: true},
			},
			expected: []uint64{1, 2, 3, 4, 5, 0, 4101},
		},
	}

	for _, v := range tcs {
		tc := v
		t.Run(tc.name, func(t *testing.T) {
			clutch := &clutch{}
			results := make([]uint64, len(tc.ops))
			for i, op := range tc.ops {
				slot := uint64(0)
				if op.accquire {
					slot = clutch.AcquireSlot()
				} else {
					clutch.ReleaseSlot(op.releaseSlot)
				}
				results[i] = slot
			}
			// Check results
			assert.Equal(t, tc.expected, results)
		})
	}
}

func TestReleaseSlotSingleThreadFullRoundTrip(t *testing.T) {
	clutch := &clutch{}
	numberTimes := 5000
	maxSlot := 4096
	resultsRound1 := make([]uint64, numberTimes)
	expectedRound1 := make([]uint64, numberTimes)

	resultsRound2 := make([]uint64, numberTimes)
	expectedRound2 := make([]uint64, numberTimes)

	// Accquire round1
	for i := 0; i < numberTimes; i++ {
		slot := clutch.AcquireSlot()
		expectedSlot := uint64(0)
		if i < maxSlot-1 {
			expectedSlot = uint64(i + 1)
		}
		resultsRound1[i] = slot
		expectedRound1[i] = expectedSlot
	}

	// Release all slots
	for i := numberTimes - 1; i >= 0; i-- {
		clutch.ReleaseSlot(resultsRound1[i])
	}

	// Accquire round3
	for i := 0; i < numberTimes; i++ {
		slot := clutch.AcquireSlot()
		expectedSlot := uint64(0)
		if i < maxSlot-1 {
			expectedSlot = uint64(maxSlot + i + 1)
		}
		resultsRound2[i] = slot
		expectedRound2[i] = expectedSlot
	}

	// Check results
	assert.Equal(t, expectedRound1, resultsRound1)
	assert.Equal(t, expectedRound2, resultsRound2)
}

func TestTagItemAndVerifyContent(t *testing.T) {
	type item struct {
		ID int
	}

	maxItem := 16 * 1000 * 1000
	clutch := &clutch{}
	slot := clutch.AcquireSlot()
	results := map[uint64]*item{}

	for i := 0; i < maxItem; i++ {
		item := &item{
			ID: i + 1,
		}
		tag := clutch.TagItem(slot, item)
		if !assert.NotZero(t, tag) {
			return
		}
		_, found := results[tag]
		if !assert.False(t, found) {
			return
		}
		results[tag] = item
	}
	if !assert.Equal(t, maxItem, len(results)) {
		return
	}
	for k, v := range results {
		item := clutch.GetItem(k)
		if !assert.Equal(t, v, item) {
			return
		}
	}
}
