# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA，再应用仓库内的可信测试补丁；不要在当前修复结果源码上期待重新出现修复前失败。

## 问题现象

任务清单会接受不可能的持续时间，导致冲突漏报，请帮我修复。

JSON 和 CSV 中填 0s、负数，或超大的纯数字/min/hour/sec 时，文件仍可能加载成功，展开出的执行 End 不晚于 Start，后续冲突检测就静默漏掉这些任务。

期望所有 duration 语法统一要求大于零，单位换算不能溢出 time.Duration；非法值在 LoadFile 阶段返回带任务/行上下文的错误，合法边界和普通复合 duration 保持兼容。修复后请保证 go test ./... 全绿。

## 含 Bug 版本

- 仓库：zhanglei10281852-gif/gogo-13
- 仓库地址：https://github.com/zhanglei10281852-gif/gogo-13.git
- parent SHA：9e18975fb1bc9075e093d94a2713daefecfa83b8

## 复现步骤

```bash
git clone -- https://github.com/zhanglei10281852-gif/gogo-13.git bug-repro
cd bug-repro
git checkout --detach 9e18975fb1bc9075e093d94a2713daefecfa83b8
git apply ../BENZHI_VALIDATION/trusted-test.patch
go test ./pkg/task -run "^TestLoadFile" -count=1 -v
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./pkg/task -run "^TestLoadFile" -count=1 -v
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0s
    task_public_test.go:54: LoadFile accepted duration "0s": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0xc000102120}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1s
    task_public_test.go:54: LoadFile accepted duration "-1s": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1s Host:host-a Resource: Schedule:0xc000102240}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0
    task_public_test.go:54: LoadFile accepted duration "0": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0xc000102360}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1
    task_public_test.go:54: LoadFile accepted duration "-1": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1m0s Host:host-a Resource: Schedule:0xc000102480}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0min
    task_public_test.go:54: LoadFile accepted duration "0min": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0xc0001025a0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1hour
    task_public_test.go:54: LoadFile accepted duration "-1hour": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1h0m0s Host:host-a Resource: Schedule:0xc0001026c0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/153722868
    task_public_test.go:54: LoadFile accepted duration "153722868": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h46m33.709551616s Host:host-a Resource: Schedule:0xc0001027e0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/153722868min
    task_public_test.go:54: LoadFile accepted duration "153722868min": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h46m33.709551616s Host:host-a Resource: Schedule:0xc000102900}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/2562048hour
    task_public_test.go:54: LoadFile accepted duration "2562048hour": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h34m33.709551616s Host:host-a Resource: Schedule:0xc000102a20}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/9223372037sec
    task_public_test.go:54: LoadFile accepted duration "9223372037sec": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h47m16.709551616s Host:host-a Resource: Schedule:0xc000102b40}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0s
    task_public_test.go:54: LoadFile accepted duration "0s": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0xc0001c0000}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1s
    task_public_test.go:54: LoadFile accepted duration "-1s": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1s Host:host-a Resource: Schedule:0xc0001c0120}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0
    task_public_test.go:54: LoadFile accepted duration "0": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0xc0001c0240}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1
    task_public_test.go:54: LoadFile accepted duration "-1": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1m0s Host:host-a Resource: Schedule:0xc0001c0360}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0min
    task_public_test.go:54: LoadFile accepted duration "0min": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0xc0001c0480}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1hour
    task_public_test.go:54: LoadFile accepted duration "-1hour": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1h0m0s Host:host-a Resource: Schedule:0xc0001c05a0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/153722868
    task_public_test.go:54: LoadFile accepted duration "153722868": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h46m33.709551616s Host:host-a Resource: Schedule:0xc0001c06c0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/153722868min
    task_public_test.go:54: LoadFile accepted duration "153722868min": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h46m33.709551616s Host:host-a Resource: Schedule:0xc0001c07e0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/2562048hour
    task_public_test.go:54: LoadFile accepted duration "2562048hour": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h34m33.709551616s Host:host-a Resource: Schedule:0xc0001c0900}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/9223372037sec
    task_public_test.go:54: LoadFile accepted duration "9223372037sec": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h47m16.709551616s Host:host-a Resource: Schedule:0xc0001c0a20}]
--- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations (0.02s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0s (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1s (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0min (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1hour (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/153722868 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/153722868min (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/2562048hour (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/9223372037sec (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0s (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1s (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0min (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1hour (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/153722868 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/153722868min (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/2562048hour (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/9223372037sec (0.00s)
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/1ns
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/1m30s
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/15
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/153722867
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/5min
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/153722867min
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/2hour
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/2562047hour
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/3sec
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/9223372036sec
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/1ns
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/1m30s
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/15
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/153722867
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/5min
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/153722867min
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/2hour
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/2562047hour
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/3sec
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/9223372036sec
--- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations (0.02s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/1ns (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/1m30s (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/15 (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/153722867 (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/5min (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/153722867min (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/2hour (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/2562047hour (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/3sec (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/9223372036sec (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/1ns (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/1m30s (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/15 (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/153722867 (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/5min (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/153722867min (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/2hour (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/2562047hour (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/3sec (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/9223372036sec (0.00s)
FAIL
FAIL	github.com/ops/cronchecker/pkg/task	0.045s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./pkg/task -run "^TestLoadFile" -count=1 -v
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0s
    task_public_test.go:54: LoadFile accepted duration "0s": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0x40001be120}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1s
    task_public_test.go:54: LoadFile accepted duration "-1s": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1s Host:host-a Resource: Schedule:0x40000f4000}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0
    task_public_test.go:54: LoadFile accepted duration "0": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0x40000f4120}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1
    task_public_test.go:54: LoadFile accepted duration "-1": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1m0s Host:host-a Resource: Schedule:0x40000f4240}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0min
    task_public_test.go:54: LoadFile accepted duration "0min": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0x40000f4360}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1hour
    task_public_test.go:54: LoadFile accepted duration "-1hour": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1h0m0s Host:host-a Resource: Schedule:0x40000f4480}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/153722868
    task_public_test.go:54: LoadFile accepted duration "153722868": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h46m33.709551616s Host:host-a Resource: Schedule:0x40000f45a0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/153722868min
    task_public_test.go:54: LoadFile accepted duration "153722868min": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h46m33.709551616s Host:host-a Resource: Schedule:0x40000f46c0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/2562048hour
    task_public_test.go:54: LoadFile accepted duration "2562048hour": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h34m33.709551616s Host:host-a Resource: Schedule:0x40000f47e0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/9223372037sec
    task_public_test.go:54: LoadFile accepted duration "9223372037sec": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h47m16.709551616s Host:host-a Resource: Schedule:0x40000f4900}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0s
    task_public_test.go:54: LoadFile accepted duration "0s": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0x40000f4a20}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1s
    task_public_test.go:54: LoadFile accepted duration "-1s": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1s Host:host-a Resource: Schedule:0x40000f4b40}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0
    task_public_test.go:54: LoadFile accepted duration "0": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0x40000f4c60}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1
    task_public_test.go:54: LoadFile accepted duration "-1": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1m0s Host:host-a Resource: Schedule:0x40000f4d80}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0min
    task_public_test.go:54: LoadFile accepted duration "0min": [{Name:job CronExpr:0 10 * * * Command:run Duration:0s Host:host-a Resource: Schedule:0x40000f4ea0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1hour
    task_public_test.go:54: LoadFile accepted duration "-1hour": [{Name:job CronExpr:0 10 * * * Command:run Duration:-1h0m0s Host:host-a Resource: Schedule:0x40000f4fc0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/153722868
    task_public_test.go:54: LoadFile accepted duration "153722868": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h46m33.709551616s Host:host-a Resource: Schedule:0x40000f50e0}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/153722868min
    task_public_test.go:54: LoadFile accepted duration "153722868min": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h46m33.709551616s Host:host-a Resource: Schedule:0x40000f5200}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/2562048hour
    task_public_test.go:54: LoadFile accepted duration "2562048hour": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h34m33.709551616s Host:host-a Resource: Schedule:0x40000f5320}]
=== RUN   TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/9223372037sec
    task_public_test.go:54: LoadFile accepted duration "9223372037sec": [{Name:job CronExpr:0 10 * * * Command:run Duration:-2562047h47m16.709551616s Host:host-a Resource: Schedule:0x40000f5440}]
--- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations (0.08s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0s (0.03s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1s (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/0min (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/-1hour (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/153722868 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/153722868min (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/2562048hour (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/json/9223372037sec (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0s (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1s (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/0min (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/-1hour (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/153722868 (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/153722868min (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/2562048hour (0.00s)
    --- FAIL: TestLoadFileRejectsNonPositiveAndOverflowingDurations/csv/9223372037sec (0.00s)
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/1ns
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/1m30s
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/15
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/153722867
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/5min
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/153722867min
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/2hour
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/2562047hour
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/3sec
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/json/9223372036sec
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/1ns
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/1m30s
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/15
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/153722867
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/5min
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/153722867min
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/2hour
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/2562047hour
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/3sec
=== RUN   TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/9223372036sec
--- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations (0.05s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/1ns (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/1m30s (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/15 (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/153722867 (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/5min (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/153722867min (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/2hour (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/2562047hour (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/3sec (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/json/9223372036sec (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/1ns (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/1m30s (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/15 (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/153722867 (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/5min (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/153722867min (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/2hour (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/2562047hour (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/3sec (0.00s)
    --- PASS: TestLoadFileAndExpandExecutionsAcceptValidDurations/csv/9223372036sec (0.00s)
FAIL
FAIL	github.com/ops/cronchecker/pkg/task	0.252s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

JSON/CSV 的非正和各单位溢出均被拒绝；合法边界正确展开且 End>Start；定向、全量、build、vet 在双架构通过；不得跳过测试。
