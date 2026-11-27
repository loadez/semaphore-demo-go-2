package work

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestFib(t *testing.T) {
	cases := []struct {
		n    int
		want uint64
	}{{0, 0}, {1, 1}, {2, 1}, {10, 55}, {50, 12586269025}, {90, 2880067194370816120}}
	for _, c := range cases {
		t.Run(fmt.Sprint(c.n), func(t *testing.T) {
			if got := Fib(c.n); got != c.want {
				t.Fatalf("Fib(%d)=%d want %d", c.n, got, c.want)
			}
			if c.n <= 25 {
				if got := FibRec(c.n); got != c.want {
					t.Fatalf("FibRec(%d)=%d want %d", c.n, got, c.want)
				}
			}
		})
	}
}

func TestPrimes(t *testing.T) {
	cases := []struct {
		n    int
		want []int
	}{{1, nil}, {2, []int{2}}, {20, []int{2, 3, 5, 7, 11, 13, 17, 19}}}
	for _, c := range cases {
		if got := Primes(c.n); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Primes(%d)=%v want %v", c.n, got, c.want)
		}
	}
	if got := len(Primes(1_000_000)); got != 78498 {
		t.Errorf("pi(1e6)=%d want 78498", got)
	}
}

func TestMul(t *testing.T) {
	id := NewMatrix(3, func(i, j int) float64 {
		if i == j {
			return 1
		}
		return 0
	})
	m := NewMatrix(3, func(i, j int) float64 { return float64(i*3 + j) })
	got, err := Mul(m, id)
	if err != nil || !reflect.DeepEqual(got, m) {
		t.Fatalf("m*I != m: %v %v", got, err)
	}
	if _, err := Mul(m, NewMatrix(2, func(int, int) float64 { return 1 })); err != ErrShape {
		t.Fatalf("want ErrShape, got %v", err)
	}
}

func TestHashChain(t *testing.T) {
	if HashChain("a", 1) == HashChain("b", 1) {
		t.Fatal("different seeds collide")
	}
	first, second := HashChain("a", 1000), HashChain("a", 1000)
	if first != second {
		t.Fatal("not deterministic")
	}
	if HashChain("a", 1) == HashChain("a", 2) {
		t.Fatal("rounds ignored")
	}
}

func TestWordFreqTopN(t *testing.T) {
	f := WordFreq("The cat and the hat. The END, and the cat!")
	want := []string{"the", "and", "cat"}
	if got := TopN(f, 3); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if len(TopN(f, 100)) != len(f) {
		t.Fatal("TopN should cap at len")
	}
}

func TestLRU(t *testing.T) {
	c := NewLRU(2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")
	c.Put("c", 3)
	if _, ok := c.Get("b"); ok {
		t.Fatal("b should be evicted")
	}
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatal("a lost")
	}
	if c.Len() != 2 {
		t.Fatalf("len %d", c.Len())
	}
}

// Exercises the race detector with real concurrent work.
func TestConcurrentHashChains(t *testing.T) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	res := map[int][32]byte{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h := HashChain(fmt.Sprint("seed", i), 200_000)
			mu.Lock()
			res[i] = h
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	if len(res) != 8 {
		t.Fatalf("got %d results", len(res))
	}
}

func BenchmarkFibRec30(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FibRec(30)
	}
}

func BenchmarkPrimes2M(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Primes(2_000_000)
	}
}

func BenchmarkMul200(b *testing.B) {
	m := NewMatrix(200, func(i, j int) float64 { return float64(i ^ j) })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Mul(m, m)
	}
}

func BenchmarkHashChain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashChain("bench", 100_000)
	}
}
