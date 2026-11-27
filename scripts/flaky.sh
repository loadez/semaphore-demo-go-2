#!/usr/bin/env bash
# Fails the first two attempts, passes the third. Meant to be wrapped in retry.
f=/tmp/flaky.count
n=$(cat "$f" 2>/dev/null || echo 0); n=$((n+1)); echo $n > "$f"
echo "flaky attempt $n"
[ "$n" -ge 3 ]
