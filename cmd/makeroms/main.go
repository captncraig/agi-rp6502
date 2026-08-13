package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/captncraig/agi-rp6502/pkg/agires"
)

const (
	gamesDir  = "games"
	baseRom   = "build/agi.rp6502"
	outDir    = "build/games"
	indexAddr = "index.bin"
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
	indexPath := filepath.Join(dir, "index.bin")
	dataPath := filepath.Join(dir, "data.bin")
	if _, err := os.Stat(indexPath); err != nil {
		return nil
	}
	if _, err := os.Stat(dataPath); err != nil {
		return nil
	}

	fmt.Printf("=== %s ===\n", name)

	outPath := filepath.Join(outDir, name+".rp6502")

	// outPath differs from baseRom, so this only ever reads baseRom.
	if err := runRp6502("-a", indexAddr, "-o", outPath, "create", indexPath, baseRom); err != nil {
		return fmt.Errorf("embedding index: %w", err)
	}

	// rp6502.py reads all inputs (including outPath below) fully into memory
	// before opening -o for writing, so writing back over outPath here is safe.
	if err := runRp6502("-a", "data.bin", "-o", outPath, "create", dataPath, outPath); err != nil {
		return fmt.Errorf("embedding data: %w", err)
	}

	if _, err := embedIfPresent(dir, outPath, "words.tok"); err != nil {
		return err
	}
	if _, err := embedIfPresent(dir, outPath, "object"); err != nil {
		return err
	}
	for _, rt := range agires.ResourceTypes {
		if _, err := embedIfPresent(dir, outPath, rt.DirName); err != nil {
			return err
		}
	}
	for vol := 0; ; vol++ {
		name := fmt.Sprintf("vol.%d", vol)
		found, err := embedIfPresent(dir, outPath, name)
		if err != nil {
			return err
		}
		if !found {
			break
		}
	}

	fmt.Printf("  -> %s\n", outPath)

	return nil
}

// embedIfPresent looks up name in dir case-insensitively and, if found,
// embeds it into outPath under its lowercase name. It reports whether the
// file was found.
func embedIfPresent(dir, outPath, name string) (bool, error) {
	path, err := agires.FindFileCaseInsensitive(dir, name)
	if err != nil {
		return false, nil
	}
	if err := runRp6502("-a", name, "-o", outPath, "create", path, outPath); err != nil {
		return false, fmt.Errorf("embedding %s: %w", name, err)
	}
	return true, nil
}

func runRp6502(args ...string) error {
	fullArgs := append([]string{rp6502py}, args...)
	fmt.Printf("  $ python3 %s\n", strings.Join(fullArgs, " "))
	cmd := exec.Command("python3", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
