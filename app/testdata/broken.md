### Failures
<details>
<summary>Failed builds</summary>

```
# github.com/vearutop/teststat/broken_test [github.com/vearutop/teststat/broken.test]
../../broken/build_failed_test.go:8:2: a declared and not used
# github.com/vearutop/teststat/broken/deeper_test [github.com/vearutop/teststat/broken/deeper.test]
../../broken/deeper/build_failed_test.go:8:2: a declared and not used
FAIL	github.com/vearutop/teststat/broken [build failed]
FAIL	github.com/vearutop/teststat/broken/deeper [build failed]
```

</details>

<details>
<summary>Failed tests (including flaky): 3</summary>

<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestAlwaysFails</code></summary>

```
=== RUN   TestAlwaysFails
--- FAIL: TestAlwaysFails (0.00s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestAlwaysFailsInSubtest</code></summary>

```
=== RUN   TestAlwaysFailsInSubtest
--- FAIL: TestAlwaysFailsInSubtest (0.00s)

```
</details>
<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestAlwaysFailsInSubtest//-&?\[]!@#$%^*()abc123_+=</code></summary>

```
=== RUN   TestAlwaysFailsInSubtest//-&?\[]!@#$%^*()abc123_+=
    --- FAIL: TestAlwaysFailsInSubtest//-&?\[]!@#$%^*()abc123_+= (0.00s)
panic: can't cope [recovered]
	panic: can't cope

goroutine 19 [running]:
testing.tRunner.func1.2({0x11bc300, 0x1219900})
	/usr/local/opt/go/libexec/src/testing/testing.go:1545 +0x366
testing.tRunner.func1()
	/usr/local/opt/go/libexec/src/testing/testing.go:1548 +0x630
panic({0x11bc300?, 0x1219900?})
	/usr/local/opt/go/libexec/src/runtime/panic.go:920 +0x270
github.com/vearutop/teststat/broken/other_test.TestAlwaysFailsInSubtest.func1(0x0?)
	/Users/vearutop/dev/teststat/broken/other/failed_test.go:21 +0x2b
testing.tRunner(0xc000236680, 0x11f0430)
	/usr/local/opt/go/libexec/src/testing/testing.go:1595 +0x239
created by testing.(*T).Run in goroutine 18
	/usr/local/opt/go/libexec/src/testing/testing.go:1648 +0x82b

```
</details>
</details>

<details>
<summary>Unfinished tests: 2</summary>

<details>
<summary><code>github.com/vearutop/teststat/broken/goroutine.TestThatPanicsInAGoroutine</code></summary>

```
=== RUN   TestThatPanicsInAGoroutine
=== PAUSE TestThatPanicsInAGoroutine
=== CONT  TestThatPanicsInAGoroutine
panic: ouch2

goroutine 35 [running]:
github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine.func1()
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:16 +0x2b
created by github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine in goroutine 34
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:15 +0x3d
=== RUN   TestThatPanicsInAGoroutine
=== PAUSE TestThatPanicsInAGoroutine
=== CONT  TestThatPanicsInAGoroutine
panic: ouch2

goroutine 19 [running]:
github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine.func1()
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:16 +0x2b
created by github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine in goroutine 18
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:15 +0x3d
=== RUN   TestThatPanicsInAGoroutine
=== PAUSE TestThatPanicsInAGoroutine
=== CONT  TestThatPanicsInAGoroutine
panic: ouch2

goroutine 35 [running]:
github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine.func1()
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:16 +0x2b
created by github.com/vearutop/teststat/broken/goroutine_test.TestThatPanicsInAGoroutine in goroutine 34
	/Users/vearutop/dev/teststat/broken/goroutine/failed_test.go:15 +0x3d
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
<details>
<summary><code>github.com/vearutop/teststat/broken/other.TestThatPanics</code></summary>

```
=== RUN   TestThatPanics
=== PAUSE TestThatPanics
=== RUN   TestThatPanics
=== PAUSE TestThatPanics
=== RUN   TestThatPanics
=== PAUSE TestThatPanics
=== RUN   TestThatPanics
=== PAUSE TestThatPanics

```
</details>
</details>

### Metrics

```
pass: 0, fail: 12, unfinished: 2, total pkg: 2
```

Elapsed: 0s
Slow: 0s

### Test time distribution (seconds)
```
[ min  max] cnt total%  sum (12 events)
[0.00 0.00] 12 100.00% 0.00 ....................................................................................................

```
### Slowest test packages
<details>
<summary>Total packages with tests: 2</summary>

| Duration | Package |
| - | - |
| 1.164s | github.com/vearutop/teststat/broken/goroutine |
| 374ms | github.com/vearutop/teststat/broken/other |
</details>

