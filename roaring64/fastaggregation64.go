package roaring64

import "github.com/RoaringBitmap/roaring/v2"

// FastAnd computes the intersection between many bitmaps quickly
// Compared to the And function, it can take many bitmaps as input, thus saving the trouble
// of manually calling "And" many times.
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
func FastOr(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	} else if len(bitmaps) == 1 {
		return bitmaps[0].Clone()
	}
	answer := Or(bitmaps[0], bitmaps[1])
	for _, bm := range bitmaps[2:] {
		answer.Or(bm)
	}
	return answer
}

// FastAndRange computes the intersection of the given bitmaps restricted to
// the half-open range [rangeStart, rangeEnd). The result equals FastAnd
// followed by the removal of every value outside the range, but only the
// containers whose key lies inside the range are visited, and inside each of
// them only the 32-bit containers the range spans: the cost is proportional
// to the number of containers the range spans times the number of inputs,
// not to the size of the inputs.
//
// When rangeStart >= rangeEnd or no bitmaps are given, the result is empty.
// rangeEnd is exclusive, so the value math.MaxUint64 cannot be selected.
//
// The inputs are never modified, including their copy-on-write bookkeeping
// at both levels. FastAnd with a single input calls Clone, and Clone flags
// every container of a copy-on-write source; this function does not. Each
// container of the result is built by the 32-bit roaring.FastAndRange, which
// shares only the containers already flagged copy-on-write and builds every
// other one freshly. The result has copy-on-write disabled.
//
// Performance hints: as with FastAnd, a tiny bitmap in first position helps,
// since the intersection at each key starts from the first input's container.
func FastAndRange(rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) *Bitmap {
	answer := NewBitmap()
	if rangeStart >= rangeEnd || len(bitmaps) == 0 {
		return answer
	}
	w := newContainerWindow(rangeStart, rangeEnd)
	first := w.cursor(&bitmaps[0].highlowcontainer)
	if len(bitmaps) == 1 {
		for ; !first.exhausted(); first.pos++ {
			lo, hi := w.bounds(first.key())
			if c := roaring.FastAndRange(uint64(lo), uint64(hi)+1, first.container()); !c.IsEmpty() {
				answer.highlowcontainer.appendContainer(first.key(), c, false)
			}
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
			if c := roaring.FastAndRange(uint64(lo), uint64(hi)+1, first.container(), second.container()); !c.IsEmpty() {
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
// key lies inside the range are visited, and inside each of them only the
// 32-bit containers the range spans: the cost is proportional to the number
// of containers the range spans times the number of inputs, not to the size
// of the inputs.
//
// When rangeStart >= rangeEnd or no bitmaps are given, the result is empty.
// rangeEnd is exclusive, so the value math.MaxUint64 cannot be selected.
//
// The inputs are never modified, including their copy-on-write bookkeeping
// at both levels. FastOr with a single input calls Clone, and Clone flags
// every container of a copy-on-write source; this function does not. Each
// container of the result is built by the 32-bit roaring.FastOrRange, which
// shares only the containers already flagged copy-on-write and builds every
// other one freshly. The result has copy-on-write disabled.
func FastOrRange(rangeStart, rangeEnd uint64, bitmaps ...*Bitmap) *Bitmap {
	answer := NewBitmap()
	if rangeStart >= rangeEnd || len(bitmaps) == 0 {
		return answer
	}
	w := newContainerWindow(rangeStart, rangeEnd)
	cursors := make([]rangeCursor, 0, len(bitmaps))
	for _, bm := range bitmaps {
		if c := w.cursor(&bm.highlowcontainer); !c.exhausted() {
			cursors = append(cursors, c)
		}
	}
	gathered := make([]*roaring.Bitmap, 0, len(cursors))
	for len(cursors) > 0 {
		key := cursors[0].key()
		for i := 1; i < len(cursors); i++ {
			if k := cursors[i].key(); k < key {
				key = k
			}
		}
		gathered = gathered[:0]
		for i := range cursors {
			if c := &cursors[i]; c.key() == key {
				gathered = append(gathered, c.container())
			}
		}
		lo, hi := w.bounds(key)
		if c := roaring.FastOrRange(uint64(lo), uint64(hi)+1, gathered...); !c.IsEmpty() {
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
// keys of the first and last containers it touches and the low 32 bits of its
// first and last values.
type containerWindow struct {
	hbStart, lbStart, hbLast, lbLast uint32
}

// newContainerWindow converts the non-empty range [rangeStart, rangeEnd) to
// container coordinates.
func newContainerWindow(rangeStart, rangeEnd uint64) containerWindow {
	return containerWindow{highbits(rangeStart), lowbits(rangeStart), highbits(rangeEnd - 1), lowbits(rangeEnd - 1)}
}

// bounds returns the closed low-bit range [lo, hi] that the window keeps in
// the container with the given key. The key must lie in [hbStart, hbLast].
func (w containerWindow) bounds(key uint32) (lo, hi uint32) {
	lo, hi = 0, maxLowBit
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
func (w containerWindow) cursor(ra *roaringArray64) rangeCursor {
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
	ra       *roaringArray64
	pos, end int
}

func (c *rangeCursor) exhausted() bool {
	return c.pos >= c.end
}

func (c *rangeCursor) key() uint32 {
	return c.ra.keys[c.pos]
}

func (c *rangeCursor) container() *roaring.Bitmap {
	return c.ra.containers[c.pos]
}
