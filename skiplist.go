// Package skiplist is an implementation of the Skiplist data structure.
// Most of the implementation is copied from Redis https://github.com/antirez/redis.
// In addition, this implementation contains an optimized function SampleInRange
// that randomly samples from a given range. According to the benchmarks,
// SampleInRange is around 12% faster than the naive implementation.
package skiplist

import (
  "fmt"
  "math/rand"
  "sort"
)

// The Ordered interface should be implemented by types that wish to be added
// into skiplists.
type Ordered interface {
  // Less reports whether we are *strictly* less than other.
  Less(other Ordered) bool
}

type skiplistNode struct {
  ordered Ordered
  backward *skiplistNode
  level []skiplistLevel
}

type skiplistLevel struct {
  forward *skiplistNode
  span int
}

type Skiplist struct {
  head, tail *skiplistNode
  length int
  level int
}

// RangeSpec is an interval with information about the inclusiveness of its
// boundaries.
type RangeSpec struct {
  Min, Max Ordered
  Minex, Maxex bool // End points are excluded if Minex or Maxex is true.
}

func (spec *RangeSpec) gteMin(ordered Ordered) bool {
  if spec.Minex {
    return spec.Min.Less(ordered)
  } else {
    return !ordered.Less(spec.Min)
  }
}

func (spec *RangeSpec) lteMax(ordered Ordered) bool {
  if spec.Maxex {
    return ordered.Less(spec.Max)
  } else {
    return !spec.Max.Less(ordered)
  }
}

const (
  MaxLevel = 32
)

// New creates a new skiplist
func New() *Skiplist {
  head := &skiplistNode{
    ordered: nil,
    backward: nil,
    level: make([]skiplistLevel, MaxLevel),
  }
  return &Skiplist{
    head: head,
    tail: nil,
    length: 0,
    level: 1,
  }
}

// Cardinality returns the number of elements in the skiplist
func (z *Skiplist) Cardinality() int {
  return z.length
}

// Add adds an item to a Skiplist
func (z *Skiplist) Add(o Ordered) {
  update := make([]*skiplistNode, MaxLevel)
  rank := make([]int, MaxLevel)
  x := z.head
  for i := z.level-1; i >= 0; i-- {
    if i == z.level-1 {
      rank[i] = 0
    } else {
      rank[i] = rank[i+1]
    }

    for x.level[i].forward != nil && x.level[i].forward.ordered.Less(o) {
      rank[i] += x.level[i].span
      x = x.level[i].forward
    }
    update[i] = x
  }

  level := randLevel()
  if level > z.level {
    for i := z.level; i < level; i++ {
      rank[i] = 0
      update[i] = z.head
      update[i].level[i].span = z.length
    }
    z.level = level
  }
  x = &skiplistNode{ordered: o, level: make([]skiplistLevel, level)}
  for i := 0; i < level; i++ {
    x.level[i].forward = update[i].level[i].forward
    update[i].level[i].forward = x

    // update span covered by update[i] as x is inserted here
    x.level[i].span = update[i].level[i].span - (rank[0]-rank[i])
    update[i].level[i].span = (rank[0]-rank[i]) + 1
  }

  // increment span for untouched levels
  for i := level; i < z.level; i++ {
    update[i].level[i].span++
  }

  if update[0] == z.head {
    x.backward = nil
  } else {
    x.backward = update[0]
  }
  if x.level[0].forward != nil {
    x.level[0].forward.backward = x
  } else {
    z.tail = x
  }
  z.length++
}

// Rem removes an element from the Skiplist.
// If the removal is successful, Rem returns true, otherwise, false.
func (z *Skiplist) Rem(ordered Ordered) bool {
  update := make([]*skiplistNode, MaxLevel)
  x := z.head
  for i := z.level-1; i >= 0; i-- {
    for x.level[i].forward != nil && x.level[i].forward.ordered.Less(ordered) {
      x = x.level[i].forward
    }
    update[i] = x
  }

  x = x.level[0].forward
  if x != nil && x.ordered == ordered {
    for i := 0; i < z.level; i++ {
      if update[i].level[i].forward == x {
        update[i].level[i].span += x.level[i].span - 1
        update[i].level[i].forward = x.level[i].forward
      } else {
        update[i].level[i].span -= 1
      }
    }
    if x.level[0].forward != nil {
      x.level[0].forward.backward = x.backward
    } else {
      z.tail = x.backward
    }
    for z.level > 1 && z.head.level[z.level-1].forward == nil {
      z.level--
    }
    z.length--
    return true
  } else {
    return false
  }
  return false
}

