### Failures
<details>
<summary>Failed tests (including flaky): 4</summary>

<details>
<summary><code>github.com/vearutop/teststat/app.TestThatFlakesToo</code></summary>

```
=== RUN   TestThatFlakesToo
=== PAUSE TestThatFlakesToo
=== CONT  TestThatFlakesToo
    imperfect_test.go:47: oh, I'm even more flaky
--- FAIL: TestThatFlakesToo (0.00s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/foo.TestThatFlakesFoo</code></summary>

```
=== RUN   TestThatFlakesFoo
=== PAUSE TestThatFlakesFoo
=== CONT  TestThatFlakesFoo
    imperfect_test.go:40: oh, I'm so flaky
--- FAIL: TestThatFlakesFoo (0.00s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/foo.Test_Suite</code></summary>

```
=== RUN   Test_Suite
--- FAIL: Test_Suite (0.01s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/foo.Test_Suite/TestThatFlakesToo</code></summary>

```
=== RUN   Test_Suite/TestThatFlakesToo
=== PAUSE Test_Suite/TestThatFlakesToo
=== CONT  Test_Suite/TestThatFlakesToo
=== CONT  Test_Suite/TestThatFlakesToo
    imperfect_test.go:93: oh, I'm so flaky
    --- FAIL: Test_Suite/TestThatFlakesToo (0.00s)

```
</details>
</details>

### Metrics

```
pass: 19, fail: 10, data races: 2, slow: 9, total pkg: 3
```

Elapsed: 5.05s
Slow: 5.05s

### Test time distribution (seconds)
```
[ min  max] cnt total%  sum (29 events)
[0.00 0.00] 20 68.97% 0.00 ....................................................................
[0.01 0.01]  3 10.34% 0.03 ..........
[0.02 0.02]  1  3.45% 0.02 ...
[1.00 1.00]  5 17.24% 5.00 .................

```
### Flaky tests
<details>
<summary>Tests: 6</summary>

| Pass | Fail | Test |
| - | - | - |
| 1 | 1 | github.com/vearutop/teststat/foo.Test_Suite/TestThatFlakesToo |
| 1 | 1 | github.com/vearutop/teststat/foo.Test_Suite |
| 1 | 1 | github.com/vearutop/teststat/foo.TestThatFlakesFoo |
| 1 | 1 | github.com/vearutop/teststat/app.TestThatIsRacy |
| 1 | 1 | github.com/vearutop/teststat/app.TestThatIsAlwaysSlow |
| 1 | 1 | github.com/vearutop/teststat/app.TestThatFlakes |
</details>

### Slow tests
<details>
<summary>Total slow runs: 9</summary>

| Result | Duration | Package | Test |
| - | - | - | - |
</details>

### Data races
<details>
<summary>Total data races: 2, unique: 1</summary>

<details>
<summary><code>github.com/vearutop/teststat/app.TestThatFlakes</code></summary>

Other affected tests:
```
github.com/vearutop/teststat.TestThatIsRacy
```

```
=== RUN   TestThatIsRacy
=== PAUSE TestThatIsRacy
=== CONT  TestThatIsRacy
==================
WARNING: DATA RACE
Read at 0x00c000094018 by goroutine 13:
  github.com/vearutop/teststat_test.TestThatIsRacy.func1()
      /Users/vearutop/dev/teststat/imperfect_test.go:23 +0x30

Previous write at 0x00c000094018 by goroutine 10:
  github.com/vearutop/teststat_test.TestThatIsRacy.func1()
      /Users/vearutop/dev/teststat/imperfect_test.go:23 +0x44

Goroutine 13 (running) created at:
  github.com/vearutop/teststat_test.TestThatIsRacy()
      /Users/vearutop/dev/teststat/imperfect_test.go:23 +0x9e
  testing.tRunner()
      /usr/local/opt/go/libexec/src/testing/testing.go:1446 +0x216
  testing.(*T).Run.func1()
      /usr/local/opt/go/libexec/src/testing/testing.go:1493 +0x47

Goroutine 10 (finished) created at:
  github.com/vearutop/teststat_test.TestThatIsRacy()
      /Users/vearutop/dev/teststat/imperfect_test.go:23 +0x9e
  testing.tRunner()
      /usr/local/opt/go/libexec/src/testing/testing.go:1446 +0x216
  testing.(*T).Run.func1()
      /usr/local/opt/go/libexec/src/testing/testing.go:1493 +0x47
==================
    testing.go:1319: race detected during execution of test
--- FAIL: TestThatIsRacy (0.01s)

```
</details>

</details>

### Slowest test packages
<details>
<summary>Total packages with tests: 3</summary>

| Duration | Package |
| - | - |
| 1.484999999s | github.com/vearutop/teststat/app |
| 1.132s | github.com/vearutop/teststat |
</details>

