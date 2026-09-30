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

// referenceStepper 是「按真实时间累积 + 子步上限」的参考实现，用于差分对拍。
type referenceStepper struct {
	accumulator float64
	steps       int
	dropped     int
}

func (r *referenceStepper) advance(frameMs float64) int {
	r.accumulator += frameMs
	substeps := int(r.accumulator / FixedStepMs)
	ran := substeps
	if ran > MaxSubsteps {
		ran = MaxSubsteps
	}
	r.dropped += substeps - ran
	r.accumulator -= float64(substeps) * FixedStepMs
	r.steps += ran
	return ran
}

func TestRealTimeAccumulation(t *testing.T) {
	stepper := &Stepper{}
	if ran := stepper.Advance(100); ran != 5 {
		t.Fatalf("100ms 应该推进 5 步，实际 %d", ran)
	}
	if stepper.Steps() != 5 || stepper.Dropped() != 0 || stepper.Accumulator(100) != 0 {
		t.Fatalf("100ms 口径不对：steps=%d dropped=%d rest=%v",
			stepper.Steps(), stepper.Dropped(), stepper.Accumulator(100))
	}
}

func TestFrameRateIndependence(t *testing.T) {
	cases := []struct {
		total float64
		split []float64
	}{
		{100, []float64{100}},
		{100, []float64{50, 50}},
		{100, []float64{25, 25, 25, 25}},
		{100, []float64{20, 20, 20, 20, 20}},
		{100, []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}},
		{100, []float64{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}},
		{100, []float64{12.5, 12.5, 12.5, 12.5, 12.5, 12.5, 12.5, 12.5}},
		{30, []float64{30}},
		{30, []float64{15, 15}},
		{30, []float64{10, 10, 10}},
		{30, []float64{7.5, 7.5, 7.5, 7.5}},
	}
	for index, testCase := range cases {
		stepper := &Stepper{}
		for _, frameMs := range testCase.split {
			stepper.Advance(frameMs)
		}
		reference := &referenceStepper{}
		reference.advance(testCase.total)
		if stepper.Steps() != reference.steps || stepper.Dropped() != reference.dropped ||
			stepper.Accumulator(0) != reference.accumulator {
			t.Fatalf("用例 %d：同样的真实时间切帧结果不同 split=%v got(steps=%d dropped=%d rest=%v) want(steps=%d dropped=%d rest=%v)",
				index, testCase.split, stepper.Steps(), stepper.Dropped(), stepper.Accumulator(0),
				reference.steps, reference.dropped, reference.accumulator)
		}
	}
}

func TestSpikeCapAndDropped(t *testing.T) {
	stepper := &Stepper{}
	if ran := stepper.Advance(5000); ran != MaxSubsteps {
		t.Fatalf("卡顿单帧最多 %d 步，实际 %d", MaxSubsteps, ran)
	}
	if stepper.Steps() != 5 || stepper.Dropped() != 245 {
		t.Fatalf("卡顿口径不对：steps=%d dropped=%d", stepper.Steps(), stepper.Dropped())
	}
	if rest := stepper.Accumulator(5000); rest != 0 {
		t.Fatalf("5000ms 是整步，残留应为 0，实际 %v", rest)
	}

	stepper = &Stepper{}
	stepper.Advance(1e9)
	if stepper.Dropped() != int(1e9/FixedStepMs)-MaxSubsteps {
		t.Fatalf("超长卡顿丢弃数不对：%d", stepper.Dropped())
	}
	if rest := stepper.Accumulator(1e9); rest < 0 || rest >= FixedStepMs {
		t.Fatalf("超长卡顿后残留必须有界：%v", rest)
	}
}

func TestResidueAccumulates(t *testing.T) {
	stepper := &Stepper{}
	for index := 0; index < 3; index++ {
		stepper.Advance(10)
	}
	if stepper.Steps() != 1 {
		t.Fatalf("30ms 只够 1 步，实际 %d", stepper.Steps())
	}
	if rest := stepper.Accumulator(30); rest != 10 {
		t.Fatalf("30ms 后应残留 10ms，实际 %v", rest)
	}
	stepper.Advance(10)
	if stepper.Steps() != 2 || stepper.Accumulator(40) != 0 {
		t.Fatalf("再凑 10ms 应补出一步：steps=%d rest=%v", stepper.Steps(), stepper.Accumulator(40))
	}
}

func TestInvariantsAndDeterminism(t *testing.T) {
	frames := []float64{0, 10, 10, 5000, 7.5, 100, 0, 3.25, 999, 20, 19.999, 20.001}
	run := func() (Stepper, []int) {
		stepper := &Stepper{}
		ranFrames := make([]int, len(frames))
		for index, frameMs := range frames {
			ranFrames[index] = stepper.Advance(frameMs)
			rest := stepper.Accumulator(0)
			if rest < 0 || rest >= FixedStepMs+MaxSubsteps*FixedStepMs {
				t.Fatalf("残留时间越界：%v", rest)
			}
			if stepper.Steps() < 0 || stepper.Dropped() < 0 {
				t.Fatal("计数不允许为负")
			}
		}
		return *stepper, ranFrames
	}
	first, ranFirst := run()
	second, ranSecond := run()
	if first != second {
		t.Fatal("同一串帧输入重复执行结果不同")
	}
	for index := range frames {
		if ranFirst[index] != ranSecond[index] {
			t.Fatalf("第 %d 帧子步数重复执行不一致", index)
		}
		if ranFirst[index] < 0 || ranFirst[index] > MaxSubsteps {
			t.Fatalf("第 %d 帧子步数越过上限：%d", index, ranFirst[index])
		}
	}

	stepsMonotonic := &Stepper{}
	previous := 0
	for range frames {
		stepsMonotonic.Advance(10)
		if stepsMonotonic.Steps() < previous {
			t.Fatal("步数必须单调不减")
		}
		previous = stepsMonotonic.Steps()
	}
}

func TestReferenceDifferential200(t *testing.T) {
	rng := rand.New(rand.NewSource(20240517))
	for sequence := 0; sequence < 200; sequence++ {
		stepper := &Stepper{}
		reference := &referenceStepper{}
		frameCount := 1 + rng.Intn(50)
		for frame := 0; frame < frameCount; frame++ {
			var frameMs float64
			switch rng.Intn(6) {
			case 0:
				frameMs = 0
			case 1:
				frameMs = 5000
			case 2:
				frameMs = float64(1+rng.Intn(20000)) * 1000
			case 3:
				frameMs = rng.Float64() * 40
			case 4:
				frameMs = float64(rng.Intn(60))
			default:
				frameMs = rng.Float64() * 250
			}
			gotRan := stepper.Advance(frameMs)
			wantRan := reference.advance(frameMs)
			if gotRan != wantRan ||
				stepper.Steps() != reference.steps ||
				stepper.Dropped() != reference.dropped ||
				stepper.Accumulator(0) != reference.accumulator {
				t.Fatalf("序列 %d 第 %d 帧 frameMs=%v 出现差异：got(ran=%d steps=%d dropped=%d rest=%v) want(ran=%d steps=%d dropped=%d rest=%v)",
					sequence, frame, frameMs,
					gotRan, stepper.Steps(), stepper.Dropped(), stepper.Accumulator(0),
					wantRan, reference.steps, reference.dropped, reference.accumulator)
			}
		}
	}
}
