package main

import (
  "github.com/fumin/sortedset"
)

type Int int

// sortedset.Ordered interface
func (i Int) Less(o sortedset.Ordered) bool {
  return i < o.(Int)
}

func main() {
  z := sortedset.New()
  for i := 0; i < 20; i++ {
    z.Add(Int(i))
  }
  z.PrintDebug()
}

