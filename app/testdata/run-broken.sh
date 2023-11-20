#!/bin/bash

rm -f ./test-broken*.jsonl
rm -f broken.md

go test -tags broken -short -race -json ../../broken/... 2>&1 > test-broken0.jsonl
go run ../.. -progress -verbosity 2 -skip-report -failed-tests failed-broken.txt -skip-parent -failed-builds broken-errors.txt test-broken0.jsonl

# Retries.
for i in {1..3}
do
test ! -f failed-broken.txt || (export FAILED=$(cat failed-broken.txt) && rm failed-broken.txt && echo "Retry $i: $FAILED" && go test -v -tags broken -short -race -run $FAILED -json ../../broken/... 2>&1 | go run ../.. -progress -skip-parent -skip-report -failed-tests failed-broken.txt -store test-broken${i}.jsonl - )
done

go run ../.. -failure-stats failure-stats-broken.txt -markdown test-broken*.jsonl > broken.md