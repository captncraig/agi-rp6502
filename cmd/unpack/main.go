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

	dirs, err := agires.LoadDirectories(dir)
	if err != nil {
		return err
	}

	for _, rt := range agires.ResourceTypes {
		if err := unpackResourceType(name, dir, outDir, rt, dirs); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %v\n", name, rt.Label, err)
		}
	}

	return nil
}

func unpackResourceType(gameName, gameDir, outDir string, rt agires.ResourceType, dirs *agires.Directories) error {
	typeDir := filepath.Join(outDir, rt.Label)
	if err := os.MkdirAll(typeDir, 0o755); err != nil {
		return err
	}

	for _, res := range dirs.Entries[rt.Label] {
		data, err := agires.ReadData(gameDir, res, dirs.Version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %s.%d: %v\n", gameName, rt.Label, res.Number, err)
			if raw, rawErr := agires.ReadRaw(gameDir, res); rawErr == nil {
				outPath := filepath.Join(typeDir, strconv.Itoa(res.Number)+".bin")
				if werr := os.WriteFile(outPath, raw, 0o644); werr != nil {
					fmt.Fprintf(os.Stderr, "  %s: %s.%d: writing raw fallback: %v\n", gameName, rt.Label, res.Number, werr)
				}
			}
			continue
		}

		// Normalize LOGIC message text to plaintext here so that decomp and
		// the runtime never have to care about AGI's XOR obfuscation. AGI v2
		// always encrypts it; v3 encrypts it only for resources stored
		// uncompressed, leaving LZW-compressed ones as plaintext already.
		if rt.Label == "logic" {
			encrypted := dirs.Version == 2
			if dirs.Version == 3 {
				info, err := agires.ReadInfo(gameDir, res, dirs.Version)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  %s: %s.%d: %v\n", gameName, rt.Label, res.Number, err)
					continue
				}
				encrypted = !info.Compressed
			}
			if encrypted {
				data = agires.DecryptLogicMessages(data)
			}
		}
		outPath := filepath.Join(typeDir, strconv.Itoa(res.Number)+".bin")
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}

		if rt.Label == "sound" {
			riaPath := filepath.Join(typeDir, strconv.Itoa(res.Number)+".ria.bin")
			riaData := agires.ConvertSoundToRia(data)
			if err := os.WriteFile(riaPath, riaData, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", riaPath, err)
			}
		}
	}

	return nil
}
