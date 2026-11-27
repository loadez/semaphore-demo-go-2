// e2eclient hits a running server and checks responses. Exit non-zero on failure.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func get(url string) (int, string, error) {
	c := http.Client{Timeout: 5 * time.Second}
	r, err := c.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	return r.StatusCode, string(b), nil
}

func main() {
	base := flag.String("base", "http://localhost:8001", "base url")
	n := flag.Int("n", 200, "requests per worker")
	workers := flag.Int("workers", 8, "workers")
	flag.Parse()
	var wg sync.WaitGroup
	var mu sync.Mutex
	failures := 0
	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < *n; i++ {
				code, body, err := get(*base + "/time")
				ok := err == nil && code == 200 && strings.Contains(body, "The current time is")
				code2, body2, err2 := get(fmt.Sprintf("%s/item/%d", *base, i))
				ok = ok && err2 == nil && code2 == 200 && strings.Contains(body2, fmt.Sprintf("/item/%d", i))
				if !ok {
					mu.Lock()
					failures++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	fmt.Printf("requests=%d failures=%d\n", *workers**n*2, failures)
	if failures > 0 {
		os.Exit(1)
	}
}
