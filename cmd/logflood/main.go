// logflood prints a large volume of log lines: short, long, ANSI colored, JSON.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	lines := flag.Int("lines", 40000, "number of lines")
	flag.Parse()
	w := bufio.NewWriterSize(os.Stdout, 1<<16)
	defer w.Flush()
	long := strings.Repeat("0123456789abcdef", 300) // 4800 chars
	colors := []string{"31", "32", "33", "34", "35", "36"}
	for i := 0; i < *lines; i++ {
		switch i % 5 {
		case 0:
			fmt.Fprintf(w, "line %d plain progress %3d%%\n", i, i%101)
		case 1:
			fmt.Fprintf(w, "\x1b[%sm[%s] step %d ok\x1b[0m\n", colors[i%len(colors)], "INFO", i)
		case 2:
			fmt.Fprintf(w, "{\"ts\":%d,\"level\":\"debug\",\"msg\":\"request handled\",\"i\":%d,\"path\":\"/api/v1/items/%d\"}\n", time.Now().UnixMilli(), i, i%977)
		case 3:
			if i%50 == 3 {
				fmt.Fprintf(w, "LONG %d %s\n", i, long)
			} else {
				fmt.Fprintf(w, "\x1b[1;33mWARN\x1b[0m slow query id=%d took=%dms\n", i, i%400)
			}
		default:
			fmt.Fprintf(w, "\x1b[2K\rspinner %d\n", i)
		}
		if i%5000 == 0 {
			w.Flush()
			time.Sleep(200 * time.Millisecond)
		}
	}
	fmt.Fprintf(w, "DONE %d lines\n", *lines)
}
