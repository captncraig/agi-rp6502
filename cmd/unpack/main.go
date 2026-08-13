package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/captncraig/agi-rp6502/pkg/agires"
)

const gamesDir = "games"

func main() {
	entries, err := os.ReadDir(gamesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", gamesDir, err)
		os.Exit(1)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		gameDir := filepath.Join(gamesDir, e.Name())
		if err := unpackGame(e.Name(), gameDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", e.Name(), err)
		}
	}
}

func unpackGame(name, dir string) error {
	outDir := filepath.Join(dir, "unpacked")

	for _, rt := range agires.ResourceTypes {
		if err := unpackResourceType(dir, outDir, rt); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %v\n", name, rt.Label, err)
		}
	}

	return nil
}

func unpackResourceType(gameDir, outDir string, rt agires.ResourceType) error {
	dirPath, err := agires.FindFileCaseInsensitive(gameDir, rt.DirName)
	if err != nil {
		return err
	}

	dirData, err := os.ReadFile(dirPath)
	if err != nil {
		return err
	}

	typeDir := filepath.Join(outDir, rt.Label)
	if err := os.MkdirAll(typeDir, 0o755); err != nil {
		return err
	}

	for _, res := range agires.ParseDir(dirData) {
		data, err := agires.ReadData(gameDir, res)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s.%d: %v\n", rt.Label, res.Number, err)
			continue
		}
		outPath := filepath.Join(typeDir, strconv.Itoa(res.Number)+".bin")
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
	}

	return nil
}
