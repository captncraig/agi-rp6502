package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/captncraig/agi-rp6502/pkg/agires"
)

const gamesDir = "games"

// resourceSlots is the number of resource-number slots reserved per type in
// the index (AGI resource numbers are one byte, 0-255).
const resourceSlots = 256

// indexEntrySize is the width of one resource's index entry: a 4-byte
// little-endian offset into data.bin, a 2-byte little-endian size, a
// loaded-bank-number byte (0 meaning unloaded), and 1 byte of padding. An
// entry with 0xFF offset/size marks a resource number as not present.
const indexEntrySize = 8

// wordsSectionSize is the space reserved for words.tok within bank0.bin;
// the file is zero-padded out to this length.
const wordsSectionSize = 7 * 1024

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
				index = append(index, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0, 0)
				continue
			}
			index = append(index,
				byte(loc.offset), byte(loc.offset>>8), byte(loc.offset>>16), byte(loc.offset>>24),
				byte(loc.size), byte(loc.size>>8),
				0, 0,
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

	if err := writeBank0(dir, index); err != nil {
		return err
	}

	return nil
}

// writeBank0 assembles bank0.bin: the resource index, followed by words.tok
// zero-padded to wordsSectionSize, followed by the OBJECT file.
func writeBank0(dir string, index []byte) error {
	wordsPath, err := findCaseInsensitive(dir, "words.tok")
	if err != nil {
		return err
	}
	words, err := os.ReadFile(wordsPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", wordsPath, err)
	}
	if len(words) > wordsSectionSize {
		return fmt.Errorf("%s: %d bytes exceeds words section size %d", wordsPath, len(words), wordsSectionSize)
	}

	objectPath, err := findCaseInsensitive(dir, "object")
	if err != nil {
		return err
	}
	object, err := os.ReadFile(objectPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", objectPath, err)
	}

	bank0 := make([]byte, 0, len(index)+wordsSectionSize+len(object))
	bank0 = append(bank0, index...)
	bank0 = append(bank0, words...)
	bank0 = append(bank0, make([]byte, wordsSectionSize-len(words))...)
	bank0 = append(bank0, object...)

	bank0Path := filepath.Join(dir, "bank0.bin")
	if err := os.WriteFile(bank0Path, bank0, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", bank0Path, err)
	}
	fmt.Printf("  -> %s (%d bytes)\n", bank0Path, len(bank0))

	return nil
}

// findCaseInsensitive looks for a file named name (case-insensitively) in
// dir, since AGI games vary in how they case WORDS.TOK/OBJECT.
func findCaseInsensitive(dir, name string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", dir, err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("%s: no file matching %q", dir, name)
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
