package skiplist

import (
  "math/rand"
  "sort"
  "testing"
)

type Int int
func (i Int) Less(j Ordered) bool {
  return i < j.(Int)
}

func TestInsertDelete(t *testing.T) {
  N := 1000
  fixture := make([]int, N)
  z := New()
  for i := 0; i < N; i++ {
    fixture[i] = i
    z.Insert(Int(i))
  }
  holes := make(map[int]bool) // holes record the gaps in fixture

  for i := 0; i < N; i++ {
    // Add an element with probablity 0.33, otherwise remove an element
    if rand.Intn(3) == 0 {
      j := 0
      // Randomly insert in the middle of the Skiplist, or at the tail
      if rand.Intn(2) == 0 && len(holes) > 0 {
        j = randomItemFromHoles(t, holes)
      } else {
        j = N + i
      }
      fixture = insertInt(t, fixture, j)
      z.Insert(Int(j))
    } else {
      pos := rand.Intn(len(fixture))
      holes[fixture[pos]] = true
      z.Delete(Int(fixture[pos]))
      fixture = removeAt(t, fixture, pos)
    }

    // Check contents of Skiplist is expected
    start := rand.Intn(len(fixture))
    stop := rand.Intn(len(fixture)-start) + start
    expected := fixture[start:stop+1]

    result := z.Range(start, stop)
    assertEqual(t, expected, result)
  }
}

func TestRange(t *testing.T) {
  N := 100
  fixture, z := genFixtureAndSkiplist(N)

  for i := 0; i < N; i++ {
    start := rand.Intn(100)
    stop := rand.Intn(N-start) + start
    expected := fixture[start:stop+1]

    result := z.Range(start, stop)
    assertEqual(t, expected, result)
  }

  result := z.Range(55, 55)
  if len(result) != 1 || fixture[55] != int(result[0].(Int)) {
    t.Fatalf("expected %d, got %d", fixture[55], int(result[0].(Int)))
  }
}

func TestRangeByScore(t *testing.T) {
  N := 100
  fixture, z := genFixtureAndSkiplist(N)

  // Randomly generate N cases
  specs := make([]RangeSpec, N)
  limits := make([]int, N)
  offsets := make([]int, N)
  for i := 0; i < N; i++ {
    spec := RangeSpec{Min: Int(rand.Intn(N))}
    spec.Max = Int(int(spec.Min.(Int)) + rand.Intn(N-int(spec.Min.(Int))))
    spec.Minex = (rand.Intn(2) == 0)
    spec.Maxex = (rand.Intn(2) == 0)
    specs[i] = spec
    limits[i] = rand.Intn(int(spec.Max.(Int))-int(spec.Min.(Int))+1)
    offsets[i] = rand.Intn(int(spec.Max.(Int))-int(spec.Min.(Int))+1)
  }

  // Test the generated cases
  for i := 0; i < N; i++ {
    spec := specs[i]
    limit := limits[i]
    offset := offsets[i]
    expected := make([]int, 0)

    // Calculate the expected slice for this test case
    for j := 0; j < len(fixture); j++ {
      if (spec.Minex && fixture[j] > int(spec.Min.(Int))) || (!spec.Minex && fixture[j] >= int(spec.Min.(Int))) {
        if (spec.Maxex && fixture[j] < int(spec.Max.(Int))) || (!spec.Maxex && fixture[j] <= int(spec.Max.(Int))) {
          expected = append(expected, fixture[j])
        }
      }
    }
    if offset < len(expected) {
      if offset+limit > len(expected) {
        expected = expected[offset:]
      } else {
        expected = expected[offset:(offset+limit)]
      }
    } else {
      expected = []int{}
    }

    result := z.RangeByScore(spec, offset, limit)
    assertEqual(t, expected, result)
  }
}

func TestSampleInRange(t *testing.T) {
  N := 100
  fixture, z := genFixtureAndSkiplist(N)

  // Randomly generate N cases
  specs := make([]RangeSpec, N)
  limits := make([]int, N)
  for i := 0; i < N; i++ {
    spec := RangeSpec{Min: Int(rand.Intn(N))}
    spec.Max = Int(int(spec.Min.(Int)) + rand.Intn(N-int(spec.Min.(Int))))
    spec.Minex = (rand.Intn(2) == 0)
    spec.Maxex = (rand.Intn(2) == 0)
    specs[i] = spec
    limits[i] = rand.Intn(int(spec.Max.(Int))-int(spec.Min.(Int))+1)
  }

  // Test the generated cases
  for i := 0; i < len(specs); i++ {
    spec := specs[i]
    limit := limits[i]

    minRank, maxRank := -1, -1
    for j := 0; j < len(fixture); j++ {
      if (spec.Minex && fixture[j] > int(spec.Min.(Int))) || (!spec.Minex && fixture[j] >= int(spec.Min.(Int))) {
        minRank = j
        break
      }
    }
    for j := len(fixture)-1; j >= 0; j-- {
      if (spec.Maxex && fixture[j] < int(spec.Max.(Int))) || (!spec.Maxex && fixture[j] <= int(spec.Max.(Int))) {
        maxRank = j
        break
      }
    }
    expected := []int{}
    if minRank != -1 && maxRank != -1 {
      rand.Seed(42)
      ranks := Sample(limit, maxRank+1-minRank)
      for j := 0; j < len(ranks); j++ {
        ranks[j] += minRank
      }
      sort.Ints(ranks)
      expected = make([]int, len(ranks))
      for j := 0; j < len(ranks); j++ {
        expected[j] = fixture[ranks[j]]
      }
    }

    rand.Seed(42)
    result := z.SampleInRange(spec, limit)
    assertEqual(t, expected, result)
  }
}

func genFixtureAndSkiplist(N int) ([]int, *Skiplist) {
  fixture := make([]int, N)
  for i := 0; i < N; i++ {
    fixture[i] = rand.Intn(N)
  }
  sort.Ints(fixture)

  z := New()
  for i := 0; i < len(fixture); i++ {
    z.Insert(Int(fixture[i]))
  }
  return fixture, z
}

func assertEqual(t *testing.T, expected []int, toBeTested []Ordered) {
  if len(expected) != len(toBeTested) {
    t.Fatalf("expected %v, got %v", expected, toBeTested)
    return
  }
  for i := 0; i < len(expected); i++ {
    if expected[i] != int(toBeTested[i].(Int)) {
      t.Fatalf("expected %v, got %v", expected, toBeTested)
    }
  }
}

func randomItemFromHoles(t *testing.T, holes map[int]bool) int {
  keys := []int{}
  for key, _ := range holes {
    keys = append(keys, key)
  }

  i := rand.Intn(len(keys))
  delete(holes, keys[i])
  return keys[i]
}

func insertInt(t *testing.T, s []int, item int) []int {
  res := []int{}
  i := 0
  for i = 0; i < len(s); i++ {
    if s[i] >= item {
      break
    }
    res = append(res, s[i])
  }
  res = append(res, item)
  for ; i < len(s); i++ {
    res = append(res, s[i])
  }
  return res
}

func removeAt(t *testing.T, s []int, pos int) []int {
  res := []int{}
  for i := 0; i < len(s); i++ {
    if i != pos {
      res = append(res, s[i])
    }
  }
  return res
}
