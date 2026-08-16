# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

船期延误后重排出来的进港预约时间窗是坏的，请帮我修复。

原预约窗口是 01:00–02:00，船延误、到 06:00 才触发重排；重排后预约状态是 rescheduled，但窗口变成 start=2026-01-01T06:00:00Z、end=2026-01-01T02:00:00Z，结束时间比开始时间还早，窗口长度成了负数，闸口按这个窗口根本放不进车。只要重排发生在原窗口开始之后就会这样。

期望：重排后的时间窗仍然是可用窗口——开始时间不早于重排时刻，结束时间晚于开始时间，窗口长度与原来一致；重排的优先级顺序与状态流转保持现有行为。修复后请保证 go test -timeout=120s -count=1 ./... 全绿，不要修改或跳过测试。

## 含 Bug 版本

- 仓库：11DingKing/goZZ-02
- 仓库地址：https://github.com/11DingKing/goZZ-02.git
- parent SHA：6fcf75de9ccb76d4e8a696e42ed934dab1ac4c55

## 复现步骤

```bash
git clone -- https://github.com/11DingKing/goZZ-02.git bug-repro
cd bug-repro
git checkout --detach 6fcf75de9ccb76d4e8a696e42ed934dab1ac4c55
go test ./internal/portapp -run "^TestRescheduleKeepsPickupWindowConsistent$" -count=1 -v
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/portapp -run "^TestRescheduleKeepsPickupWindowConsistent$" -count=1 -v
=== RUN   TestRescheduleKeepsPickupWindowConsistent
    reschedule_window_test.go:50: rescheduled window must end after it starts, got start=2026-01-01T06:00:00Z end=2026-01-01T02:00:00Z
--- FAIL: TestRescheduleKeepsPickupWindowConsistent (0.00s)
FAIL
FAIL	arcticdispatch/internal/portapp	0.038s
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
$ go test ./internal/portapp -run "^TestRescheduleKeepsPickupWindowConsistent$" -count=1 -v
=== RUN   TestRescheduleKeepsPickupWindowConsistent
    reschedule_window_test.go:50: rescheduled window must end after it starts, got start=2026-01-01T06:00:00Z end=2026-01-01T02:00:00Z
--- FAIL: TestRescheduleKeepsPickupWindowConsistent (0.00s)
FAIL
FAIL	arcticdispatch/internal/portapp	0.002s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

定向测试通过：go test ./internal/portapp -run '^TestRescheduleKeepsPickupWindowConsistent$' -count=1 -v
全量回归 go test -timeout=120s -count=1 ./... 通过，go build ./... 与 go vet ./... 通过
断言重排后 WindowEnd 晚于 WindowStart、窗口时长仍为原始 1h、WindowStart 不早于重排时刻；既有的延误重排优先级测试保持通过
