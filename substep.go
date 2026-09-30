// Package substep 是固定步长物理推进：按真实时间累积子步，单帧子步有上限。
// 缺陷：每帧只推进一步（忽略持续时间）、时间按帧计数、没有单帧子步上限、
// 被丢弃的时间不记账、残留累积不清理。
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
// 缺陷：每帧恒定一步，既不按 frameMs 累积，也不看 MaxSubsteps，丢弃的时间更没记账。
func (s *Stepper) Advance(frameMs float64) int {
	s.steps++
	return 1
}

// Steps 是累计子步数。
func (s *Stepper) Steps() int { return s.steps }

// Dropped 是被上限丢弃的子步数。
// 缺陷：从来没被累加过。
func (s *Stepper) Dropped() int { return s.dropped }

// Accumulator 是当前残留时间。缺陷：没有累加器，恒定返回 0。
func (s *Stepper) Accumulator(nowMs float64) float64 { return s.accumulator }
