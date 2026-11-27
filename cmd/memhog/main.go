// memhog allocates and touches memory in steps, verifying the contents.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func memTotalMB() int {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 4096
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "MemTotal:") {
			kb, _ := strconv.Atoi(strings.Fields(s.Text())[1])
			return kb / 1024
		}
	}
	return 4096
}

func main() {
	target := flag.Int("mb", 2500, "target MB")
	step := flag.Int("step", 250, "step MB")
	hold := flag.Duration("hold", 20*time.Second, "hold at peak")
	flag.Parse()
	if max := memTotalMB() * 6 / 10; *target > max {
		fmt.Printf("capping target %d -> %d MB (60%% of MemTotal)\n", *target, max)
		*target = max
	}
	var chunks [][]byte
	for have := 0; have < *target; have += *step {
		b := make([]byte, *step<<20)
		for i := 0; i < len(b); i += 4096 {
			b[i] = byte(i >> 12)
		}
		chunks = append(chunks, b)
		fmt.Printf("allocated %d MB\n", have+*step)
		time.Sleep(2 * time.Second)
	}
	time.Sleep(*hold)
	bad := 0
	for _, b := range chunks {
		for i := 0; i < len(b); i += 4096 {
			if b[i] != byte(i>>12) {
				bad++
			}
		}
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("verified, corrupt pages=%d, heap sys=%d MB\n", bad, ms.HeapSys>>20)
	if bad > 0 {
		os.Exit(1)
	}
	chunks = nil
	runtime.GC()
	time.Sleep(5 * time.Second)
	fmt.Println("released")
}
