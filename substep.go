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
// 按真实时间累积：accumulator 先加 frameMs，本帧子步数为 floor(accumulator/FixedStepMs)，
// 最多执行 MaxSubsteps 步；超出上限的整步丢弃并计入 Dropped，跑完减去已执行与已丢弃
// 的步长，残留（不足一个步长）留在 accumulator 中供下一帧继续累积。
func (s *Stepper) Advance(frameMs float64) int {
	s.accumulator += frameMs
	available := int(s.accumulator / FixedStepMs)
	if available < 0 {
		available = 0
	}
	ran := available
	if ran > MaxSubsteps {
		ran = MaxSubsteps
	}
	s.dropped += available - ran
	s.accumulator -= float64(available) * FixedStepMs
	if s.accumulator < 0 {
		s.accumulator = 0
	}
	s.steps += ran
	return ran
}

// Steps 是累计子步数。
func (s *Stepper) Steps() int { return s.steps }

// Dropped 是被单帧上限丢弃的子步数累计。
func (s *Stepper) Dropped() int { return s.dropped }

// Accumulator 是当前残留时间（不足一个 FixedStepMs），下一帧继续累积。
// nowMs 仅用于兼容调用方的时间戳签名，不参与计算。
func (s *Stepper) Accumulator(nowMs float64) float64 { return s.accumulator }
