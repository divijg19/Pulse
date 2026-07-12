package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/divijg19/Pulse/internal/tui"
)

func main() {
	dir := "tmp/pulse-audit"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("resolving directory: %v", err)
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		log.Fatalf("creating directory %s: %v", absDir, err)
	}

	surfaces := tui.AllAuditSurfaces()
	written := 0

	for _, surface := range surfaces {
		for _, size := range tui.AuditSizes {
			path, err := tui.WriteAuditCapture(surface, size.W, size.H, absDir)
			if err != nil {
				log.Fatalf("capturing %s at %dx%d: %v", surface.Name, size.W, size.H, err)
			}
			fmt.Printf("  %s\n", path)
			written++
		}
	}

	fmt.Printf("\n%d captures written to %s\n", written, absDir)
	fmt.Println("View with: cat <file>.ansi  or  less -R <file>.ansi")
}
