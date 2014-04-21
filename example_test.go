package skiplist

import (
	"fmt"
	"math/rand"
)

func ExampleSkiplist_Insert() {
	rand.Seed(1)

	z := New()
	for i := 0; i < 20; i++ {
		z.Insert(Int(i))
	}
	z.PrintDebug()

	// Output:
	// length: 20, level: 5
	// head____________ 6________________________
	// head____________ 6__________12____________
	// head____________ 6__ 8______12____15______
	// head__________ 5 6 7 8____1112____15__17__
	// head 0 1 2 3 4 5 6 7 8 910111213141516171819
}

func ExampleSkiplist_SampleInRange() {
	rand.Seed(1)

	z := New()
	for i := 0; i < 20; i++ {
		z.Insert(Int(i))
	}
	sampled := z.SampleInRange(RangeSpec{Min: Int(4), Max: Int(17)}, 5)
	fmt.Println(sampled)

	// Output:
	// [5 8 10 13 14]
}
