package substep

import "testing"

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
