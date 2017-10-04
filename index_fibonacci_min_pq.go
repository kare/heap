package heap // import "kkn.fi/heap"

import (
	"errors"
	"fmt"
)

// IndexFibonacciMinPQ struct represents an indexed priority queue of float32 keys.
// It supports the usual insert and delete-the-minimum operations,
// along with delete and change-the-key methods.
// In order to let the client refer to keys on the priority queue,
// an integer between 0 and N-1 is associated with each key ; the client
// uses this integer to specify which key to delete or change.
// It also supports methods for peeking at the minimum key,
// testing if the priority queue is empty, and iterating through
// the keys.
//
// This implementation uses a Fibonacci heap along with an array to associate
// keys with integers in the given range.
// The Insert, Len, IsEmpty, Contains, MinIndex, MinKey
// and KeyOf take constant time.
// The DecreaseKey operation takes amortized constant time.
// The Delete, IncreaseKey, DelMin, ChangeKey take amortized logarithmic time.
// Construction takes time proportional to the specified capacity
type IndexFibonacciMinPQ struct {
	nodes  []*node       // Array of Nodes in the heap
	head   *node         // Head of the circular root list
	min    *node         // Minimum Node in the heap
	length int           // Number of keys in the heap
	max    int           // Maximum number of elements in the heap
	table  map[int]*node // Used for the consolidate operation
}

// node represents a node of a tree.
type node struct {
	key           float32 // Key of the Node
	order         int     // The order of the tree rooted by this Node
	index         int     // Index associated with the key
	prev, next    *node   // siblings of the Node
	parent, child *node   // parent and child of this Node
	mark          bool    // Indicates if this Node already lost a child
}

func (pq IndexFibonacciMinPQ) String() string {
	return fmt.Sprintf("pq{nodes=%v,head=%v}", pq.nodes, pq.head)
}

func (n node) String() string {
	return fmt.Sprintf("node{key=%.1f,order=%d,index=%d}", n.key, n.order, n.index)
}

// NewIndexFibonacciMinPQ initializes an empty indexed priority queue with indices between 0 and given max-1.
// Worst case is O(n).
func NewIndexFibonacciMinPQ(max int) (*IndexFibonacciMinPQ, error) {
	if max < 0 {
		return nil, errors.New("cannot create a priority queue of negative size")
	}
	pq := &IndexFibonacciMinPQ{
		max:   max,
		nodes: make([]*node, max),
	}
	return pq, nil
}

// IsEmpty returns true if the priority queue is empty, false if not.
// Worst case is O(1).
func (pq IndexFibonacciMinPQ) IsEmpty() bool {
	return pq.length == 0
}

// Contains returns true if i is on the priority queue, false if not.
// Worst case is O(1).
func (pq IndexFibonacciMinPQ) Contains(i int) bool {
	if i < 0 || i >= pq.max {
		return false
	}
	return pq.nodes[i] != nil
}

// Len returns the number of elements currently on the priority queue.
// Worst case is O(1).
func (pq IndexFibonacciMinPQ) Len() int {
	return pq.length
}

// Insert associates a key with an index.
// Worst case is O(1).
func (pq *IndexFibonacciMinPQ) Insert(i int, key float32) error {
	if i < 0 || i >= pq.max {
		return errors.New("illegal argument")
	}
	if pq.Contains(i) {
		return errors.New("specified index is already in the queue")
	}
	x := &node{
		key:   key,
		index: i,
	}
	pq.nodes[i] = x
	pq.length++
	pq.head = pq.insertNode(x, pq.head)
	if pq.min == nil {
		pq.min = pq.head
	} else {
		if greater(pq.min.key, key) {
			pq.min = pq.head
		}
	}
	return nil
}

// MinIndex returns the index associated with the minimum key.
// Worst case is O(1).
func (pq IndexFibonacciMinPQ) MinIndex() (int, error) {
	if pq.IsEmpty() {
		return 0, errors.New("priority queue is empty")
	}
	return pq.min.index, nil
}

