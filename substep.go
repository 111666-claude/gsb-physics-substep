// Package substep 是固定步长物理推进：按真实时间累积子步，单帧子步有上限。
package substep

// FixedStepMs 是固定物理步长。
const FixedStepMs = 20.0

// MaxSubsteps 是单帧允许执行的最大子步数。
const MaxSubsteps = 5

// Stepper 是物理推进器。
type Stepper struct {
	accumulator float64
	steps       int
	dropped     int
}

// Advance 推进一帧，返回本帧执行的子步数。
//
// 口径：accumulator += frameMs，子步数 = floor(accumulator / FixedStepMs)，
// 最多 MaxSubsteps 步；跑完把 accumulator 减去已执行步长。超出上限的部分
// 丢弃并按子步个数累加进 Dropped，对应时间一并从 accumulator 移除，因此
// accumulator 始终小于一个步长，不会因卡顿无限增长。
func (s *Stepper) Advance(frameMs float64) int {
	s.accumulator += frameMs
	substeps := int(s.accumulator / FixedStepMs)
	ran := substeps
	if ran > MaxSubsteps {
		ran = MaxSubsteps
	}
	s.dropped += substeps - ran
	s.accumulator -= float64(substeps) * FixedStepMs
	s.steps += ran
	return ran
}

// Steps 是累计子步数，单调不减。
func (s *Stepper) Steps() int { return s.steps }

// Dropped 是被单帧上限丢弃的子步数，单调不减。
func (s *Stepper) Dropped() int { return s.dropped }

// Accumulator 是当前残留时间（尚未凑成一个子步的零头），恒小于 FixedStepMs。
func (s *Stepper) Accumulator(nowMs float64) float64 { return s.accumulator }
