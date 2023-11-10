### Failures
<details>
<summary>Failed builds</summary>

```
# github.com/vearutop/teststat/broken_test [github.com/vearutop/teststat/broken.test]
broken/build_failed_test.go:8:2: a declared and not used
# github.com/vearutop/teststat/broken/deeper_test [github.com/vearutop/teststat/broken/deeper.test]
broken/deeper/build_failed_test.go:8:2: a declared and not used
FAIL	github.com/vearutop/teststat/broken [build failed]
FAIL	github.com/vearutop/teststat/broken/deeper [build failed]
```

</details>

<details>
<summary>Failed tests (including flaky): 2</summary>

<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestAlwaysFails</code></summary>

```
=== RUN   TestAlwaysFails
--- FAIL: TestAlwaysFails (0.00s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestThatPanics</code></summary>

```
=== RUN   TestThatPanics
=== PAUSE TestThatPanics
=== CONT  TestThatPanics
--- FAIL: TestThatPanics (0.00s)
panic: ouch [recovered]
	panic: ouch

goroutine 19 [running]:
testing.tRunner.func1.2({0x122e8e0, 0x129e4d0})
	/usr/local/opt/go/libexec/src/testing/testing.go:1545 +0x366
testing.tRunner.func1()
	/usr/local/opt/go/libexec/src/testing/testing.go:1548 +0x630
panic({0x122e8e0?, 0x129e4d0?})
	/usr/local/opt/go/libexec/src/runtime/panic.go:920 +0x270
github.com/vearutop/teststat/broken/other_test.TestThatPanics(0x0?)
	/Users/vearutop/dev/teststat/broken/other/failed_test.go:16 +0x3a
testing.tRunner(0xc000102ea0, 0x126b1e8)
	/usr/local/opt/go/libexec/src/testing/testing.go:1595 +0x239
created by testing.(*T).Run in goroutine 1
	/usr/local/opt/go/libexec/src/testing/testing.go:1648 +0x82b

```
</details>
</details>

<details>
<summary>Unfinished tests: 1</summary>

<details>
<summary><code>github.com/vearutop/teststat/broken/goroutine.TestThatPanicsInAGoroutine</code></summary>

```
=== RUN   TestThatPanicsInAGoroutine
=== PAUSE TestThatPanicsInAGoroutine
=== CONT  TestThatPanicsInAGoroutine
panic: ouch2

goroutine 19 [running]:
github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine.func1()
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:16 +0x2b
created by github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine in goroutine 18
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:15 +0x3d

```
</details>
</details>

### Metrics

```
pass: 4, fail: 2, unfinished: 1, slow: 1, total pkg: 4
```

Elapsed: 30ms
Slow: 30ms

### Test time distribution (seconds)
```
[ min  max] cnt total%  sum (6 events)
[0.00 0.00] 5 83.33% 0.00 ...................................................................................
[0.03 0.03] 1 16.67% 0.03 ................

```
### Slow tests
<details>
<summary>Total slow runs: 1</summary>

| Result | Duration | Package | Test |
| - | - | - | - |
</details>

### Slowest test packages
<details>
<summary>Total packages with tests: 4</summary>

| Duration | Package |
| - | - |
| 1.823999999s | github.com/vearutop/teststat/broken/goroutine |
| 1.562s | github.com/vearutop/teststat/app |
</details>

