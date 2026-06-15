package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/divijg19/Pulse/internal/server"
	"github.com/divijg19/Pulse/internal/tui"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "tui" {
		return tui.Run()
	}

	switch args[0] {
	case "web":
		return runWeb(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runWeb(args []string) error {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		printUsage()
		return nil
	}

	addr, err := parseWebAddr(args)
	if err != nil {
		return err
	}

	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("load embedded static files: %w", err)
	}

	fmt.Printf("Pulse WebUI running at %s\n", displayURL(addr))
	err = server.ListenAndServe(server.Options{
		Addr:     addr,
		StaticFS: subFS,
	})
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func parseWebAddr(args []string) (string, error) {
	addr := server.DefaultAddr
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--addr":
			if i+1 >= len(args) {
				return "", errors.New("--addr requires a value")
			}
			addr = args[i+1]
			i++
		case strings.HasPrefix(arg, "--addr="):
			addr = strings.TrimPrefix(arg, "--addr=")
		default:
			return "", fmt.Errorf("unknown web option: %s", arg)
		}
	}
	if addr == "" {
		return "", errors.New("--addr cannot be empty")
	}
	return addr, nil
}

func displayURL(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	return "http://" + addr
}

func printUsage() {
	fmt.Print(`Pulse

Usage:
  pulse                 Start the canonical terminal UI
  pulse tui             Start the canonical terminal UI
  pulse web [--addr :8080]
                        Start the browser WebUI
  pulse help            Show this help
`)
}
