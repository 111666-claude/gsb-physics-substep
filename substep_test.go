package substep

import (
	"math/rand"
	"testing"
)

func TestAdvanceRunsOneStepForOneStep(t *testing.T) {
	stepper := &Stepper{}
	if got := stepper.Advance(FixedStepMs); got != 1 {
		t.Fatalf("正好一个步长应该跑一步：%d", got)
	}
}

func TestStepsCountUp(t *testing.T) {
	stepper := &Stepper{}
	stepper.Advance(FixedStepMs)
	stepper.Advance(FixedStepMs)
	if stepper.Steps() != 2 {
		t.Fatalf("步数应该累加：%d", stepper.Steps())
	}
}

func TestDroppedStartsAtZero(t *testing.T) {
	if (&Stepper{}).Dropped() != 0 {
		t.Fatal("初始丢弃数应该是 0")
	}
}

func TestConstants(t *testing.T) {
	if FixedStepMs != 20 || MaxSubsteps != 5 {
		t.Fatal("物理步长常量被改了")
	}
}

func repeatFrames(frameMs float64, count int) []float64 {
	frames := make([]float64, count)
	for index := range frames {
		frames[index] = frameMs
	}
	return frames
}

func TestFrameRateIndependence(t *testing.T) {
	cases := []struct {
		name   string
		frames []float64
		steps  int
		rest   float64
	}{
		{"单帧100ms", []float64{100}, 5, 0},
		{"10帧x10ms", repeatFrames(10, 10), 5, 0},
		{"5帧x20ms", repeatFrames(20, 5), 5, 0},
		{"2帧x50ms", []float64{50, 50}, 5, 0},
		{"16帧x6.25ms", repeatFrames(6.25, 16), 5, 0},
		{"不规则分帧100ms", []float64{7.5, 2.5, 40, 25, 12.5, 12.5}, 5, 0},
		{"3帧x10ms", []float64{10, 10, 10}, 1, 10},
		{"单帧30ms", []float64{30}, 1, 10},
		{"2帧x15ms", []float64{15, 15}, 1, 10},
		{"分数帧30.5ms", []float64{10.25, 10.25, 10}, 1, 10.5},
		{"单帧30.5ms", []float64{30.5}, 1, 10.5},
	}
	for _, testCase := range cases {
		stepper := &Stepper{}
		total := 0
		for _, frame := range testCase.frames {
			total += stepper.Advance(frame)
		}
		if total != testCase.steps || stepper.Steps() != testCase.steps {
			t.Fatalf("%s：子步数 %d（累计 %d），口径 %d", testCase.name, total, stepper.Steps(), testCase.steps)
		}
		if rest := stepper.Accumulator(0); rest != testCase.rest {
			t.Fatalf("%s：残留 %v，口径 %v", testCase.name, rest, testCase.rest)
		}
	}
}

func TestSpikeCapAndDropped(t *testing.T) {
	stepper := &Stepper{}
	if ran := stepper.Advance(5000); ran != MaxSubsteps {
		t.Fatalf("卡顿帧最多跑 %d 步，实际 %d", MaxSubsteps, ran)
	}
	if stepper.Steps() != 5 || stepper.Dropped() != 245 {
		t.Fatalf("5 秒卡顿口径 steps=5 dropped=245，实际 steps=%d dropped=%d", stepper.Steps(), stepper.Dropped())
	}
	if rest := stepper.Accumulator(0); rest != 0 {
		t.Fatalf("溢出时间丢弃后残留应为 0，实际 %v", rest)
	}
	if ran := stepper.Advance(FixedStepMs); ran != 1 {
		t.Fatalf("卡顿不得污染后续帧，应跑 1 步，实际 %d", ran)
	}

	overrun := &Stepper{}
	if ran := overrun.Advance(120); ran != 5 || overrun.Dropped() != 1 {
		t.Fatalf("120ms 应跑 5 步丢 1 步，实际 ran=%d dropped=%d", ran, overrun.Dropped())
	}

	nearCap := &Stepper{}
	if ran := nearCap.Advance(119.75); ran != 5 || nearCap.Dropped() != 0 {
		t.Fatalf("119.75ms 应跑 5 步不丢，实际 ran=%d dropped=%d", ran, nearCap.Dropped())
	}
	if rest := nearCap.Accumulator(0); rest != 19.75 {
		t.Fatalf("119.75ms 残留应为 19.75，实际 %v", rest)
	}
}

