package main

import (
  "github.com/fumin/skiplist"
)

type Int int
func (i Int) Less(o skiplist.Ordered) bool {
  return i < o.(Int)
}

func main() {
  z := skiplist.New()
  for i := 0; i < 20; i++ {
    z.Add(Int(i))
  }
  z.PrintDebug()
}

