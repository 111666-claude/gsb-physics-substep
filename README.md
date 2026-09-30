# physics-substep

固定步长物理推进：按真实时间累积子步，单帧子步有上限，只用 Go 标准库。

```
go test ./...
go run ./cmd/physics-substep --sample rate
go run ./cmd/physics-substep --sample fps
go run ./cmd/physics-substep --sample spike
```

## 口径（README 为准）

- **按真实时间累积**：`accumulator += frameMs`，子步数 `= floor(accumulator / FixedStepMs)`，
  最多 `MaxSubsteps` 步；跑完把 `accumulator` 减去已执行的步长。
- **帧率无关**：同样的真实时间，分几帧喂进来，累计子步数必须相同。
- **卡顿保护**：单帧最多 `MaxSubsteps` 步，溢出的时间要丢弃并累加进 `Dropped()`。
- **不变量**：`accumulator` 小于一个步长加上限步数乘步长；步数单调不减；
  同一串帧输入重复执行结果相同。
- **规模**：每秒 100 帧、单帧 1 万个刚体，单帧推进 O(1)，内存 O(1)。

## 输出契约（不改格式）

```
steps=5 ran=5 rest=0
steps=5
ran=5 steps=5 dropped=245
steps=2 rest=10
```

## 常量

```
FixedStepMs = 20
MaxSubsteps = 5
```

## 目录

```
substep.go              固定步长推进
cmd/physics-substep     命令行入口
substep_test.go         go test 用例
```