func TestResidualAccumulatesAcrossFrames(t *testing.T) {
	stepper := &Stepper{}
	wantRan := []int{0, 1, 0, 1}
	for index, want := range wantRan {
		if got := stepper.Advance(10); got != want {
			t.Fatalf("第 %d 帧应跑 %d 步，实际 %d", index, want, got)
		}
	}
	if stepper.Steps() != 2 || stepper.Accumulator(0) != 0 {
		t.Fatalf("40ms 应累计 2 步且无残留，实际 steps=%d rest=%v", stepper.Steps(), stepper.Accumulator(0))
	}
	if got := stepper.Advance(0); got != 0 {
		t.Fatalf("零帧长不应推进，实际 %d", got)
	}

	fraction := &Stepper{}
	fraction.Advance(19.75)
	if got := fraction.Advance(0.25); got != 1 {
		t.Fatalf("19.75+0.25 应凑出 1 步，实际 %d", got)
	}
	if rest := fraction.Accumulator(0); rest != 0 {
		t.Fatalf("凑整后应无残留，实际 %v", rest)
	}
}

func TestInvariantsMonotonicBoundedDeterministic(t *testing.T) {
	frames := []float64{10, 0, 33.5, 5000, 16.6, 20, 7.25, 100, 0.5, 250}
	run := func() (int, int, float64) {
		stepper := &Stepper{}
		previous := 0
		for _, frame := range frames {
			stepper.Advance(frame)
			if stepper.Steps() < previous {
				t.Fatalf("步数必须单调不减：%d -> %d", previous, stepper.Steps())
			}
			previous = stepper.Steps()
			if rest := stepper.Accumulator(0); rest < 0 || rest >= FixedStepMs+MaxSubsteps*FixedStepMs {
				t.Fatalf("accumulator 必须小于一个步长加上限步数乘步长，实际 %v", rest)
			}
		}
		return stepper.Steps(), stepper.Dropped(), stepper.Accumulator(0)
	}
	steps1, dropped1, rest1 := run()
	steps2, dropped2, rest2 := run()
	if steps1 != steps2 || dropped1 != dropped2 || rest1 != rest2 {
		t.Fatal("同一串帧输入重复执行结果必须相同")
	}
}

// referenceStepper 是「按真实时间累积 + 子步上限」的独立参考实现：
// 用循环逐步消耗，与被测实现的整除写法互为对照。
type referenceStepper struct {
	accumulator float64
	steps       int
	dropped     int
}

func (r *referenceStepper) advance(frameMs float64) int {
	r.accumulator += frameMs
	ran := 0
	for r.accumulator >= FixedStepMs && ran < MaxSubsteps {
		r.accumulator -= FixedStepMs
		r.steps++
		ran++
	}
	for r.accumulator >= FixedStepMs {
		r.accumulator -= FixedStepMs
		r.dropped++
	}
	return ran
}

func TestAgainstReferenceRandomSequences(t *testing.T) {
	rng := rand.New(rand.NewSource(20261001))
	for trial := 0; trial < 200; trial++ {
		length := 1 + rng.Intn(40)
		frames := make([]float64, length)
		for index := range frames {
			switch rng.Intn(6) {
			case 0:
				frames[index] = 0
			case 1:
				frames[index] = float64(rng.Intn(250)+1) * FixedStepMs
			case 2:
				frames[index] = float64(rng.Intn(81)) * 0.25
			default:
				frames[index] = float64(rng.Intn(161)) * 0.25
			}
		}
		stepper := &Stepper{}
		reference := &referenceStepper{}
		for index, frame := range frames {
			gotRan := stepper.Advance(frame)
			wantRan := reference.advance(frame)
			if gotRan != wantRan {
				t.Fatalf("序列 %d 第 %d 帧（%vms）：子步数 %d != 参考 %d", trial, index, frame, gotRan, wantRan)
			}
			if stepper.Steps() != reference.steps {
				t.Fatalf("序列 %d 第 %d 帧：累计步数 %d != 参考 %d", trial, index, stepper.Steps(), reference.steps)
			}
			if stepper.Dropped() != reference.dropped {
				t.Fatalf("序列 %d 第 %d 帧：丢弃数 %d != 参考 %d", trial, index, stepper.Dropped(), reference.dropped)
			}
			if rest := stepper.Accumulator(0); rest != reference.accumulator {
				t.Fatalf("序列 %d 第 %d 帧：残留 %v != 参考 %v", trial, index, rest, reference.accumulator)
			}
		}
	}
}

func TestAdvanceDoesNotAllocate(t *testing.T) {
	stepper := &Stepper{}
	if allocs := testing.AllocsPerRun(1000, func() {
		stepper.Advance(16.6)
	}); allocs != 0 {
		t.Fatalf("单帧推进应 O(1) 内存零分配，实际 %v 次", allocs)
	}
}
