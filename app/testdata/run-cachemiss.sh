#!/bin/bash

rm -f ./test-cachemiss*.jsonl
rm -f cachemiss.md

export GOCACHE=$(mktemp -d)
export GODEBUG=gocachetest=1

# Two rounds against the same GOCACHE: nothing here ever fails, so there's no retry loop -- the
# point is proving the second (warm) round still doesn't show (cached).
go test -tags cachemiss -json ../../cachemiss/... > test-cachemiss0.jsonl 2>&1
go test -tags cachemiss -json ../../cachemiss/... > test-cachemiss1.jsonl 2>&1

go run ../.. -markdown test-cachemiss*.jsonl > cachemiss.md
