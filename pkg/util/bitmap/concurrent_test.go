// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bitmap

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentBitmapSet(t *testing.T) {
	const loopCount = 1000
	const interval = 2

	bm := NewConcurrentBitmap(loopCount * interval)
	wg := &sync.WaitGroup{}
	for i := range loopCount {
		wg.Add(1)
		go func(bitIndex int) {
			bm.Set(bitIndex)
			wg.Done()
		}(i * interval)
	}
	wg.Wait()

	for i := range loopCount {
		if i%interval == 0 {
			assert.Equal(t, true, bm.UnsafeIsSet(i))
		} else {
			assert.Equal(t, false, bm.UnsafeIsSet(i))
		}
	}
}

// TestConcurrentBitmapUniqueSetter checks if isSetter is unique everytime
// when a bit is set.
func TestConcurrentBitmapUniqueSetter(t *testing.T) {
	const loopCount = 10000
	const competitorsPerSet = 50

	wg := &sync.WaitGroup{}
	bm := NewConcurrentBitmap(32)
	var setterCounter uint64
	var clearCounter uint64
	// Concurrently set bit, and check if isSetter count matches zero clearing count.
	for range loopCount {
		// Clear bitmap to zero.
		if atomic.CompareAndSwapUint32(&(bm.segments[0]), 0x00000001, 0x00000000) {
			atomic.AddUint64(&clearCounter, 1)
		}
		// Concurrently set.
		for range competitorsPerSet {
			wg.Add(1)
			go func() {
				if bm.Set(31) {
					atomic.AddUint64(&setterCounter, 1)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	assert.Less(t, clearCounter, uint64(loopCount))
	assert.Equal(t, setterCounter, clearCounter+1)
}

// TestResetConcurrentBitmap test the reset of concurrentBitmap.
func TestResetConcurrentBitmap(t *testing.T) {
	bm := NewConcurrentBitmap(32)
	bm.Set(1)
	bm.Set(3)
	bm.Set(7)
	bm.Set(16)
	bm.Reset(8)
	assert.Equal(t, bm.bitLen, 8)
	assert.Equal(t, bm.UnsafeIsSet(1), false)
	assert.Equal(t, bm.UnsafeIsSet(3), false)
	assert.Equal(t, bm.UnsafeIsSet(7), false)
}