// MinKey gets the minimum key currently in the queue.
// Worst case is O(1).
func (pq IndexFibonacciMinPQ) MinKey() (float32, error) {
	if pq.IsEmpty() {
		return 0, errors.New("priority queue is empty")
	}
	return pq.min.key, nil
}

// DelMin deletes minimum key.
// Worst case is O(log(n)) (amortized).
func (pq *IndexFibonacciMinPQ) DelMin() (int, error) {
	if pq.IsEmpty() {
		return 0, errors.New("priority queue is empty")
	}
	pq.head = pq.cutNode(pq.min, pq.head)
	x := pq.min.child
	index := pq.min.index
	if x != nil {
		for ok := true; ok; ok = (*x != *pq.min.child) {
			x.parent = nil
			x = x.next
		}
		pq.head = pq.meld(pq.head, x)
		pq.min.child = nil // For garbage collection
	}
	pq.length--
	if !pq.IsEmpty() {
		pq.consolidate()
	} else {
		pq.min = nil
	}
	pq.nodes[index] = nil
	return index, nil
}

// KeyOf returns the key associated with index i.
// Worst case is O(1).
func (pq IndexFibonacciMinPQ) KeyOf(i int) (float32, error) {
	if i < 0 || i >= pq.max {
		return 0, errors.New("illegal argument")
	}
	if !pq.Contains(i) {
		return 0, errors.New("specified index is not in the queue")
	}
	return pq.nodes[i].key, nil
}

// ChangeKey changes the key associated with index i to the given key.
// If the given key is greater, worst case is O(log(n)).
// If the given key is lower, worst case is O(1) (amortized).
func (pq *IndexFibonacciMinPQ) ChangeKey(i int, key float32) error {
	if i < 0 || i >= pq.max {
		return errors.New("illegal argument")
	}
	if !pq.Contains(i) {
		return errors.New("specified index is not in the queue")
	}
	if greater(key, pq.nodes[i].key) {
		if err := pq.IncreaseKey(i, key); err != nil {
			return err
		}
	} else {
		if err := pq.DecreaseKey(i, key); err != nil {
			return err
		}
	}
	return nil
}

// DecreaseKey decreases the key associated with index i to the given key.
// Worst case is O(1) (amortized).
func (pq *IndexFibonacciMinPQ) DecreaseKey(i int, key float32) error {
	if i < 0 || i >= pq.max {
		return errors.New("illegal argument")
	}
	if !pq.Contains(i) {
		return errors.New("specified index is not in the queue")
	}
	if greater(key, pq.nodes[i].key) {
		return errors.New("calling with this argument would not decrease the key")
	}
	x := pq.nodes[i]
	x.key = key
	if greater(pq.min.key, key) {
		pq.min = x
	}
	if x.parent != nil && greater(x.parent.key, key) {
		pq.cut(i)
	}
	return nil
}

// IncreaseKey increases the key associated with index i to the given key
// Worst case is O(log(n))
func (pq *IndexFibonacciMinPQ) IncreaseKey(i int, key float32) error {
	if i < 0 || i >= pq.max {
		return errors.New("illegal argument")
	}
	if !pq.Contains(i) {
		return errors.New("specified index is not in the queue")
	}
	if greater(pq.nodes[i].key, key) {
		return errors.New("calling with this argument would not increase the key")
	}
	if err := pq.Delete(i); err != nil {
		return err
	}
	if err := pq.Insert(i, key); err != nil {
		return err
	}
	return nil
}

