#!/bin/bash

rm -f ./test-imperfect*.jsonl

go test -tags imperfect -short -race -json ../../imperfect/... 2>&1 > test-imperfect0.jsonl
go run ../.. -progress -verbosity 2 -skip-report -failed-tests failed-imperfect.txt -skip-parent -failed-builds errors.txt test-imperfect0.jsonl

# Retries.
for i in {1..20}
do
test ! -f failed-imperfect.txt || (export FAILED=$(cat failed-imperfect.txt) && rm failed-imperfect.txt && echo "Retry $i: $FAILED" && go test -v -tags imperfect -short -race -run $FAILED -json ../../imperfect/... 2>&1 | go run ../.. -progress -skip-parent -skip-report -failed-tests failed-imperfect.txt -store test-imperfect${i}.jsonl - )
done

go run ../.. -failure-stats failure-stats-imperfect.txt -markdown test-imperfect*.jsonl > imperfect.md