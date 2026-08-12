package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	gamesDir  = "games"
	baseRom   = "build/agi.rp6502"
	outDir    = "build/games"
	indexAddr = "0x8000"
	rp6502py  = "tools/rp6502.py"
)

func main() {
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
		gameDir := filepath.Join(gamesDir, e.Name())
		if err := makeRom(e.Name(), gameDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", e.Name(), err)
		}
	}
}

func makeRom(name, dir string) error {
	indexPath := filepath.Join(dir, "index.bin")
	dataPath := filepath.Join(dir, "data.bin")
	if _, err := os.Stat(indexPath); err != nil {
		return nil
	}
	if _, err := os.Stat(dataPath); err != nil {
		return nil
	}

	fmt.Printf("=== %s ===\n", name)

	stepRom, err := os.CreateTemp("", "makeroms-*.rp6502")
	if err != nil {
		return err
	}
	stepPath := stepRom.Name()
	stepRom.Close()
	defer os.Remove(stepPath)

	if err := runRp6502("-a", indexAddr, "-o", stepPath, "create", indexPath, baseRom); err != nil {
		return fmt.Errorf("embedding index: %w", err)
	}

	outPath := filepath.Join(outDir, name+".rp6502")
	if err := runRp6502("-a", "data", "-o", outPath, "create", dataPath, stepPath); err != nil {
		return fmt.Errorf("embedding data: %w", err)
	}

	fmt.Printf("  -> %s\n", outPath)

	return nil
}

func runRp6502(args ...string) error {
	cmd := exec.Command("python3", append([]string{rp6502py}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
