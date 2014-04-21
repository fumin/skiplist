package skiplist

import (
	"math/rand"
)

var z *Skiplist = New()

func ExampleSkiplist_Insert() {
	rand.Seed(1)

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
