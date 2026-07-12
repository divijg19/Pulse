package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	for _, arg := range []string{"version", "-v", "--version"} {
		t.Run(arg, func(t *testing.T) {
			old := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("pipe: %v", err)
			}
			os.Stdout = w

			runErr := run([]string{arg})

			w.Close()
			os.Stdout = old
			out, _ := io.ReadAll(r)

			if runErr != nil {
				t.Fatalf("run %q returned error: %v", arg, runErr)
			}
			if !strings.Contains(string(out), "pulse") {
				t.Fatalf("version output missing binary name: %q", string(out))
			}
		})
	}
}
