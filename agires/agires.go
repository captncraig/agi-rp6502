// Package agires reads AGI v2 game resource directories (LOGDIR, PICDIR,
// VIEWDIR, SNDDIR) and the VOL files they point into.
package agires

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResourceType pairs a resource directory file with the label used to
// describe the resources it indexes (e.g. "logdir" -> "logic").
type ResourceType struct {
	DirName string
	Label   string
}

// ResourceTypes lists the four AGI v2 resource directories.
var ResourceTypes = []ResourceType{
	{"logdir", "logic"},
	{"picdir", "pic"},
	{"viewdir", "view"},
	{"snddir", "sound"},
}

// Entry is a single 3-byte directory entry describing where a resource
// lives: which VOL file, and the byte offset within it.
type Entry struct {
	Number int
	Volume int
	Offset int
}

// ParseDir decodes an AGI resource directory file (e.g. LOGDIR) into its
// resource entries. Each entry is 3 bytes: the high nibble of the first
// byte is the VOL number, the remaining 20 bits are the byte offset into
// that VOL file. A volume nibble of 0xF marks an unused/empty slot.
func ParseDir(data []byte) []Entry {
	var entries []Entry
	for i := 0; i+3 <= len(data); i += 3 {
		b0, b1, b2 := data[i], data[i+1], data[i+2]
		volume := int(b0 >> 4)
		if volume == 0x0F {
			continue
		}
		offset := int(b0&0x0F)<<16 | int(b1)<<8 | int(b2)
		entries = append(entries, Entry{
			Number: i / 3,
			Volume: volume,
			Offset: offset,
		})
	}
	return entries
}

// ReadHeader opens the VOL file for the given entry and reads the resource
// header at its offset. The header format is: 2-byte signature (0x12 0x34),
// 1-byte volume number, 2-byte little-endian length. It returns the
// declared resource size and the offset immediately following the header
// (where the resource data begins).
func ReadHeader(gameDir string, entry Entry) (size, dataOffset int, err error) {
	volPath, err := FindFileCaseInsensitive(gameDir, fmt.Sprintf("vol.%d", entry.Volume))
	if err != nil {
		return 0, 0, err
	}

	f, err := os.Open(volPath)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	header := make([]byte, 5)
	if _, err := f.ReadAt(header, int64(entry.Offset)); err != nil {
		return 0, 0, fmt.Errorf("reading header at offset %d: %w", entry.Offset, err)
	}

	if header[0] != 0x12 || header[1] != 0x34 {
		return 0, 0, fmt.Errorf("bad signature at offset %d: %02x%02x", entry.Offset, header[0], header[1])
	}

	size = int(header[3]) | int(header[4])<<8
	return size, entry.Offset + len(header), nil
}

// ReadData returns the raw resource bytes for the given entry.
func ReadData(gameDir string, entry Entry) ([]byte, error) {
	size, dataOffset, err := ReadHeader(gameDir, entry)
	if err != nil {
		return nil, err
	}

	volPath, err := FindFileCaseInsensitive(gameDir, fmt.Sprintf("vol.%d", entry.Volume))
	if err != nil {
		return nil, err
	}

	f, err := os.Open(volPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data := make([]byte, size)
	if _, err := f.ReadAt(data, int64(dataOffset)); err != nil {
		return nil, fmt.Errorf("reading %d bytes at offset %d: %w", size, dataOffset, err)
	}

	return data, nil
}

// FindFileCaseInsensitive looks for a file in dir whose name matches target
// case-insensitively, since AGI games ship with inconsistent capitalization
// (LOGDIR, Logdir, logdir, ...) depending on the original distribution.
func FindFileCaseInsensitive(dir, target string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(e.Name(), target) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("%s not found in %s", target, dir)
}