// Delete deletes the key associated the given index.
// Worst case is O(log(n)) (amortized).
func (pq *IndexFibonacciMinPQ) Delete(i int) error {
	if i < 0 || i >= pq.max {
		return errors.New("illegal argument")
	}
	if !pq.Contains(i) {
		return errors.New("specified index is not in the queue")
	}
	x := pq.nodes[i]
	if x.parent != nil {
		pq.cut(i)
	}
	pq.head = pq.cutNode(x, pq.head)
	if x.child != nil {
		child := x.child
		x.child = nil // For garbage collection
		x = child
		for ok := true; ok; ok = (*child != *x) {
			child.parent = nil
			child = child.next
		}
		pq.head = pq.meld(pq.head, child)
	}
	if !pq.IsEmpty() {
		pq.consolidate()
	} else {
		pq.min = nil
	}
	pq.nodes[i] = nil
	pq.length--
	return nil
}

// greater compares two keys
func greater(n float32, m float32) bool {
	return n > m
}

// link links a new root key. Assuming root1 holds a greater key than root2, root2 becomes the new root
func (pq *IndexFibonacciMinPQ) link(root1, root2 *node) {
	root1.parent = root2
	root2.child = pq.insertNode(root1, root2.child)
	root2.order++
}

// cut removes a Node from its parent's child list and insert it in the root list.
// If the parent Node already lost a child, reshapes the heap accordingly.
func (pq *IndexFibonacciMinPQ) cut(i int) {
	x := pq.nodes[i]
	parent := x.parent
	parent.child = pq.cutNode(x, parent.child)
	x.parent = nil
	parent.order--
	pq.head = pq.insertNode(x, pq.head)
	parent.mark = !parent.mark
	if !parent.mark && parent.parent != nil {
		pq.cut(parent.index)
	}
}

// consolidate coalesces the roots, thus reshapes the heap.
func (pq *IndexFibonacciMinPQ) consolidate() {
	//TODO: Caching a map greatly improves performances
	//TODO: Check for dangling memory references!!!
	pq.table = make(map[int]*node)
	x := pq.head
	maxOrder := 0
	pq.min = pq.head
	var y, z *node
	for ok := true; ok; ok = (*x != *pq.head) {
		y = x
		x = x.next
		z = pq.table[y.order]
		for z != nil {
			delete(pq.table, y.order)
			if greater(y.key, z.key) {
				pq.link(y, z)
				y = z
			} else {
				pq.link(z, y)
			}
			z = pq.table[y.order]
		}
		pq.table[y.order] = y
		if y.order > maxOrder {
			maxOrder = y.order
		}
	}
	pq.head = nil
	for _, n := range pq.table {
		if greater(pq.min.key, n.key) {
			pq.min = n
		}
		pq.head = pq.insertNode(n, pq.head)
	}
}

// insertNode inserts a Node in a circular list containing head, returns a new head.
func (pq *IndexFibonacciMinPQ) insertNode(x, head *node) *node {
	if head == nil {
		x.prev = x
		x.next = x
	} else {
		head.prev.next = x
		x.next = head
		x.prev = head.prev
		head.prev = x
	}
	return x
}

// cutNode removes a tree from the list defined by the head pointer.
func (pq *IndexFibonacciMinPQ) cutNode(x, head *node) *node {
	if *x.next == *x {
		x.next = nil
		x.prev = nil
		return nil
	}
	x.next.prev = x.prev
	x.prev.next = x.next
	res := x.next
	x.next = nil
	x.prev = nil
	if *head == *x {
		return res
	}
	return head
}

// meld merges two lists together.
func (pq *IndexFibonacciMinPQ) meld(x, y *node) *node {
	if x == nil {
		return y
	}
	if y == nil {
		return x
	}
	x.prev.next = y.next
	y.next.prev = x.prev
	x.prev = y
	y.next = x
	return x
}

// Slice returns a slice over the indexes in the priority queue in ascending order.
// Returns an empty slice on error.
// Worst case is O(n).
func (pq IndexFibonacciMinPQ) Slice() []int {
	result := make([]int, 0, pq.max)
	for _, n := range pq.nodes {
		if n != nil {
			result = append(result, n.index)
		}
	}
	return result
}
