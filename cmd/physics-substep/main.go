// Command physics-substep 跑固定步长样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/substep"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("physics-substep", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "rate", "rate / fps / spike / idle")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "rate":
		stepper := &substep.Stepper{}
		ran := stepper.Advance(100)
		fmt.Fprintf(stdout, "steps=%d ran=%d rest=%.0f\n", stepper.Steps(), ran, stepper.Accumulator(100))
	case "fps":
		stepper := &substep.Stepper{}
		for index := 0; index < 10; index++ {
			stepper.Advance(10)
		}
		fmt.Fprintf(stdout, "steps=%d\n", stepper.Steps())
	case "spike":
		stepper := &substep.Stepper{}
		ran := stepper.Advance(5000)
		fmt.Fprintf(stdout, "ran=%d steps=%d dropped=%d\n", ran, stepper.Steps(), stepper.Dropped())
	case "idle":
		stepper := &substep.Stepper{}
		for index := 0; index < 5; index++ {
			stepper.Advance(10)
		}
		fmt.Fprintf(stdout, "steps=%d rest=%.0f\n", stepper.Steps(), stepper.Accumulator(50))
	default:
		fmt.Fprintln(stderr, "需要 --sample rate|fps|spike|idle")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
