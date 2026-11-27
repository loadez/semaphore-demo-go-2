// Package work holds small CPU-bound building blocks used by tests and benchmarks.
package work

import (
	"crypto/sha256"
	"errors"
	"sort"
	"strings"
)

// Fib returns the n-th Fibonacci number iteratively.
func Fib(n int) uint64 {
	if n < 2 {
		return uint64(n)
	}
	var a, b uint64 = 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// FibRec is the deliberately slow recursive version.
func FibRec(n int) uint64 {
	if n < 2 {
		return uint64(n)
	}
	return FibRec(n-1) + FibRec(n-2)
}

// Primes returns all primes <= n using a sieve.
func Primes(n int) []int {
	if n < 2 {
		return nil
	}
	composite := make([]bool, n+1)
	var out []int
	for i := 2; i <= n; i++ {
		if composite[i] {
			continue
		}
		out = append(out, i)
		for j := i * i; j <= n; j += i {
			composite[j] = true
		}
	}
	return out
}

// Matrix is a dense square matrix.
type Matrix [][]float64

// NewMatrix builds an n x n matrix filled by f(i, j).
func NewMatrix(n int, f func(i, j int) float64) Matrix {
	m := make(Matrix, n)
	for i := range m {
		m[i] = make([]float64, n)
		for j := range m[i] {
			m[i][j] = f(i, j)
		}
	}
	return m
}

var ErrShape = errors.New("work: incompatible matrix shapes")

// Mul multiplies two square matrices.
func Mul(a, b Matrix) (Matrix, error) {
	if len(a) != len(b) {
		return nil, ErrShape
	}
	n := len(a)
	out := NewMatrix(n, func(int, int) float64 { return 0 })
	for i := 0; i < n; i++ {
		for k := 0; k < n; k++ {
			aik := a[i][k]
			for j := 0; j < n; j++ {
				out[i][j] += aik * b[k][j]
			}
		}
	}
	return out, nil
}

// HashChain applies sha256 rounds times, starting from seed.
func HashChain(seed string, rounds int) [32]byte {
	h := sha256.Sum256([]byte(seed))
	for i := 1; i < rounds; i++ {
		h = sha256.Sum256(h[:])
	}
	return h
}

// WordFreq counts lower-cased words in text.
func WordFreq(text string) map[string]int {
	out := map[string]int{}
	for _, w := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		out[w]++
	}
	return out
}

// TopN returns the n most frequent words, ties broken alphabetically.
func TopN(freq map[string]int, n int) []string {
	words := make([]string, 0, len(freq))
	for w := range freq {
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		if freq[words[i]] != freq[words[j]] {
			return freq[words[i]] > freq[words[j]]
		}
		return words[i] < words[j]
	})
	if n < len(words) {
		words = words[:n]
	}
	return words
}

// LRU is a tiny fixed-size LRU cache (not goroutine safe).
type LRU struct {
	cap   int
	order []string
	data  map[string]int
}

func NewLRU(capacity int) *LRU { return &LRU{cap: capacity, data: map[string]int{}} }

func (c *LRU) touch(k string) {
	for i, o := range c.order {
		if o == k {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	c.order = append(c.order, k)
}

func (c *LRU) Get(k string) (int, bool) {
	v, ok := c.data[k]
	if ok {
		c.touch(k)
	}
	return v, ok
}

func (c *LRU) Put(k string, v int) {
	if _, ok := c.data[k]; !ok && len(c.data) >= c.cap {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.data, oldest)
	}
	c.data[k] = v
	c.touch(k)
}

func (c *LRU) Len() int { return len(c.data) }
