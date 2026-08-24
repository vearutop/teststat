#!/bin/bash

# Unlike run-imperfect.sh/run-broken.sh, this one runs go test and teststat itself from the repo
# root, not from here (app/testdata): the whole point of this scenario is the miss reason's file
# path, and running from repo root -- the natural place a user would actually invoke teststat
# from -- gives the clean "cachemiss/marker.txt" rather than a "../../cachemiss/marker.txt"
# artifact of this script's own directory (see inputFilePath/relativeTo in app/testcache.go).
cd "$(dirname "$0")/../.." || exit 1

rm -f app/testdata/test-cachemiss*.jsonl
rm -f app/testdata/cachemiss.md

export GOCACHE=$(mktemp -d)
export GODEBUG=gocachetest=1

# Two rounds against the same GOCACHE: nothing here ever fails, so there's no retry loop -- the
# point is proving the second (warm) round still doesn't show (cached).
go test -tags cachemiss -json ./cachemiss/... > app/testdata/test-cachemiss0.jsonl 2>&1
go test -tags cachemiss -json ./cachemiss/... > app/testdata/test-cachemiss1.jsonl 2>&1

go run . -markdown app/testdata/test-cachemiss*.jsonl > app/testdata/cachemiss.md
