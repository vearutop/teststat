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
pass: 15, fail: 7, slow: 3, cached pkg runs: 6, total pkg: 2
```

Elapsed: 3.01s
Slow: 3s

### Test time distribution (seconds)
```
[ min  max] cnt total%  sum (22 events)
[0.00 0.00] 18 81.82% 0.00 .................................................................................
[0.01 0.01]  1  4.55% 0.01 ....
[1.00 1.00]  3 13.64% 3.00 .............

```
### Flaky tests
<details>
<summary>Tests: 2</summary>

| Pass | Fail | Test |
| - | - | - |
| 1 | 2 | github.com/vearutop/teststat/imperfect.TestThatFlakesToo |
| 1 | 5 | github.com/vearutop/teststat/imperfect.TestThatFlakes |
</details>

### Slow tests
<details>
<summary>Total slow runs: 3</summary>

| Result | Duration | Package | Test |
| - | - | - | - |
| pass | 1s | github.com/vearutop/teststat/imperfect/foo | TestThatIsSometimesSlowFoo |
| pass | 1s | github.com/vearutop/teststat/imperfect/foo | TestThatIsAlwaysSlowFoo |
| pass | 1s | github.com/vearutop/teststat/imperfect | TestThatIsAlwaysSlow |
</details>

### Slowest test packages
<details>
<summary>Total packages with tests: 2</summary>

| Duration | Package |
| - | - |
| 1.234s | github.com/vearutop/teststat/imperfect |
| 0s (cached) | github.com/vearutop/teststat/imperfect/foo |
</details>

