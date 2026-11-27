#!/usr/bin/env bash
set -euo pipefail
DIR=${1:-/tmp/diskjob}
mkdir -p "$DIR"
avail=$(df -m "$DIR" | awk 'NR==2{print $4}')
mb=3072
if [ "$avail" -lt $((mb * 2)) ]; then mb=$((avail / 3)); fi
echo "disk avail=${avail}MB using ${mb}MB"
echo "--- sequential write (direct)"
dd if=/dev/zero of="$DIR/big1" bs=1M count="$mb" oflag=direct conv=fsync status=progress 2>&1 | tail -3
echo "--- sequential read + checksum"
sync; echo 3 | sudo tee /proc/sys/vm/drop_caches >/dev/null || true
dd if="$DIR/big1" of=/dev/null bs=1M iflag=direct status=progress 2>&1 | tail -3
sha256sum "$DIR/big1"
echo "--- random data file (cpu + disk)"
head -c 1G /dev/urandom > "$DIR/rand1"
cp "$DIR/rand1" "$DIR/rand2"
cmp "$DIR/rand1" "$DIR/rand2" && echo "copy identical"
echo "--- many small files"
mkdir -p "$DIR/small"
for i in $(seq 1 20000); do head -c 4096 /dev/zero > "$DIR/small/f$i"; done
find "$DIR/small" -type f | wc -l
tar czf "$DIR/small.tgz" -C "$DIR" small
ls -l "$DIR/small.tgz"
rm -rf "$DIR/small"
tar xzf "$DIR/small.tgz" -C "$DIR"
find "$DIR/small" -type f | wc -l
if command -v fio >/dev/null; then
  echo "--- fio random rw"
  fio --name=rw --directory="$DIR" --size=512M --rw=randrw --bs=4k --iodepth=16 --ioengine=libaio --direct=1 --runtime=30 --time_based --group_reporting | tail -20
fi
df -h "$DIR"
rm -rf "$DIR"
