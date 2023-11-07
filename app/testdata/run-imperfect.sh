#!/bin/bash

go test -tags imperfect -short -coverprofile=unit.coverprofile -covermode=atomic -race -json ./... 2>&1 > test-report0.jsonl
go run . -progress -verbosity 2 -skip-report -failed-tests failed.txt -skip-parent -failed-builds errors.txt test-report0.jsonl

# Retries.
for i in {1..20}
do
test ! -f failed.txt || (export FAILED=$(cat failed.txt) && rm failed.txt && echo "Retry $i: $FAILED" && go test -tags imperfect -short -race -run $FAILED -json ./... 2>&1 > test-report${i}.jsonl && go run . -progress -skip-report -failed-tests failed.txt test-report${i}.jsonl)
done

go run . -failure-stats failure-stats.txt -markdown test-report*.jsonl