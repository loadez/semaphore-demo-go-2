#!/usr/bin/env bash
# Shard-aware e2e run against the prebuilt binary and postgres.
set -euo pipefail
idx=${SEMAPHORE_JOB_INDEX:-1}
port=$((8100 + idx))
chmod +x bin/main-linux-amd64
psql -h localhost -U postgres -c "CREATE DATABASE s2" 2>/dev/null || true
psql -h localhost -U postgres -d s2 -c "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, first_name TEXT, last_name TEXT)"
psql -h localhost -U postgres -d s2 -c "INSERT INTO users(first_name,last_name) SELECT 'f'||g, 'l'||g FROM generate_series(1,100) g"
./bin/main-linux-amd64 $port > server-$idx.log 2>&1 &
pid=$!
for i in $(seq 1 30); do curl -fs localhost:$port/time >/dev/null && break; sleep 1; done
echo "shard $idx of ${SEMAPHORE_JOB_COUNT:-6} on port $port"
body=$(curl -fs localhost:$port/getdata)
echo "$body" | grep -c "<h3" | tee /dev/stderr | grep -q '^[1-9]'
go run ./cmd/e2eclient -base http://localhost:$port -workers $((idx + 3)) -n $((100 * idx))
kill $pid
