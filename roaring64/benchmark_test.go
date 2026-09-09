package roaring64

import (
	"math"
	"math/rand"
	"testing"
)

// BENCHMARKS, to run them type "go test -bench Benchmark -run -"

// go test -bench BenchmarkIteratorAlloc -benchmem -run -
func BenchmarkIteratorAlloc(b *testing.B) {
	bm := New()
	domain := uint64(100000000)
	count := 10000
	for j := 0; j < count; j++ {
		v := uint64(rand.Int63n(int64(domain)))
		bm.Add(v)
	}
	i := IntIterator64{}
	expectedCardinality := bm.GetCardinality()
	counter := uint64(0)
	b.Run("simple iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			i := bm.Iterator()
			for i.HasNext() {
				i.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("simple iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			i.Initialize(bm)
			for i.HasNext() {
				i.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("values iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			Values(bm)(func(_ uint64) bool {
				counter++
				return true
			})
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("reverse iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			ir := bm.ReverseIterator()
			for ir.HasNext() {
				ir.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	ir := IntReverseIterator64{}

	b.Run("reverse iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			ir.Initialize(bm)
			for ir.HasNext() {
				ir.Next()
				counter++
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	b.Run("backward iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			Backward(bm)(func(_ uint64) bool {
				counter++
				return true
			})
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}

	b.Run("many iteration with alloc", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			buf := make([]uint64, 1024)
			im := bm.ManyIterator()
			for n := im.NextMany(buf); n != 0; n = im.NextMany(buf) {
				counter += uint64(n)
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
	im := ManyIntIterator64{}
	buf := make([]uint64, 1024)

	b.Run("many iteration", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			counter = 0
			im.Initialize(bm)
			for n := im.NextMany(buf); n != 0; n = im.NextMany(buf) {
				counter += uint64(n)
			}
		}
		b.StopTimer()
	})
	if counter != expectedCardinality {
		b.Fatalf("Cardinalities don't match: %d, %d", counter, expectedCardinality)
	}
}

func BenchmarkAndNotArrayRun(b *testing.B) {
	left, right := andNotArrayRunInputs()
	if left.Stats().ArrayContainers != 1 || right.Stats().RunContainers != 1 {
		b.Fatal("benchmark inputs do not use array/run containers")
	}
	if got := AndNot(left, right).GetCardinality(); got != 1024 {
		b.Fatalf("unexpected result cardinality: got %d, want 1024", got)
	}

	b.ResetTimer()
	for b.Loop() {
		AndNot(left, right)
	}
}

func BenchmarkAndNotArrayArray(b *testing.B) {
	left, right := andNotArrayArrayInputs()
	if left.Stats().ArrayContainers != 1 || right.Stats().ArrayContainers != 1 {
		b.Fatal("benchmark inputs do not use array/array containers")
	}
	if got := AndNot(left, right).GetCardinality(); got != 1920 {
		b.Fatalf("unexpected result cardinality: got %d, want 1920", got)
	}

	b.ResetTimer()
	for b.Loop() {
		AndNot(left, right)
	}
}

// rangeBenchmarkInputs builds four bitmaps of 128 containers in each of two
// 32-bit containers, all arrays, all bitmaps or all runs depending on shape.
// The inputs overlap: the array and bitmap shapes draw each container from a
// pool shared by the four bitmaps, and the run shape offsets the same runs
// per bitmap.
func rangeBenchmarkInputs(b *testing.B, shape string) []*Bitmap {
	const numBitmaps = 4
	const numContainers = 128
	rng := rand.New(rand.NewSource(42))
	bms := make([]*Bitmap, numBitmaps)
	for i := range bms {
		bms[i] = NewBitmap()
	}
	for _, high := range []uint64{1, 3} {
		for key := 0; key < numContainers; key++ {
			base := high<<32 + uint64(key)<<16
			switch shape {
			case "array", "bitmap":
				poolSize := 3000
				if shape == "bitmap" {
					poolSize = 30000
				}
				pool := make([]uint64, poolSize)
				for j := range pool {
					pool[j] = base + uint64(rng.Intn(1<<16))
				}
				for _, bm := range bms {
					values := make([]uint64, 0, poolSize)
					for _, v := range pool {
						if rng.Intn(3) != 0 {
							values = append(values, v)
						}
					}
					bm.AddMany(values)
				}
			case "run":
				for i, bm := range bms {
					offset := uint64(rng.Intn(1000))
					for r := 0; r < 8; r++ {
						start := base + offset + uint64(r)*8000 + uint64(i*37)
						end := start + 6000
						if end > base+(1<<16) {
							end = base + (1 << 16)
						}
						if start >= base+(1<<16) {
							break
						}
						bm.AddRange(start, end)
					}
				}
			}
		}
	}
	if shape == "run" {
		for _, bm := range bms {
			bm.RunOptimize()
		}
	}
	for _, bm := range bms {
		stats := bm.Stats()
		var got uint64
		switch shape {
		case "array":
			got = stats.ArrayContainers
		case "bitmap":
			got = stats.BitmapContainers
		case "run":
			got = stats.RunContainers
		}
		if got != stats.Containers {
			b.Fatalf("%s workload produced %d containers of which %d have the expected shape", shape, stats.Containers, got)
		}
	}
	return bms
}

// benchmarkFastRange times a range aggregation against the aggregate-then-clip
// computation it replaces on a window two 16-bit containers wide, and against
// the plain aggregation on the whole value space, where both do the same
// work.
func benchmarkFastRange(b *testing.B, name string, rangeFunc func(uint64, uint64, ...*Bitmap) *Bitmap, aggregate func(...*Bitmap) *Bitmap, viaRangeBitmap func(uint64, uint64, ...*Bitmap) *Bitmap) {
	const start = 3<<32 + uint64(100)<<16 + 1000
	const end = 3<<32 + uint64(102)<<16 + 1000
	for _, shape := range []string{"array", "bitmap", "run"} {
		bms := rangeBenchmarkInputs(b, shape)
		want := aggregate(bms...)
		want.RemoveRange(0, start)
		want.RemoveRange(end, math.MaxUint64)
		if !rangeFunc(start, end, bms...).Equals(want) {
			b.Fatalf("%sRange disagrees with %s on the %s workload", name, name, shape)
		}
		if !viaRangeBitmap(start, end, bms...).Equals(want) {
			b.Fatalf("the range-bitmap form of %s disagrees with %s on the %s workload", name, name, shape)
		}
		b.Run(shape+"/narrow/"+name+"Range", func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				rangeFunc(start, end, bms...)
			}
		})
		b.Run(shape+"/narrow/"+name+" via range bitmap", func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				viaRangeBitmap(start, end, bms...)
			}
		})
		b.Run(shape+"/narrow/"+name+"+RemoveRange", func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				r := aggregate(bms...)
				r.RemoveRange(0, start)
				r.RemoveRange(end, math.MaxUint64)
			}
		})
		b.Run(shape+"/whole/"+name+"Range", func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				rangeFunc(0, math.MaxUint64, bms...)
			}
		})
		b.Run(shape+"/whole/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				aggregate(bms...)
			}
		})
	}
}

// The range-bitmap forms are what a caller can write today to keep the cost
// proportional to the window: every input intersected with a range bitmap
// placed first, whose key merge gallops, and the small results united.
func orViaRangeBitmap(start, end uint64, bitmaps ...*Bitmap) *Bitmap {
	r := NewBitmap()
	r.AddRange(start, end)
	parts := make([]*Bitmap, len(bitmaps))
	for i, bm := range bitmaps {
		parts[i] = And(r, bm)
	}
	return FastOr(parts...)
}

func andViaRangeBitmap(start, end uint64, bitmaps ...*Bitmap) *Bitmap {
	r := NewBitmap()
	r.AddRange(start, end)
	return FastAnd(append([]*Bitmap{r}, bitmaps...)...)
}

// go test -bench 'BenchmarkFast(Or|And)Range' -benchmem -run -
func BenchmarkFastOrRange(b *testing.B) {
	benchmarkFastRange(b, "FastOr", FastOrRange, FastOr, orViaRangeBitmap)
}

func BenchmarkFastAndRange(b *testing.B) {
	benchmarkFastRange(b, "FastAnd", FastAndRange, FastAnd, andViaRangeBitmap)
}
