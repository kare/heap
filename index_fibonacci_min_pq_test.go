package heap

import "testing"

func TestInsert(t *testing.T) {
	pq, err := NewIndexFibonacciMinPQ(10)
	if err != nil {
		t.Fatal(err)
	}
	err = pq.Insert(2, 1)
	if err != nil {
		t.Fatalf("insert returned unexpected error: %v", err)
	}
	err = pq.Insert(1, 2)
	if err != nil {
		t.Fatalf("insert returned unexpected error: %v", err)
	}
	if pq.Len() != 2 {
		t.Fatalf("expected pq length 2, but got %d", pq.Len())
	}
	if pq.IsEmpty() {
		t.Fatal("expected non empty queue")
	}
}

func TestInsertAndDelMin(t *testing.T) {
	pq, err := NewIndexFibonacciMinPQ(10)
	if err != nil {
		t.Fatal(err)
	}
	testData := []struct {
		i int
		k float32
	}{
		{1, 0.1},
		{4, 0.4},
		{9, 0.9},
		{2, 0.2},
		{3, 0.3},
		{5, 0.5},
		{7, 0.7},
		{8, 0.8},
		{6, 0.6},
	}
	for _, testCase := range testData {
		if err := pq.Insert(testCase.i, testCase.k); err != nil {
			t.Fatal(err)
		}
	}
	if pq.Len() != 9 {
		t.Fatalf("expected pq length 9, but got %d", pq.Len())
	}

	expectedDel := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for n := 0; !pq.IsEmpty(); n++ {
		i, err := pq.DelMin()
		if err != nil {
			t.Fatalf("delete minimun failed: %v", err)
		}
		if expectedDel[n] != i {
			t.Fatalf("expected %d, but got %d", expectedDel[i], i)
		}
	}
}

func TestDecreaseKey(t *testing.T) {
	pq, err := NewIndexFibonacciMinPQ(10)
	if err != nil {
		t.Fatal(err)
	}
	testData := []struct {
		i int
		k float32
	}{
		{1, 0.1},
		{4, 0.4},
		{9, 0.9},
		{2, 0.2},
		{3, 0.3},
		{5, 0.5},
		{7, 0.7},
		{8, 0.8},
		{6, 0.6},
	}
	for _, testCase := range testData {
		if err = pq.Insert(testCase.i, testCase.k); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < len(testData); i++ {
		err = pq.DecreaseKey(i+1, 0.01)
		if err != nil {
			t.Fatal(err)
		}
		if key, err := pq.DelMin(); key != i+1 {
			t.Fatalf("expected key %d, but got %d", i+1, key)
			if err != nil {
				t.Fatalf("delete minumum returned an error: %v", err)
			}
		}
	}
}
