package skiplist

func ExampleAdd() {
  z := New()
  for i := 0; i < 20; i++ {
    z.Insert(Int(i))
  }
  z.PrintDebug()
}
