### Failures
<details>
<summary>Failed tests (including flaky): 2</summary>

<details>
<summary><code>github.com/vearutop/teststat/imperfect.TestThatFlakes</code></summary>

```
=== RUN   TestThatFlakes
=== PAUSE TestThatFlakes
=== CONT  TestThatFlakes
    imperfect_test.go:36: oh, I'm so flaky
--- FAIL: TestThatFlakes (0.00s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/imperfect.TestThatFlakesToo</code></summary>

```
=== RUN   TestThatFlakesToo
=== PAUSE TestThatFlakesToo
=== CONT  TestThatFlakesToo
    imperfect_test.go:46: oh, I'm even more flaky
--- FAIL: TestThatFlakesToo (0.00s)

```
</details>
</details>

### Metrics

```
pass: 15, fail: 20, data races: 6, slow: 4, cached pkg runs: 5, total pkg: 2
```

Elapsed: 4.13s
Slow: 4s

### Test time distribution (seconds)
```
[ min  max] cnt total%  sum (35 events)
[0.00 0.00] 23 65.71% 0.00 .................................................................
[0.01 0.01]  3  8.57% 0.03 ........
[0.02 0.02]  5 14.29% 0.10 ..............
[1.00 1.00]  4 11.43% 4.00 ...........

```
### Flaky tests
<details>
<summary>Tests: 4</summary>

| Pass | Fail | Test |
| - | - | - |
| 1 | 7 | github.com/vearutop/teststat/imperfect.TestThatIsRacy |
| 1 | 1 | github.com/vearutop/teststat/imperfect.TestThatIsAlwaysSlow |
| 1 | 7 | github.com/vearutop/teststat/imperfect.TestThatFlakesToo |
| 1 | 5 | github.com/vearutop/teststat/imperfect.TestThatFlakes |
</details>

### Slow tests
<details>
<summary>Total slow runs: 4</summary>

| Result | Duration | Package | Test |
| - | - | - | - |
| pass | 1s | github.com/vearutop/teststat/imperfect/foo | TestThatIsAlwaysSlowFoo |
| pass | 1s | github.com/vearutop/teststat/imperfect/foo | TestThatIsSometimesSlowFoo |
| fail | 1s | github.com/vearutop/teststat/imperfect | TestThatIsAlwaysSlow |
| pass | 1s | github.com/vearutop/teststat/imperfect | TestThatIsAlwaysSlow |
</details>

### Data races
<details>
<summary>Total data races: 3, unique: 1</summary>

<details>
<summary><code>github.com/vearutop/teststat/imperfect.TestThatFlakesToo</code></summary>

Other affected tests:
```
github.com/vearutop/teststat/imperfect.TestThatIsRacy
github.com/vearutop/teststat/imperfect.TestThatFlakes
```

```
=== RUN   TestThatFlakesToo
=== PAUSE TestThatFlakesToo
=== CONT  TestThatFlakesToo
    imperfect_test.go:46: oh, I'm even more flaky
--- FAIL: TestThatFlakesToo (0.00s)
==================
WARNING: DATA RACE
Read at 0x00c000184018 by goroutine 9:
  github.com/vearutop/teststat/imperfect_test.TestThatIsRacy.func1()
      /Users/vearutop/dev/teststat/imperfect/imperfect_test.go:22 +0x2e

Previous write at 0x00c000184018 by goroutine 10:
  github.com/vearutop/teststat/imperfect_test.TestThatIsRacy.func1()
      /Users/vearutop/dev/teststat/imperfect/imperfect_test.go:22 +0x44

Goroutine 9 (running) created at:
  github.com/vearutop/teststat/imperfect_test.TestThatIsRacy()
      /Users/vearutop/dev/teststat/imperfect/imperfect_test.go:22 +0x97
  testing.tRunner()
      /usr/local/opt/go/libexec/src/testing/testing.go:1595 +0x238
  testing.(*T).Run.func1()
      /usr/local/opt/go/libexec/src/testing/testing.go:1648 +0x44

Goroutine 10 (finished) created at:
  github.com/vearutop/teststat/imperfect_test.TestThatIsRacy()
      /Users/vearutop/dev/teststat/imperfect/imperfect_test.go:22 +0x97
  testing.tRunner()
      /usr/local/opt/go/libexec/src/testing/testing.go:1595 +0x238
  testing.(*T).Run.func1()
      /usr/local/opt/go/libexec/src/testing/testing.go:1648 +0x44
==================

```
</details>

</details>

### Slowest test packages
<details>
<summary>Total packages with tests: 2</summary>

| Duration | Package |
| - | - |
| 1.27s | github.com/vearutop/teststat/imperfect |
| 0s (cached) | github.com/vearutop/teststat/imperfect/foo |
</details>