// Range returns elements whose rank is between start and stop.
// Both arguments, start and stop are inclusive, and are 0 based.
func (z *Skiplist) Range(start, stop int) (reply []Ordered) {
  if start > stop || start >= z.length {
    return nil
  }
  if stop >= z.length {
    stop = z.length-1
  }

  ln := z.getElementByRank(start)
  for rangelen := stop-start+1; rangelen > 0; rangelen-- {
    reply = append(reply, ln.ordered)
    ln = ln.level[0].forward
  }

  return
}

// RangeByScore returns elements within the range spec in accending order.
func (z *Skiplist) RangeByScore(spec RangeSpec, offset, limit int) (reply []Ordered) {
  ln := z.firstInRange(spec)
  if ln == nil {
    return
  }

  for ln != nil && offset > 0 {
    offset--
    ln = ln.level[0].forward
  }

  for ln != nil && limit > 0 {
    limit--
    if !spec.lteMax(ln.ordered) {
      break
    }

    reply = append(reply, ln.ordered)

    ln = ln.level[0].forward
  }

  return
}

func (z *Skiplist) SampleInRange_Slow(spec RangeSpec, limit int) (reply []Ordered) {
  rankFirst := z.RankOfFirstInRange(spec)
  rankLast := z.RankOfLastInRange(spec)
  if rankFirst == -1 || rankLast == -1 {
    return nil
  }

  // randomly sample limit number of indices from [rankFirst, rankLast]
  ranks := Sample(limit, rankLast - rankFirst + 1)
  for i := 0; i < len(ranks); i++ {
    ranks[i] += rankFirst
  }

  // get the items from the skiplist
  for i := 0; i < len(ranks); i++ {
    reply = append(reply, z.Range(ranks[i], ranks[i])[0])
  }
  return
}

// SampleInRange is similar to RangeByScore in that it also returns elements
// within the range spec. The difference is however, the elements returned
// by RangeSample are randomly and evenly sampled from the range.
func (z *Skiplist) SampleInRange(spec RangeSpec, limit int) (reply []Ordered) {
  if !z.isInRange(spec) {
    return nil
  }

  // Find the path and rank of the first node in the range spec.
  firstNodePath, firstNodeRanks := z.firstNodeInRange(spec)
  if firstNodePath == nil || firstNodeRanks == nil {
    return nil
  }

  // Find the rank of the last node in the range spec.
  lastNodeRank := z.RankOfLastInRange(spec)
  if lastNodeRank == -1 {
    return nil
  }

  // Randomly sample from [firstNodeRanks[0], lastNodeRank]
  ranks := Sample(limit, lastNodeRank - firstNodeRanks[0] + 1)
  for i := 0; i < len(ranks); i++ {
    ranks[i] += firstNodeRanks[0]
  }
  sort.Ints(ranks)

  // Find the level that we want to traverse, based on the probability that
  // we will get the elements we need after `limit` walks.
  level := 0
  numElem := lastNodeRank - firstNodeRanks[0]
  for numElem > limit && level < z.level-1 {
    numElem /= 2
    level++
  }

  // Find the items whose ranks are the randomly sampled ranks, by walking along
  // the `level` level. On average, we should only need to walk `limit` times.
  // Since for each walk, we are able to get to the item with rank `rank` within
  // `level` steps, the total complexity here is `limit * level`.
  levelNode := firstNodePath[level]
  levelNodeRank := firstNodeRanks[level]
  for _, rank := range ranks {
    if levelNodeRank + levelNode.level[level].span <= rank+1 {
      levelNodeRank += levelNode.level[level].span
      levelNode = levelNode.level[level].forward
    }

    x := levelNode
    traversed := levelNodeRank
    for i := level; i >= 0; i-- {
      for x.level[i].forward != nil && (traversed + x.level[i].span) <= rank+1 {
        traversed += x.level[i].span
        x = x.level[i].forward
      }
      if traversed == rank+1 {
        reply = append(reply, x.ordered)
        break
      }
    }
  }

  return
}

// RankOfFirstInRange returns the rank of the first node in the range spec.
// A rank of -1 means that no node exists in the range spec.
func (z *Skiplist) RankOfFirstInRange(spec RangeSpec) int {
  _, firstNodeRanks := z.firstNodeInRange(spec)
  if firstNodeRanks == nil {
    return -1
  }
  return firstNodeRanks[0]
}

