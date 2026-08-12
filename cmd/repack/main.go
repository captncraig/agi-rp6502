package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/captncraig/agi-rp6502/agires"
)

const gamesDir = "games"

// resourceSlots is the number of resource-number slots reserved per type in
// the index (AGI resource numbers are one byte, 0-255).
const resourceSlots = 256

// indexEntrySize is the width of one resource's index entry: a 3-byte
// little-endian offset into data.bin and a 2-byte little-endian size. An
// entry of all 0xFF marks a resource number as not present.
const indexEntrySize = 5

func main() {
	// Optional game-name arguments restrict repacking to those games only,
	// for fast single-game iteration (see `make <game>` in the Makefile).
	only := os.Args[1:]

	entries, err := os.ReadDir(gamesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", gamesDir, err)
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
		if err := repackGame(e.Name(), gameDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", e.Name(), err)
		}
	}
}

func repackGame(name, dir string) error {
	unpackedDir := filepath.Join(dir, "unpacked")
	if _, err := os.Stat(unpackedDir); err != nil {
		return nil
	}

	fmt.Printf("=== %s ===\n", name)

	var out []byte
	index := make([]byte, 0, resourceSlots*indexEntrySize*len(agires.ResourceTypes))

	for _, rt := range agires.ResourceTypes {
		typeDir := filepath.Join(unpackedDir, rt.Label)
		nums, err := listResourceNumbers(typeDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", rt.Label, err)
		}

		type location struct{ offset, size int }
		byNumber := make(map[int]location, len(nums)) // resource number -> location in out
		for _, n := range nums {
			path := filepath.Join(typeDir, strconv.Itoa(n)+".bin")
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			if len(data) > 0xFFFF {
				return fmt.Errorf("%s.%d: resource too large for 16-bit size: %d bytes", rt.Label, n, len(data))
			}
			fmt.Printf("  %-6s %3d  offset=%-8d size=%d\n", rt.Label, n, len(out), len(data))
			byNumber[n] = location{offset: len(out), size: len(data)}
			out = append(out, data...)
		}

		for n := range resourceSlots {
			loc, ok := byNumber[n]
			if !ok {
				index = append(index, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF)
				continue
			}
			index = append(index,
				byte(loc.offset), byte(loc.offset>>8), byte(loc.offset>>16),
				byte(loc.size), byte(loc.size>>8),
			)
		}
	}

	outPath := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}
	fmt.Printf("  -> %s (%d bytes)\n", outPath, len(out))

	indexPath := filepath.Join(dir, "index.bin")
	if err := os.WriteFile(indexPath, index, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", indexPath, err)
	}
	fmt.Printf("  -> %s (%d bytes)\n", indexPath, len(index))

	return nil
}

func listResourceNumbers(dir string) ([]int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var nums []int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".bin") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSuffix(name, ".bin"))
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}
	sort.Ints(nums)
	return nums, nil
}
