package skiplist

import (
	"math/rand"
	"testing"
)

func BenchmarkSampleInRange(b *testing.B) {
	N := 1000000
	specCnt := 100
	rand.Seed(42)
	z, specs, limits := prepareSampleInRangeFixture(N, specCnt)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {

		caseIdx := rand.Intn(specCnt)
		z.SampleInRange(specs[caseIdx], limits[caseIdx])

	}
}

func BenchmarkSampleInRange_Slow(b *testing.B) {
	N := 1000000
	specCnt := 100
	rand.Seed(42)
	z, specs, limits := prepareSampleInRangeFixture(N, specCnt)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {

		caseIdx := rand.Intn(specCnt)
		z.SampleInRange_Slow(specs[caseIdx], limits[caseIdx])

	}
}

func prepareSampleInRangeFixture(N, specCnt int) (*Skiplist, []RangeSpec, []int) {
	z := New()
	for i := 0; i < N; i++ {
		z.Insert(Int(rand.Intn(N)))
	}

	specs := make([]RangeSpec, specCnt)
	limits := make([]int, specCnt)
	for i := 0; i < specCnt; i++ {
		spec := RangeSpec{Min: Int(rand.Intn(N))}
		//spec.Max = Int(int(spec.Min.(Int)) + rand.Intn(N-int(spec.Min.(Int))))
		spec.Max = Int(rand.Intn(1000) + int(spec.Min.(Int)))
		spec.Minex = (rand.Intn(2) == 0)
		spec.Maxex = (rand.Intn(2) == 0)
		specs[i] = spec
		limits[i] = rand.Intn(30) + 10
	}

	return z, specs, limits
}
