package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const (
	gamesDir  = "games"
	baseRom   = "build/agi.rp6502"
	outDir    = "build/games"
	bank0Addr = "$8000"
	rp6502py  = "tools/rp6502.py"
)

func main() {
	// Optional game-name arguments restrict this run to those games only,
	// for fast single-game iteration (see `make <game>` in the Makefile).
	only := os.Args[1:]

	entries, err := os.ReadDir(gamesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", gamesDir, err)
		os.Exit(1)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating %s: %v\n", outDir, err)
		os.Exit(1)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if len(only) > 0 && !slices.Contains(only, e.Name()) {
			continue
		}
		gameDir := filepath.Join(gamesDir, e.Name())
		if err := makeRom(e.Name(), gameDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", e.Name(), err)
		}
	}
}

func makeRom(name, dir string) error {
	bank0Path := filepath.Join(dir, "bank0.bin")
	dataPath := filepath.Join(dir, "data.bin")
	if _, err := os.Stat(bank0Path); err != nil {
		return nil
	}
	if _, err := os.Stat(dataPath); err != nil {
		return nil
	}

	fmt.Printf("=== %s ===\n", name)

	outPath := filepath.Join(outDir, name+".rp6502")

	// outPath differs from baseRom, so this only ever reads baseRom.
	if err := runRp6502("-a", bank0Addr, "-o", outPath, "create", bank0Path, baseRom); err != nil {
		return fmt.Errorf("embedding bank0: %w", err)
	}

	// rp6502.py reads all inputs (including outPath below) fully into memory
	// before opening -o for writing, so writing back over outPath here is safe.
	if err := runRp6502("-a", "data.bin", "-o", outPath, "create", dataPath, outPath); err != nil {
		return fmt.Errorf("embedding data: %w", err)
	}

	fmt.Printf("  -> %s\n", outPath)

	return nil
}

func runRp6502(args ...string) error {
	fullArgs := append([]string{rp6502py}, args...)
	fmt.Printf("  $ python3 %s\n", strings.Join(fullArgs, " "))
	cmd := exec.Command("python3", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
