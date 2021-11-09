# teststat

[![Build Status](https://github.com/vearutop/teststat/workflows/test-unit/badge.svg)](https://github.com/vearutop/teststat/actions?query=branch%3Amaster+workflow%3Atest-unit)
[![Coverage Status](https://codecov.io/gh/vearutop/teststat/branch/master/graph/badge.svg)](https://codecov.io/gh/vearutop/teststat)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/vearutop/teststat)
[![Time Tracker](https://wakatime.com/badge/github/vearutop/teststat.svg)](https://wakatime.com/badge/github/vearutop/teststat)
![Code lines](https://sloc.xyz/github/vearutop/teststat/?category=code)
![Comments](https://sloc.xyz/github/vearutop/teststat/?category=comments)


A tool to aggregate and mine data from JSON reports of Go tests.

## Usage

```
Usage: teststat [options] report.jsonl ...
        Use `-` as file name to read from STDIN.
  -buckets int
        number of buckets for histogram (default 10)
  -markdown
        render output as markdown
  -race-depth int
        stacktrace depth to group similar data races (default 5)
  -slow duration
        minimal duration of slow test (default 1s)
  -slowest int
        limit number of slowest tests to list (default 30)
  -version
        show version and exit

```

## Examples

### Read from multiple files

Once you've collected JSONL test report, you can analyze it with this tool.
```
teststat -race-depth 4 -buckets 15 -slowest 7 ./flaky.jsonl ./test.jsonl 
```

<details>
<summary>Sample report.</summary>

```
Flaky tests:
github.com/acme/foo/core/affiliate/networks.TestBarSuite/TestOisGetReinvented: 2 passed, 8 failed
github.com/acme/foo/core/affiliate/networks.TestBarSuite/TestOisGetReinstallCallbacks: 2 passed, 8 failed
github.com/acme/foo/core/affiliate/networks.TestBarSuite: 2 passed, 8 failed
github.com/acme/foo/core/kafka.TestClose_Graceful_Pooled: 15 passed, 1 failed
github.com/acme/foo/core/kafka.TestClose_ClosePause: 14 passed, 2 failed

Slowest tests:
pass github.com/acme/foo/manipulation_services/api_server TestCreateLeafTracer_Ok 1m26.4s
pass github.com/acme/foo/manipulation_services/api_server TestCreateTracer_Ok 1m16.55s
pass github.com/acme/foo/manipulation_services/api_server TestCreateTracer_Ok/D4 1m16.45s
pass github.com/acme/foo/manipulation_services/api_server TestCreateLeafTracer_Ok 1m3.28s
pass github.com/acme/foo/manipulation_services/refresh_worker TestConsumeImpression_Success 52.85s
pass github.com/acme/foo/manipulation_services/api_server TestCreateLeafTracer_Ok 31.58s
pass github.com/acme/foo/manipulation_services/refresh_worker TestSubscriptionConsumer_DifferentEventSubtypes 30.39s

Events: map[cont:2368 fail:196 flaky:32 output:1805716 pass:660182 pause:2336 run:780596 skip:120154 slow:863]
Elapsed: 1h36m1.129999952s 
Slow: 40m34.649999952s

Elapsed distribution (seconds):
[  min   max]   cnt total% (37862 events)
[ 0.01  0.10] 32284 85.27% .....................................................................................
[ 0.11  0.24]  3383  8.94% ........
[ 0.25  0.52]   814  2.15% ..
[ 0.53  1.05]   574  1.52% .
[ 1.06  2.03]   552  1.46% .
[ 2.04  3.21]   122  0.32%
[ 3.30  4.90]    37  0.10%
[ 4.99  6.22]    36  0.10%
[ 6.40  8.68]    27  0.07%
[ 8.69 11.41]    22  0.06%
[12.48 14.30]     3  0.01%
[17.92 17.92]     1  0.00%
[30.39 31.58]     2  0.01%
[52.85 63.28]     2  0.01%
[76.45 86.40]     3  0.01%
```

</details>


### Read from STDIN

```
go test -count 5 -json -race ./... | teststat -
```