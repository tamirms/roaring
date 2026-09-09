package roaring

import (
	"container/heap"
	"math/bits"
)

// Or function that requires repairAfterLazy
func lazyOR(x1, x2 *Bitmap) *Bitmap {
	answer := NewBitmap()
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()
main:
	for (pos1 < length1) && (pos2 < length2) {
		s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
		s2 := x2.highlowcontainer.getKeyAtIndex(pos2)

		for {
			if s1 < s2 {
				answer.highlowcontainer.appendCopy(x1.highlowcontainer, pos1)
				pos1++
				if pos1 == length1 {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
			} else if s1 > s2 {
				answer.highlowcontainer.appendCopy(x2.highlowcontainer, pos2)
				pos2++
				if pos2 == length2 {
					break main
				}
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			} else {
				c1 := x1.highlowcontainer.getContainerAtIndex(pos1)
				answer.highlowcontainer.appendContainer(s1, c1.lazyOR(x2.highlowcontainer.getContainerAtIndex(pos2)), false)
				pos1++
				pos2++
				if (pos1 == length1) || (pos2 == length2) {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			}
		}
	}
	if pos1 == length1 {
		answer.highlowcontainer.appendCopyMany(x2.highlowcontainer, pos2, length2)
	} else if pos2 == length2 {
		answer.highlowcontainer.appendCopyMany(x1.highlowcontainer, pos1, length1)
	}
	return answer
}

// In-place Or function that requires repairAfterLazy
func (x1 *Bitmap) lazyOR(x2 *Bitmap) *Bitmap {
	pos1 := 0
	pos2 := 0
	length1 := x1.highlowcontainer.size()
	length2 := x2.highlowcontainer.size()
main:
	for (pos1 < length1) && (pos2 < length2) {
		s1 := x1.highlowcontainer.getKeyAtIndex(pos1)
		s2 := x2.highlowcontainer.getKeyAtIndex(pos2)

		for {
			if s1 < s2 {
				pos1++
				if pos1 == length1 {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
			} else if s1 > s2 {
				x1.highlowcontainer.insertNewKeyValueAt(pos1, s2, x2.highlowcontainer.getContainerAtIndex(pos2).clone())
				pos2++
				pos1++
				length1++
				if pos2 == length2 {
					break main
				}
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			} else {
				c1 := x1.highlowcontainer.getWritableContainerAtIndex(pos1)
				// runContainer16.lazyIOR falls back to a slow ior path
				// (O(N log R) per merged element); promote to bitmapContainer
				// first, whose lazy union is O(1024) regardless of cardinality.
				// See https://github.com/RoaringBitmap/roaring/issues/81.
				if rc, ok := c1.(*runContainer16); ok && !rc.isFull() {
					c1 = rc.toBitmapContainer()
				}
				x1.highlowcontainer.containers[pos1] = c1.lazyIOR(x2.highlowcontainer.getContainerAtIndex(pos2))
				x1.highlowcontainer.needCopyOnWrite[pos1] = false
				pos1++
				pos2++
				if (pos1 == length1) || (pos2 == length2) {
					break main
				}
				s1 = x1.highlowcontainer.getKeyAtIndex(pos1)
				s2 = x2.highlowcontainer.getKeyAtIndex(pos2)
			}
		}
	}
	if pos1 == length1 {
		x1.highlowcontainer.appendCopyMany(x2.highlowcontainer, pos2, length2)
	}
	return x1
}

// to be called after lazy aggregates
func (x1 *Bitmap) repairAfterLazy() {
	for pos := 0; pos < x1.highlowcontainer.size(); pos++ {
		c := x1.highlowcontainer.getContainerAtIndex(pos)
		switch c.(type) {
		case *bitmapContainer:
			if c.(*bitmapContainer).cardinality == invalidCardinality {
				c = x1.highlowcontainer.getWritableContainerAtIndex(pos)
				c.(*bitmapContainer).computeCardinality()
				if c.(*bitmapContainer).getCardinality() <= arrayDefaultMaxSize {
					x1.highlowcontainer.setContainerAtIndex(pos, c.(*bitmapContainer).toArrayContainer())
				} else if c.(*bitmapContainer).isFull() {
					x1.highlowcontainer.setContainerAtIndex(pos, newRunContainer16Range(0, MaxUint16))
				}
			}
		}
	}
}

// FastAnd computes the intersection between many bitmaps quickly
// Compared to the And function, it can take many bitmaps as input, thus saving the trouble
// of manually calling "And" many times.
//
// Performance hints: if you have very large and tiny bitmaps,
// it may be beneficial performance-wise to put a tiny bitmap
// in first position.
func FastAnd(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	} else if len(bitmaps) == 1 {
		return bitmaps[0].Clone()
	}
	answer := And(bitmaps[0], bitmaps[1])
	for _, bm := range bitmaps[2:] {
		answer.And(bm)
	}
	return answer
}

// FastOr computes the union between many bitmaps quickly, as opposed to having to call Or repeatedly.
// It might also be faster than calling Or repeatedly.
func FastOr(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	} else if len(bitmaps) == 1 {
		return bitmaps[0].Clone()
	}
	answer := lazyOR(bitmaps[0], bitmaps[1])
	for _, bm := range bitmaps[2:] {
		answer = answer.lazyOR(bm)
	}
	// here is where repairAfterLazy is called.
	answer.repairAfterLazy()
	return answer
}

// FastAndRange computes the intersection of the given bitmaps restricted to
// the half-open range [rangeStart, rangeEnd). The result equals FastAnd
// followed by the removal of every value outside the range, but only the
// containers whose key lies inside the range are visited: the cost is
// proportional to the number of containers the range spans times the number
// of inputs, not to the size of the inputs.
//
// rangeEnd is clamped to MaxRange, as in CardinalityInRange. When
// rangeStart >= rangeEnd or no bitmaps are given, the result is empty.
//
// The inputs are never modified, including their copy-on-write bookkeeping.
// FastAnd with a single input calls Clone, and Clone flags every container of
// a copy-on-write source; this function does not. A container that an input
// already flags as copy-on-write is shared with the result, every other
// container of the result is freshly built, and the result has copy-on-write
// disabled.
//
// Performance hints: as with FastAnd, a tiny bitmap in first position helps,
// since the intersection at each key starts from the first input's container.
func FastAndRange(rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) *Bitmap {
	answer := NewBitmap()
	w, ok := newContainerWindow(rangeStart, rangeEnd)
	if !ok || len(bitmaps) == 0 {
		return answer
	}
	first := w.cursor(&bitmaps[0].highlowcontainer)
	if len(bitmaps) == 1 {
		for ; !first.exhausted(); first.pos++ {
			lo, hi := w.bounds(first.key())
			answer.highlowcontainer.appendRange(first.ra, first.pos, lo, hi)
		}
		return answer
	}
	second := w.cursor(&bitmaps[1].highlowcontainer)
	for !first.exhausted() && !second.exhausted() {
		s1, s2 := first.key(), second.key()
		if s1 < s2 {
			first.pos = first.ra.advanceUntil(s2, first.pos)
		} else if s1 > s2 {
			second.pos = second.ra.advanceUntil(s1, second.pos)
		} else {
			lo, hi := w.bounds(s1)
			if c := andRange(first.container(), second.container(), lo, hi); c != nil {
				answer.highlowcontainer.appendContainer(s1, c, false)
			}
			first.pos++
			second.pos++
		}
	}
	// answer now holds fresh containers under keys inside the window only,
	// so the remaining inputs are intersected in place without clipping. And
	// gallops over the other input's keys, so each pass costs the keys of
	// answer, not the keys of the input.
	for _, bm := range bitmaps[2:] {
		answer.And(bm)
	}
	return answer
}

// FastOrRange computes the union of the given bitmaps restricted to the
// half-open range [rangeStart, rangeEnd). The result equals FastOr followed by
// the removal of every value outside the range, but only the containers whose
// key lies inside the range are visited: the cost is proportional to the
// number of containers the range spans times the number of inputs, not to
// the size of the inputs.
//
// rangeEnd is clamped to MaxRange, as in CardinalityInRange. When
// rangeStart >= rangeEnd or no bitmaps are given, the result is empty.
//
// The inputs are never modified, including their copy-on-write bookkeeping.
// FastOr with a single input calls Clone, and Clone flags every container of
// a copy-on-write source; this function does not. A container that an input
// already flags as copy-on-write is shared with the result, every other
// container of the result is freshly built, and the result has copy-on-write
// disabled.
func FastOrRange(rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) *Bitmap {
	answer := NewBitmap()
	w, ok := newContainerWindow(rangeStart, rangeEnd)
	if !ok || len(bitmaps) == 0 {
		return answer
	}
	cursors := make([]rangeCursor, 0, len(bitmaps))
	for _, bm := range bitmaps {
		if c := w.cursor(&bm.highlowcontainer); !c.exhausted() {
			cursors = append(cursors, c)
		}
	}
	gathered := make([]container, 0, len(cursors))
	for len(cursors) > 0 {
		key := cursors[0].key()
		for i := 1; i < len(cursors); i++ {
			if k := cursors[i].key(); k < key {
				key = k
			}
		}
		gathered = gathered[:0]
		var single *rangeCursor
		for i := range cursors {
			if c := &cursors[i]; c.key() == key {
				gathered = append(gathered, c.container())
				single = c
			}
		}
		lo, hi := w.bounds(key)
		if len(gathered) == 1 {
			answer.highlowcontainer.appendRange(single.ra, single.pos, lo, hi)
		} else if c := orRange(gathered, lo, hi); c != nil {
			answer.highlowcontainer.appendContainer(key, c, false)
		}
		n := 0
		for i := range cursors {
			c := cursors[i]
			if c.key() == key {
				c.pos++
			}
			if !c.exhausted() {
				cursors[n] = c
				n++
			}
		}
		cursors = cursors[:n]
	}
	return answer
}

// containerWindow is a half-open value range in container coordinates: the
// keys of the first and last containers it touches and the low bits of its
// first and last values.
type containerWindow struct {
	hbStart, lbStart, hbLast, lbLast uint16
}

// newContainerWindow converts [rangeStart, rangeEnd) to container
// coordinates. rangeEnd is clamped to MaxRange. ok is false when the range is
// empty.
func newContainerWindow(rangeStart, rangeEnd uint64) (w containerWindow, ok bool) {
	if rangeEnd > MaxRange {
		rangeEnd = MaxRange
	}
	if rangeStart >= rangeEnd {
		return w, false
	}
	first := uint32(rangeStart)
	last := uint32(rangeEnd - 1)
	return containerWindow{highbits(first), lowbits(first), highbits(last), lowbits(last)}, true
}

// bounds returns the closed low-bit range [lo, hi] that the window keeps in
// the container with the given key. The key must lie in [hbStart, hbLast].
func (w containerWindow) bounds(key uint16) (lo, hi uint16) {
	lo, hi = 0, MaxUint16
	if key == w.hbStart {
		lo = w.lbStart
	}
	if key == w.hbLast {
		hi = w.lbLast
	}
	return lo, hi
}

// cursor positions a rangeCursor on the containers of ra whose key lies in
// [hbStart, hbLast].
func (w containerWindow) cursor(ra *roaringArray) rangeCursor {
	pos := ra.getIndex(w.hbStart)
	if pos < 0 {
		pos = -pos - 1
	}
	end := ra.getIndex(w.hbLast)
	if end < 0 {
		end = -end - 1
	} else {
		end++
	}
	return rangeCursor{ra, pos, end}
}

// rangeCursor walks the containers of one input inside a containerWindow: pos
// is the next index to visit and end the first index past the window.
type rangeCursor struct {
	ra       *roaringArray
	pos, end int
}

func (c *rangeCursor) exhausted() bool {
	return c.pos >= c.end
}

func (c *rangeCursor) key() uint16 {
	return c.ra.keys[c.pos]
}

func (c *rangeCursor) container() container {
	return c.ra.containers[c.pos]
}

// appendRange appends the values of sa.containers[pos] that lie in [lo, hi]
// to ra. A container the range covers entirely goes through appendCopy, which
// shares it when sa already flags it as copy-on-write and clones it otherwise.
// ra must have copy-on-write disabled: that is what keeps appendCopy from
// writing sa's flags. A partial container is freshly built and skipped when
// empty.
func (ra *roaringArray) appendRange(sa *roaringArray, pos int, lo, hi uint16) {
	if lo == 0 && hi == MaxUint16 {
		ra.appendCopy(*sa, pos)
		return
	}
	if c := clipContainer(sa.containers[pos], lo, hi); !c.isEmpty() {
		ra.appendContainer(sa.keys[pos], c, false)
	}
}

// orRange returns the union of the values of cs that lie in [lo, hi] as a
// fresh container, or nil when that union is empty. cs must hold at least two
// containers; its entries are overwritten but the containers are left
// unchanged.
func orRange(cs []container, lo, hi uint16) container {
	if lo != 0 || hi != MaxUint16 {
		cs = clipAll(cs, lo, hi)
	}
	switch len(cs) {
	case 0:
		return nil
	case 1:
		// Only reachable after clipping, so this is a fresh container.
		return cs[0]
	}
	c := cs[0].lazyOR(cs[1])
	for _, x := range cs[2:] {
		// runContainer16.lazyIOR is not lazy and is slow for long merges, so
		// promote to a bitmap container first, as (*Bitmap).lazyOR does.
		if rc, ok := c.(*runContainer16); ok && !rc.isFull() {
			c = rc.toBitmapContainer()
		}
		c = c.lazyIOR(x)
	}
	return repairAfterLazy(c)
}

// andRange returns the intersection of the values of a and b that lie in
// [lo, hi] as a fresh container, or nil when that intersection is empty. a
// and b are left unchanged.
func andRange(a, b container, lo, hi uint16) container {
	if lo != 0 || hi != MaxUint16 {
		if a = clipContainer(a, lo, hi); a.isEmpty() {
			return nil
		}
		if b = clipContainer(b, lo, hi); b.isEmpty() {
			return nil
		}
	}
	if c := a.and(b); !c.isEmpty() {
		return c
	}
	return nil
}

// clipAll replaces every entry of cs with a fresh container holding its
// values in [lo, hi] and drops the empty results. The containers themselves
// are left unchanged.
func clipAll(cs []container, lo, hi uint16) []container {
	n := 0
	for _, c := range cs {
		if c = clipContainer(c, lo, hi); !c.isEmpty() {
			cs[n] = c
			n++
		}
	}
	return cs[:n]
}

// clipContainer returns a fresh container holding the values of c that lie in
// [lo, hi]. c is left unchanged. The result may be empty and is never c.
func clipContainer(c container, lo, hi uint16) container {
	switch t := c.(type) {
	case *arrayContainer:
		return clipArrayContainer(t, lo, hi)
	case *bitmapContainer:
		return clipBitmapContainer(t, lo, hi)
	case *runContainer16:
		return clipRunContainer(t, lo, hi)
	}
	panic("unsupported container type")
}

func clipArrayContainer(ac *arrayContainer, lo, hi uint16) container {
	start := binarySearch(ac.content, lo)
	if start < 0 {
		start = -start - 1
	}
	end := binarySearch(ac.content, hi)
	if end < 0 {
		end = -end - 1
	} else {
		end++
	}
	answer := newArrayContainerCapacity(end - start)
	answer.content = append(answer.content, ac.content[start:end]...)
	return answer
}

func clipBitmapContainer(bc *bitmapContainer, lo, hi uint16) container {
	first, last := int(lo>>6), int(hi>>6)
	const allOnes = ^uint64(0)
	firstMask := allOnes << (lo & 63)
	lastMask := allOnes >> (63 - (hi & 63))
	cardinality := bc.getCardinalityInRange(uint(lo), uint(hi)+1)
	if cardinality > arrayDefaultMaxSize {
		answer := newBitmapContainer()
		copy(answer.bitmap[first:last+1], bc.bitmap[first:last+1])
		answer.bitmap[first] &= firstMask
		answer.bitmap[last] &= lastMask
		answer.cardinality = cardinality
		return answer
	}
	answer := newArrayContainerCapacity(cardinality)
	for i := first; i <= last; i++ {
		w := bc.bitmap[i]
		if i == first {
			w &= firstMask
		}
		if i == last {
			w &= lastMask
		}
		for w != 0 {
			answer.content = append(answer.content, uint16(i*64+bits.TrailingZeros64(w)))
			w &= w - 1
		}
	}
	return answer
}

func clipRunContainer(rc *runContainer16, lo, hi uint16) container {
	// first is the first interval that ends at or after lo, last the last
	// interval that starts at or before hi.
	first, present, _ := rc.search(int(lo))
	if !present {
		first++
	}
	last, _, _ := rc.search(int(hi))
	if first > last {
		return newArrayContainer()
	}
	iv := make([]interval16, last-first+1)
	copy(iv, rc.iv[first:last+1])
	if iv[0].start < lo {
		iv[0] = newInterval16Range(lo, iv[0].last())
	}
	if n := len(iv) - 1; iv[n].last() > hi {
		iv[n] = newInterval16Range(iv[n].start, hi)
	}
	return newRunContainer16TakeOwnership(iv).toEfficientContainer()
}

// HeapOr computes the union between many bitmaps quickly using a heap.
// It might be faster than calling Or repeatedly.
func HeapOr(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	}
	// TODO:  for better speed, we could do the operation lazily, see Java implementation
	pq := make(priorityQueue, len(bitmaps))
	for i, bm := range bitmaps {
		pq[i] = &item{bm, i}
	}
	heap.Init(&pq)

	for pq.Len() > 1 {
		x1 := heap.Pop(&pq).(*item)
		x2 := heap.Pop(&pq).(*item)
		heap.Push(&pq, &item{Or(x1.value, x2.value), 0})
	}
	return heap.Pop(&pq).(*item).value
}