// firstNodeInRange returns information about the first node in the range spec.
// This information includes the nodes and their ranks that were visited
// in order to find the first node.
// The indices of the returned slices represent the level in which these nodes
// are visited. In other words, the first node itself is given by the node
// visited in level 0, i.e. firstNodePath[0].
func (z *Skiplist) firstNodeInRange(spec RangeSpec) ([]*skiplistNode, []int) {
  if !z.isInRange(spec) {
    return nil, nil
  }

  firstNodePath := make([]*skiplistNode, z.level)
  firstNodeRanks := make([]int, z.level)
  x := z.head
  for i := z.level-1; i >= 0; i-- {
    rank := 0
    for x.level[i].forward != nil && !spec.gteMin(x.level[i].forward.ordered) {
      rank += x.level[i].span
      x = x.level[i].forward
    }
    firstNodePath[i] = x
    firstNodeRanks[i] = rank
  }

  if !spec.lteMax(x.level[0].forward.ordered) {
    return nil, nil
  }

  for i := z.level-2; i >= 0; i-- {
    firstNodeRanks[i] += firstNodeRanks[i+1]
  }

  return firstNodePath, firstNodeRanks
}

// RankOfLastInRange returns the rank of the last node in the range spec.
// A rank of -1 means that no node exists in the range spec.
func (z *Skiplist) RankOfLastInRange(spec RangeSpec) int {
  if !z.isInRange(spec) {
    return -1
  }

  lastNodeRank := -1
  x := z.head
  for i := z.level-1; i >= 0; i-- {
    for x.level[i].forward != nil && spec.lteMax(x.level[i].forward.ordered) {
      lastNodeRank += x.level[i].span
      x = x.level[i].forward
    }
  }

  if !spec.gteMin(x.ordered) {
    return -1
  }

  return lastNodeRank
}

func (z *Skiplist) isInRange(spec RangeSpec) bool {
  x := z.tail
  if x == nil || !spec.gteMin(x.ordered) {
    return false
  }

  x = z.head.level[0].forward
  if x == nil || !spec.lteMax(x.ordered) {
    return false
  }

  return true
}

func (z *Skiplist) firstInRange(spec RangeSpec) *skiplistNode {
  if !z.isInRange(spec) {
    return nil
  }

  x := z.head
  for i := z.level-1; i >= 0; i-- {
    for x.level[i].forward != nil && !spec.gteMin(x.level[i].forward.ordered) {
      x = x.level[i].forward
    }
  }

  x = x.level[0].forward

  if !spec.lteMax(x.ordered) {
    return nil
  }
  return x
}

func (z *Skiplist) lastInRange(spec RangeSpec) *skiplistNode {
  if !z.isInRange(spec) {
    return nil
  }

  x := z.head
  for i := z.level-1; i >= 0; i-- {
    for x.level[i].forward != nil && spec.lteMax(x.level[i].forward.ordered) {
      x = x.level[i].forward
    }
  }

  if !spec.gteMin(x.ordered) {
    return nil
  }
  return x
}

func (z *Skiplist) getElementByRank(rank int) *skiplistNode {
  traversed := 0
  x := z.head
  for i := z.level-1; i >= 0; i-- {
    for x.level[i].forward != nil && (traversed + x.level[i].span <= rank+1) {
      traversed += x.level[i].span
      x = x.level[i].forward
    }
    if traversed == rank+1 {
      return x
    }
  }
  return nil
}

func (z *Skiplist) PrintDebug() {
  fmt.Printf("length: %d, level: %d\n", z.length, z.level)

  for i := z.level-1; i >= 0; i-- {
    for node := z.head; node != nil; node = node.level[i].forward {
      fmt.Printf("%2d", node.ordered)
      for j := 1; j < node.level[i].span; j++ {
        fmt.Printf("__")
      }
    }
    fmt.Printf("\n")
  }
}

func randLevel() int {
  level := 1
  for rand.Intn(2) == 0 && level < MaxLevel {
    level++
  }
  return level
}

// Sample randomly selects k integers from the range [0, max).
// Given that Sample's algorithm is to run Fisher-Yates shuffle k times, it's complexity is O(k).
func Sample(k, max int) (sampled []int) {
  if k >= max {
    for i := 0; i < max; i++ {
      sampled = append(sampled, i)
    }
    return
  }

  swapped := make(map[int]int, k)
  for i := 0; i < k; i++ {
    // generate a random number from [i, max)
    r := rand.Intn(max-i) + i

    // swapped[i], swapped[r] = swapped[r], swapped[i]
    vr, ok := swapped[r]
    if ok {
      sampled = append(sampled, vr)
    } else {
      sampled = append(sampled, r)
    }
    vi, ok := swapped[i]
    if ok {
      swapped[r] = vi
    } else {
      swapped[r] = i
    }
  }
  return
}
