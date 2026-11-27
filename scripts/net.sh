#!/usr/bin/env bash
set -euo pipefail
case "${SEMAPHORE_JOB_INDEX:-1}" in
  1) f=go1.22.5.linux-amd64.tar.gz ;;
  2) f=go1.21.13.linux-amd64.tar.gz ;;
  *) f=go1.22.5.linux-arm64.tar.gz ;;
esac
mkdir -p /tmp/net && cd /tmp/net
for round in 1 2; do
  echo "--- round $round: $f"
  retry --times 3 --sleep 3 curl -fsSL -o "$f" "https://dl.google.com/go/$f"
  want=$(curl -fsSL "https://dl.google.com/go/$f.sha256")
  got=$(sha256sum "$f" | cut -d' ' -f1)
  echo "want=$want got=$got size=$(stat -c%s "$f")"
  [ "$want" = "$got" ]
  tar tzf "$f" | wc -l
  rm -f "$f"
done
echo "--- parallel ranged downloads"
for i in 0 1 2 3; do
  curl -fsSL -r $((i*20000000))-$((i*20000000+19999999)) -o part$i "https://dl.google.com/go/go1.22.5.linux-amd64.tar.gz" &
done
wait
ls -l part*