// HeapXor computes the symmetric difference between many bitmaps quickly (as opposed to calling Xor repeated).
// Internally, this function uses a heap.
// It might be faster than calling Xor repeatedly.
func HeapXor(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	}

	pq := make(priorityQueue, len(bitmaps))
	for i, bm := range bitmaps {
		pq[i] = &item{bm, i}
	}
	heap.Init(&pq)

	for pq.Len() > 1 {
		x1 := heap.Pop(&pq).(*item)
		x2 := heap.Pop(&pq).(*item)
		heap.Push(&pq, &item{Xor(x1.value, x2.value), 0})
	}
	return heap.Pop(&pq).(*item).value
}

// AndAny provides a result equivalent to x1.And(FastOr(bitmaps)).
// It's optimized to minimize allocations. It also might be faster than separate calls.
func (x1 *Bitmap) AndAny(bitmaps ...*Bitmap) {
	if len(bitmaps) == 0 {
		return
	} else if len(bitmaps) == 1 {
		x1.And(bitmaps[0])
		return
	}

	type withPos struct {
		bitmap *roaringArray
		pos    int
		key    uint16
	}
	filters := make([]withPos, 0, len(bitmaps))

	for _, b := range bitmaps {
		if b.highlowcontainer.size() > 0 {
			filters = append(filters, withPos{
				bitmap: &b.highlowcontainer,
				pos:    0,
				key:    b.highlowcontainer.getKeyAtIndex(0),
			})
		}
	}

	basePos := 0
	intersections := 0
	keyContainers := make([]container, 0, len(filters))
	var (
		tmpArray   *arrayContainer
		tmpBitmap  *bitmapContainer
		minNextKey uint16
	)

	for basePos < x1.highlowcontainer.size() && len(filters) > 0 {
		baseKey := x1.highlowcontainer.getKeyAtIndex(basePos)

		// accumulate containers for current key, find next minimal key in filters
		// and exclude filters that do not have related values anymore
		i := 0
		maxPossibleOr := 0
		minNextKey = MaxUint16
		for _, f := range filters {
			if f.key < baseKey {
				f.pos = f.bitmap.advanceUntil(baseKey, f.pos)
				if f.pos == f.bitmap.size() {
					continue
				}
				f.key = f.bitmap.getKeyAtIndex(f.pos)
			}

			if f.key == baseKey {
				cont := f.bitmap.getContainerAtIndex(f.pos)
				keyContainers = append(keyContainers, cont)
				maxPossibleOr += cont.getCardinality()

				f.pos++
				if f.pos == f.bitmap.size() {
					continue
				}
				f.key = f.bitmap.getKeyAtIndex(f.pos)
			}

			minNextKey = minOfUint16(minNextKey, f.key)
			filters[i] = f
			i++
		}
		filters = filters[:i]

		if len(keyContainers) == 0 {
			basePos = x1.highlowcontainer.advanceUntil(minNextKey, basePos)
			continue
		}

		var ored container

		if len(keyContainers) == 1 {
			ored = keyContainers[0]
		} else {
			//TODO: special case for run containers?
			if maxPossibleOr > arrayDefaultMaxSize {
				if tmpBitmap == nil {
					tmpBitmap = newBitmapContainer()
				}
				tmpBitmap.resetTo(keyContainers[0])
				ored = tmpBitmap
			} else {
				if tmpArray == nil {
					tmpArray = newArrayContainerCapacity(maxPossibleOr)
				}
				tmpArray.realloc(maxPossibleOr)
				tmpArray.resetTo(keyContainers[0])
				ored = tmpArray
			}
			for _, c := range keyContainers[1:] {
				ored = ored.ior(c)
			}
		}

		result := x1.highlowcontainer.getWritableContainerAtIndex(basePos).iand(ored)
		if !result.isEmpty() {
			x1.highlowcontainer.replaceKeyAndContainerAtIndex(intersections, baseKey, result, false)
			intersections++
		}

		keyContainers = keyContainers[:0]
		basePos = x1.highlowcontainer.advanceUntil(minNextKey, basePos)
	}

	x1.highlowcontainer.resize(intersections)
}
