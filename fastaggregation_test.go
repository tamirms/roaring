package roaring

// to run just these tests: go test -run TestFastAggregations*

import (
	"container/heap"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFastAggregationsSize(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	for i := uint32(0); i < 1000000; i += 3 {
		rb1.Add(i)
	}
	for i := uint32(0); i < 1000000; i += 7 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 1000000; i += 1001 {
		rb3.Add(i)
	}
	pq := make(priorityQueue, 3)
	pq[0] = &item{rb1, 0}
	pq[1] = &item{rb2, 1}
	pq[2] = &item{rb3, 2}
	heap.Init(&pq)

	assert.Equal(t, rb3.GetSizeInBytes(), heap.Pop(&pq).(*item).value.GetSizeInBytes())
	assert.Equal(t, rb2.GetSizeInBytes(), heap.Pop(&pq).(*item).value.GetSizeInBytes())
	assert.Equal(t, rb1.GetSizeInBytes(), heap.Pop(&pq).(*item).value.GetSizeInBytes())
}

func TestFastAggregationsCont(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	for i := uint32(0); i < 10; i += 3 {
		rb1.Add(i)
	}
	for i := uint32(0); i < 10; i += 7 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 10; i += 1001 {
		rb3.Add(i)
	}
	for i := uint32(1000000); i < 1000000+10; i += 1001 {
		rb1.Add(i)
	}
	for i := uint32(1000000); i < 1000000+10; i += 7 {
		rb2.Add(i)
	}
	for i := uint32(1000000); i < 1000000+10; i += 3 {
		rb3.Add(i)
	}
	rb1.Add(500000)
	pq := make(containerPriorityQueue, 3)
	pq[0] = &containeritem{rb1, 0, 0}
	pq[1] = &containeritem{rb2, 0, 1}
	pq[2] = &containeritem{rb3, 0, 2}
	heap.Init(&pq)
	expected := []int{6, 4, 5, 6, 5, 4, 6}
	counter := 0
	for pq.Len() > 0 {
		x1 := heap.Pop(&pq).(*containeritem)
		assert.EqualValues(t, expected[counter], x1.value.GetCardinality())

		counter++
		x1.keyindex++
		if x1.keyindex < x1.value.highlowcontainer.size() {
			heap.Push(&pq, x1)
		}
	}
}

func TestFastAggregationsAdvanced_run(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	for i := uint32(500); i < 75000; i++ {
		rb1.Add(i)
	}
	for i := uint32(0); i < 1000000; i += 7 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 1000000; i += 1001 {
		rb3.Add(i)
	}
	for i := uint32(1000000); i < 2000000; i += 1001 {
		rb1.Add(i)
	}
	for i := uint32(1000000); i < 2000000; i += 3 {
		rb2.Add(i)
	}
	for i := uint32(1000000); i < 2000000; i += 7 {
		rb3.Add(i)
	}
	rb1.RunOptimize()
	rb1.Or(rb2)
	rb1.Or(rb3)
	bigand := And(And(rb1, rb2), rb3)
	bigxor := Xor(Xor(rb1, rb2), rb3)

	assert.True(t, FastOr(rb1, rb2, rb3).Equals(rb1))
	assert.True(t, HeapOr(rb1, rb2, rb3).Equals(rb1))
	assert.Equal(t, rb1.GetCardinality(), HeapOr(rb1, rb2, rb3).GetCardinality())
	assert.True(t, HeapXor(rb1, rb2, rb3).Equals(bigxor))
	assert.True(t, FastAnd(rb1, rb2, rb3).Equals(bigand))
}

func TestFastAggregationsXOR(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()

	for i := uint32(0); i < 40000; i++ {
		rb1.Add(i)
	}
	for i := uint32(0); i < 40000; i += 4000 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 40000; i += 5000 {
		rb3.Add(i)
	}

	assert.EqualValues(t, 40000, rb1.GetCardinality())

	xor1 := Xor(rb1, rb2)
	xor1alt := Xor(rb2, rb1)
	assert.True(t, xor1alt.Equals(xor1))
	assert.True(t, HeapXor(rb1, rb2).Equals(xor1))

	xor2 := Xor(rb2, rb3)
	xor2alt := Xor(rb3, rb2)
	assert.True(t, xor2alt.Equals(xor2))
	assert.True(t, HeapXor(rb2, rb3).Equals(xor2))

	bigxor := Xor(Xor(rb1, rb2), rb3)
	bigxoralt1 := Xor(rb1, Xor(rb2, rb3))
	bigxoralt2 := Xor(rb1, Xor(rb3, rb2))
	bigxoralt3 := Xor(rb3, Xor(rb1, rb2))
	bigxoralt4 := Xor(Xor(rb1, rb2), rb3)

	assert.True(t, bigxoralt2.Equals(bigxor))
	assert.True(t, bigxoralt1.Equals(bigxor))
	assert.True(t, bigxoralt3.Equals(bigxor))
	assert.True(t, bigxoralt4.Equals(bigxor))

	assert.True(t, HeapXor(rb1, rb2, rb3).Equals(bigxor))
}

func TestFastAggregationsXOR_run(t *testing.T) {
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()

	for i := uint32(0); i < 40000; i++ {
		rb1.Add(i)
	}
	rb1.RunOptimize()
	for i := uint32(0); i < 40000; i += 4000 {
		rb2.Add(i)
	}
	for i := uint32(0); i < 40000; i += 5000 {
		rb3.Add(i)
	}

	assert.EqualValues(t, 40000, rb1.GetCardinality())

	xor1 := Xor(rb1, rb2)
	xor1alt := Xor(rb2, rb1)
	assert.True(t, xor1alt.Equals(xor1))
	assert.True(t, HeapXor(rb1, rb2).Equals(xor1))

	xor2 := Xor(rb2, rb3)
	xor2alt := Xor(rb3, rb2)
	assert.True(t, xor2alt.Equals(xor2))
	assert.True(t, HeapXor(rb2, rb3).Equals(xor2))

	bigxor := Xor(Xor(rb1, rb2), rb3)
	bigxoralt1 := Xor(rb1, Xor(rb2, rb3))
	bigxoralt2 := Xor(rb1, Xor(rb3, rb2))
	bigxoralt3 := Xor(rb3, Xor(rb1, rb2))
	bigxoralt4 := Xor(Xor(rb1, rb2), rb3)

	assert.True(t, bigxoralt2.Equals(bigxor))
	assert.True(t, bigxoralt1.Equals(bigxor))
	assert.True(t, bigxoralt3.Equals(bigxor))
	assert.True(t, bigxoralt4.Equals(bigxor))

	assert.True(t, HeapXor(rb1, rb2, rb3).Equals(bigxor))
}

func TestFastAggregationsAndAny(t *testing.T) {
	base := NewBitmap()
	rb1 := NewBitmap()
	rb2 := NewBitmap()
	rb3 := NewBitmap()
	// only one filter has some values
	from := uint32(maxCapacity * 4)
	for i := from; i < from+100; i += 2 {
		rb1.Add(i)
	}
	// only base has values
	from = maxCapacity * 7
	for i := from; i < from+100; i += 2 {
		base.Add(i)
	}
	// base and one of filters have same values
	from = maxCapacity * 8
	for i := from; i < from+100; i += 2 {
		base.Add(i)
		rb1.Add(i)
	}
	// small union
	from = maxCapacity * 10
	for i := from; i < from+1000; i += 10 {
		base.Add(i)
		base.Add(i + i%3)

		rb1.Add(i)
		rb1.Add(i + 1)

		rb2.Add(i + 2)
		rb2.Add(i + i%7)

		rb3.Add(200 + i)
	}
	// run filters
	from = maxCapacity * 10
	for i := from; i < from+1000; i += 3 {
		base.Add(i)
	}
	for i := from; i < from+100; i++ {
		rb1.Add(i)
		rb2.Add(i + 333)
		rb3.Add(i + 433)
	}
	// large union
	from = maxCapacity * 16
	for i := from; i < from+arrayDefaultMaxSize*10; i += 3 {
		base.Add(i)
		base.Add(i + i%2 + 1)
		rb2.Add(i)
		rb3.Add(i + 1)
	}

	// some extra base values
	from = maxCapacity * 17
	for i := from; i < from+1000; i++ {
		base.Add(i)
	}

	base.RunOptimize()
	rb1.RunOptimize()
	rb2.RunOptimize()
	rb3.RunOptimize()

	orFirst := base.Clone()
	orFirst.And(FastOr(rb1, rb2, rb3))

	fast := base.Clone()
	fast.AndAny(rb1, rb2, rb3)

	assert.True(t, fast.Equals(orFirst))
}

func TestFastRangeAggregations(t *testing.T) {
	andFunc := func(bitmaps ...*Bitmap) *Bitmap {
		return FastAndRange(0, MaxRange, bitmaps...)
	}
	orFunc := func(bitmaps ...*Bitmap) *Bitmap {
		return FastOrRange(0, MaxRange, bitmaps...)
	}
	testAggregations(t, andFunc, orFunc, nil)
}

// clipToRange removes every value of b outside [rangeStart, rangeEnd), with
// the same clamping as the range aggregations.
func clipToRange(b *Bitmap, rangeStart, rangeEnd uint64) {
	if rangeEnd > MaxRange {
		rangeEnd = MaxRange
	}
	if rangeStart >= rangeEnd {
		b.Clear()
		return
	}
	b.RemoveRange(0, rangeStart)
	b.RemoveRange(rangeEnd, MaxRange)
}

// referenceOrRange is the materialize-then-clip computation FastOrRange
// replaces.
func referenceOrRange(rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) *Bitmap {
	answer := FastOr(bitmaps...)
	clipToRange(answer, rangeStart, rangeEnd)
	return answer
}

// referenceAndRange is the materialize-then-clip computation FastAndRange
// replaces.
func referenceAndRange(rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) *Bitmap {
	answer := FastAnd(bitmaps...)
	clipToRange(answer, rangeStart, rangeEnd)
	return answer
}

// checkRangeAggregation runs both range aggregations of bitmaps over
// [rangeStart, rangeEnd), checks that the inputs are left untouched, and
// compares the validated results with the reference. The reference runs last
// because FastOr and FastAnd clone a single input, and Clone flags the
// containers of a copy-on-write source.
func checkRangeAggregation(t *testing.T, rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) {
	t.Helper()
	snapshots := make([]roaringArraySnapshot, len(bitmaps))
	for i, b := range bitmaps {
		snapshots[i] = snapshotRoaringArray(&b.highlowcontainer)
	}
	gotOr := FastOrRange(rangeStart, rangeEnd, bitmaps...)
	gotAnd := FastAndRange(rangeStart, rangeEnd, bitmaps...)
	for i, b := range bitmaps {
		assertRoaringArrayUnchanged(t, snapshots[i], &b.highlowcontainer)
	}

	assert.NoError(t, gotOr.Validate())
	assert.False(t, gotOr.GetCopyOnWrite())
	wantOr := referenceOrRange(rangeStart, rangeEnd, bitmaps...)
	assert.True(t, gotOr.Equals(wantOr), "FastOrRange(%d, %d): got %s, want %s", rangeStart, rangeEnd, gotOr, wantOr)

	assert.NoError(t, gotAnd.Validate())
	assert.False(t, gotAnd.GetCopyOnWrite())
	wantAnd := referenceAndRange(rangeStart, rangeEnd, bitmaps...)
	assert.True(t, gotAnd.Equals(wantAnd), "FastAndRange(%d, %d): got %s, want %s", rangeStart, rangeEnd, gotAnd, wantAnd)
}

// rangeFixtures returns bitmaps that together cover every container
// representation and key layout the range aggregations distinguish: array
// containers at keys 0-2, bitmap containers at keys 3-5, a run that straddles
// the container boundaries of keys 6-8 plus a short run at key 9, a mixed
// bitmap whose keys interleave with the others and leave gaps, values at the
// top of the value space, and keys far away from everything else.
func rangeFixtures() []*Bitmap {
	arrays := New()
	for key := uint32(0); key < 3; key++ {
		for i := uint32(0); i < 3855; i++ {
			arrays.Add(key<<16 + i*17)
		}
	}

	bitmaps := New()
	for key := uint32(3); key < 6; key++ {
		for i := uint32(0); i < 1<<16; i += 3 {
			bitmaps.Add(key<<16 + i)
		}
	}

	runs := New()
	runs.AddRange(6<<16+1000, 8<<16+5000)
	runs.AddRange(9<<16+100, 9<<16+200)
	runs.RunOptimize()

	mixed := New()
	mixed.AddMany([]uint32{5, 10, 15, 2 << 16, 2<<16 + 1})
	for i := uint32(0); i < 1<<16; i += 2 {
		mixed.Add(4<<16 + i)
	}
	mixed.AddRange(7<<16+20000, 7<<16+40000)
	mixed.AddRange(8<<16+4000, 8<<16+6000)
	mixed.AddRange(12<<16+10, 12<<16+20)
	mixed.RunOptimize()

	top := New()
	top.AddRange(MaxUint32-1000, MaxRange)
	top.Add(MaxUint32 - 70000)
	top.RunOptimize()

	far := New()
	far.AddMany([]uint32{100 << 16, 100<<16 + 1, 200<<16 + 7})

	return []*Bitmap{arrays, bitmaps, runs, mixed, top, far}
}

// rangeWindows returns the [rangeStart, rangeEnd) pairs exercised against the
// fixtures: empty ranges, ranges outside every key, ranges inside single
// containers of each representation, ranges cut by runs on both edges, ranges
// on and across container boundaries, and ranges at the top of the value
// space, including a rangeEnd past MaxRange.
func rangeWindows() []struct {
	name       string
	start, end uint64
} {
	return []struct {
		name       string
		start, end uint64
	}{
		{"empty start==end", 5, 5},
		{"empty start>end", 10, 3},
		{"start beyond MaxUint32", 1 << 33, 1 << 34},
		{"key gap", 13 << 16, 100 << 16},
		{"above every key", 201 << 16, MaxRange},
		{"inside array container", 100, 5000},
		{"inside bitmap container", 3<<16 + 100, 3<<16 + 5000},
		{"inside partial run", 6<<16 + 2000, 6<<16 + 3000},
		{"inside full run", 7<<16 + 100, 7<<16 + 5000},
		{"run straddles both edges", 6<<16 + 50000, 8<<16 + 3000},
		{"across array containers", 1<<16 - 100, 1<<16 + 100},
		{"array to bitmap boundary", 2<<16 + 60000, 3<<16 + 100},
		{"bitmap to run boundary", 5<<16 + 60000, 6<<16 + 2000},
		{"container aligned", 1 << 16, 3 << 16},
		{"one aligned container", 4 << 16, 5 << 16},
		{"one aligned full run", 7 << 16, 8 << 16},
		{"single present value", 3<<16 + 3, 3<<16 + 4},
		{"single absent value", 3<<16 + 4, 3<<16 + 5},
		{"top of value space", MaxUint32 - 5, MaxRange},
		{"last value only", MaxUint32, MaxRange},
		{"top container boundary", MaxUint32 - 100000, MaxRange},
		{"end clamped", MaxUint32 - 5, 1 << 40},
		{"whole clamped", 0, 1 << 40},
		{"whole", 0, MaxRange},
		{"wide", 1<<16 + 5, 200<<16 + 7},
	}
}

func TestFastRangeAggregationsWindows(t *testing.T) {
	fixtures := rangeFixtures()
	empty := New()
	subsets := map[string][]*Bitmap{
		"none":       {},
		"empty":      {empty},
		"empty+runs": {empty, fixtures[2]},
		"twice":      {fixtures[3], fixtures[3]},
		"all":        fixtures,
	}
	for i, a := range fixtures {
		subsets[fmt.Sprintf("single%d", i)] = []*Bitmap{a}
		for j, b := range fixtures[i+1:] {
			subsets[fmt.Sprintf("pair%d%d", i, i+1+j)] = []*Bitmap{a, b}
		}
	}
	for name, bitmaps := range subsets {
		for _, w := range rangeWindows() {
			t.Run(name+"/"+w.name, func(t *testing.T) {
				checkRangeAggregation(t, w.start, w.end, bitmaps...)
			})
		}
	}
}

// roaringArraySnapshot records everything a range aggregation must leave
// alone in an input: the copy-on-write setting, the keys, the container
// pointers and their contents, and the per-container copy-on-write flags.
type roaringArraySnapshot struct {
	copyOnWrite bool
	keys        []uint16
	containers  []container
	contents    []container
	flags       []bool
}

func snapshotRoaringArray(ra *roaringArray) roaringArraySnapshot {
	s := roaringArraySnapshot{copyOnWrite: ra.copyOnWrite}
	s.keys = append(s.keys, ra.keys...)
	s.containers = append(s.containers, ra.containers...)
	s.flags = append(s.flags, ra.needCopyOnWrite...)
	for _, c := range ra.containers {
		s.contents = append(s.contents, c.clone())
	}
	return s
}

func assertRoaringArrayUnchanged(t *testing.T, s roaringArraySnapshot, ra *roaringArray) {
	t.Helper()
	assert.Equal(t, s.copyOnWrite, ra.copyOnWrite, "copy-on-write setting changed")
	if !assert.Equal(t, len(s.keys), len(ra.keys), "number of containers changed") {
		return
	}
	for i := range s.keys {
		assert.Equal(t, s.keys[i], ra.keys[i], "key %d changed", i)
		assert.Equal(t, s.flags[i], ra.needCopyOnWrite[i], "copy-on-write flag %d changed", i)
		assert.True(t, s.containers[i] == ra.containers[i], "container %d was replaced", i)
		assert.True(t, s.contents[i].equals(ra.containers[i]), "container %d was modified", i)
	}
}

// rangeInputs returns the fixtures with the given copy-on-write setting, plus
// inputs whose containers are already flagged copy-on-write: a Clone of a
// copy-on-write fixture (Clone flags both sides) and a FromBuffer-backed
// bitmap.
func rangeInputs(cow bool) []*Bitmap {
	inputs := rangeFixtures()
	for _, b := range inputs {
		b.SetCopyOnWrite(cow)
	}
	if cow {
		inputs = append(inputs, inputs[0].Clone())
	}
	buf, err := inputs[1].ToBytes()
	if err != nil {
		panic(err)
	}
	backed := New()
	if _, err := backed.FromBuffer(buf); err != nil {
		panic(err)
	}
	return append(inputs, backed)
}

func TestFastRangeAggregationsInputsUntouched(t *testing.T) {
	windows := []struct {
		name       string
		start, end uint64
	}{
		{"narrow", 3<<16 + 100, 3<<16 + 5000},
		{"straddling", 1<<16 - 100, 7<<16 + 100},
		{"whole", 0, MaxRange},
		{"empty", 7, 7},
		{"top", MaxUint32 - 5000, 1 << 40},
	}
	for _, cow := range []bool{false, true} {
		inputs := rangeInputs(cow)
		subsets := map[string][]*Bitmap{
			"all": inputs,
		}
		for i, b := range inputs {
			subsets[fmt.Sprintf("single%d", i)] = []*Bitmap{b}
		}
		for i := 0; i+1 < len(inputs); i++ {
			subsets[fmt.Sprintf("pair%d", i)] = []*Bitmap{inputs[i], inputs[i+1]}
		}
		for name, bitmaps := range subsets {
			for _, w := range windows {
				t.Run(fmt.Sprintf("cow=%v/%s/%s", cow, name, w.name), func(t *testing.T) {
					snapshots := make([]roaringArraySnapshot, len(bitmaps))
					for i, b := range bitmaps {
						snapshots[i] = snapshotRoaringArray(&b.highlowcontainer)
					}
					or := FastOrRange(w.start, w.end, bitmaps...)
					and := FastAndRange(w.start, w.end, bitmaps...)
					for i, b := range bitmaps {
						assertRoaringArrayUnchanged(t, snapshots[i], &b.highlowcontainer)
					}
					// Writes to the results must not reach the inputs, even
					// through shared containers.
					for _, r := range []*Bitmap{or, and} {
						r.AddRange(w.start, w.start+10)
						r.Add(1 << 20)
						r.RemoveRange(0, MaxRange/2)
						r.Flip(0, 1<<18)
					}
					for i, b := range bitmaps {
						assertRoaringArrayUnchanged(t, snapshots[i], &b.highlowcontainer)
					}
				})
			}
		}
	}
}

func TestFastRangeAggregationsResultIsolated(t *testing.T) {
	// Writes to an input after the call must not reach the result, even
	// through shared containers.
	for _, cow := range []bool{false, true} {
		inputs := rangeInputs(cow)
		for i, b := range inputs {
			t.Run(fmt.Sprintf("cow=%v/input%d", cow, i), func(t *testing.T) {
				or := FastOrRange(0, MaxRange, b)
				and := FastAndRange(1<<16, 8<<16, b, b)
				orWant := or.Clone()
				andWant := and.Clone()
				b.Add(2<<16 + 3)
				b.Remove(2 << 16)
				b.AddRange(4<<16+1, 4<<16+3)
				b.RemoveRange(7<<16, 8<<16)
				b.Flip(3<<16, 3<<16+64)
				assert.True(t, or.Equals(orWant))
				assert.True(t, and.Equals(andWant))
			})
		}
	}
}

// randomRangeBitmap builds a bitmap over the given number of keys where every
// container is absent, an array, a bitmap or a set of runs at random, and
// applies RunOptimize and copy-on-write at random.
func randomRangeBitmap(rng *rand.Rand, keys int) *Bitmap {
	b := New()
	var values []uint32
	for key := 0; key < keys; key++ {
		base := uint64(key) << 16
		switch rng.Intn(4) {
		case 0:
		case 1:
			values = values[:0]
			for i, n := 0, 1+rng.Intn(3000); i < n; i++ {
				values = append(values, uint32(base)+uint32(rng.Intn(1<<16)))
			}
			b.AddMany(values)
		case 2:
			values = values[:0]
			for i, n := 0, 5000+rng.Intn(20000); i < n; i++ {
				values = append(values, uint32(base)+uint32(rng.Intn(1<<16)))
			}
			b.AddMany(values)
		case 3:
			for r, runs := 0, 1+rng.Intn(4); r < runs; r++ {
				start := uint64(rng.Intn(1 << 16))
				end := start + 1 + uint64(rng.Intn(20000))
				if end > 1<<16 {
					end = 1 << 16
				}
				b.AddRange(base+start, base+end)
			}
		}
	}
	if rng.Intn(2) == 0 {
		b.RunOptimize()
	}
	if rng.Intn(2) == 0 {
		b.SetCopyOnWrite(true)
	}
	return b
}

// randomRangeBound returns a range bound that is often on or next to a
// container boundary, sometimes at the top of the value space, and otherwise
// uniform over the keys in use.
func randomRangeBound(rng *rand.Rand, keys int) uint64 {
	switch rng.Intn(5) {
	case 0:
		return uint64(rng.Intn(keys+1)) << 16
	case 1:
		return uint64(rng.Intn(keys+1))<<16 + uint64(rng.Intn(3))
	case 2:
		return uint64(1+rng.Intn(keys+1))<<16 - uint64(1+rng.Intn(2))
	case 3:
		return []uint64{MaxUint32, MaxRange, 1 << 40}[rng.Intn(3)]
	}
	return uint64(rng.Intn((keys + 1) << 16))
}

func TestFastRangeAggregationsRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(20260909))
	const keys = 5
	for iteration := 0; iteration < 250; iteration++ {
		bitmaps := make([]*Bitmap, 1+rng.Intn(4))
		for i := range bitmaps {
			if i > 0 && rng.Intn(6) == 0 {
				bitmaps[i] = bitmaps[rng.Intn(i)]
				continue
			}
			bitmaps[i] = randomRangeBitmap(rng, keys)
		}
		start, end := randomRangeBound(rng, keys), randomRangeBound(rng, keys)
		if start > end && rng.Intn(4) != 0 {
			start, end = end, start
		}
		checkRangeAggregation(t, start, end, bitmaps...)
		if t.Failed() {
			t.Fatalf("iteration %d, range [%d, %d), %d inputs", iteration, start, end, len(bitmaps))
		}
	}
}

