package main

import (
	"fmt"
	"os"
	"path/filepath"

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
		if err := scanGame(e.Name(), gameDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", e.Name(), err)
		}
	}
}

func scanGame(name, dir string) error {
	fmt.Printf("== %s ==\n", name)

	for _, rt := range agires.ResourceTypes {
		if err := scanResourceType(dir, rt); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", rt.Label, err)
		}
	}

	return nil
}

func scanResourceType(dir string, rt agires.ResourceType) error {
	dirPath, err := agires.FindFileCaseInsensitive(dir, rt.DirName)
	if err != nil {
		return err
	}

	dirData, err := os.ReadFile(dirPath)
	if err != nil {
		return err
	}

	resources := agires.ParseDir(dirData)

	for _, res := range resources {
		size, _, err := agires.ReadHeader(dir, res)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s.%d: %v\n", rt.Label, res.Number, err)
			continue
		}
		fmt.Printf("  %s.%d: %d bytes\n", rt.Label, res.Number, size)
	}

	return nil
}
