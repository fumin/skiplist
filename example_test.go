package skiplist_test

import (
	"math/rand"

	"github.com/fumin/skiplist"
)

type Int int

func (i Int) Less(j skiplist.Ordered) bool {
	return i < j.(Int)
}

func ExampleSkiplist_Insert() {
	rand.Seed(1)

	z := skiplist.New()
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
