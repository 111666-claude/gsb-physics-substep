package main

import (
	"bytes"
	"testing"
)

func TestSamplesMatchReadmeContract(t *testing.T) {
	cases := map[string]string{
		"rate":  "steps=5 ran=5 rest=0\n",
		"fps":   "steps=5\n",
		"spike": "ran=5 steps=5 dropped=245\n",
		"idle":  "steps=2 rest=10\n",
	}
	for sample, want := range cases {
		var stdout, stderr bytes.Buffer
		if code := Run([]string{"--sample", sample}, &stdout, &stderr); code != 0 {
			t.Fatalf("sample %s 退出码 %d，stderr=%q", sample, code, stderr.String())
		}
		if stdout.String() != want {
			t.Fatalf("sample %s 输出 %q，README 口径 %q", sample, stdout.String(), want)
		}
	}
}