func TestFastRangeAggregationsCorpus(t *testing.T) {
	// Every pair of the property-test corpus, as generated and run-optimized,
	// over ranges anchored on the extent of the pair.
	corpus := getBitmapCorpus()
	for i, ga := range corpus {
		for _, gb := range corpus[i:] {
			for _, optimize := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/optimized=%v", ga.name, gb.name, optimize), func(t *testing.T) {
					a, b := ga.gen(), gb.gen()
					if optimize {
						a.RunOptimize()
						b.RunOptimize()
					}
					var maxVal uint64 = 1000
					for _, x := range []*Bitmap{a, b} {
						if !x.IsEmpty() && uint64(x.Maximum())+1000 > maxVal {
							maxVal = uint64(x.Maximum()) + 1000
						}
					}
					for _, w := range [][2]uint64{
						{0, maxVal}, {maxVal / 3, maxVal / 2}, {maxVal / 2, maxVal},
						{1 << 16, 2 << 16}, {1<<16 - 1, 1<<16 + 1}, {maxVal, maxVal}, {0, MaxRange},
					} {
						checkRangeAggregation(t, w[0], w[1], a, b)
						checkRangeAggregation(t, w[0], w[1], a)
						checkRangeAggregation(t, w[0], w[1], b, a, b)
					}
				})
			}
		}
	}
}
