// Package agires reads AGI v2 and v3 game resource directories (LOGDIR,
// PICDIR, VIEWDIR, SNDDIR, or their v3 combined equivalent) and the VOL
// files they point into.
package agires

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ResourceType pairs a resource directory file with the label used to
// describe the resources it indexes (e.g. "logdir" -> "logic").
type ResourceType struct {
	DirName string
	Label   string
}

// ResourceTypes lists the four AGI v2 resource directories, in the order
// they appear concatenated in an AGI v3 combined directory file.
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

// Directories holds the parsed resource entries for a game, keyed by
// ResourceType.Label, along with the detected AGI resource format version
// (2 or 3) needed to read the VOL files correctly.
type Directories struct {
	Version int
	Entries map[string][]Entry
}

// LoadDirectories detects whether a game uses the AGI v2 (four separate
// directory files) or v3 (single combined directory file, optionally
// compressed resources) format, and parses the resource directories
// accordingly.
func LoadDirectories(gameDir string) (*Directories, error) {
	if allV2DirsPresent(gameDir) {
		entries := map[string][]Entry{}
		for _, rt := range ResourceTypes {
			path, err := FindFileCaseInsensitive(gameDir, rt.DirName)
			if err != nil {
				return nil, err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			entries[rt.Label] = ParseDir(data)
		}
		return &Directories{Version: 2, Entries: entries}, nil
	}

	dirPath, err := findCombinedDirFile(gameDir)
	if err != nil {
		return nil, fmt.Errorf("no v2 directory files and no v3 combined dir file found: %w", err)
	}
	data, err := os.ReadFile(dirPath)
	if err != nil {
		return nil, err
	}
	entries, err := parseCombinedDir(data)
	if err != nil {
		return nil, err
	}
	return &Directories{Version: 3, Entries: entries}, nil
}

func allV2DirsPresent(gameDir string) bool {
	for _, rt := range ResourceTypes {
		if _, err := FindFileCaseInsensitive(gameDir, rt.DirName); err != nil {
			return false
		}
	}
	return true
}

// findCombinedDirFile locates the single v3 directory file, named
// "<GAMEID>DIR" (e.g. MHDIR, MH2DIR), by looking for the one file in
// gameDir whose name ends in "dir" case-insensitively.
func findCombinedDirFile(gameDir string) (string, error) {
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.Name()), "dir") {
			return filepath.Join(gameDir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no *dir file found in %s", gameDir)
}

// parseCombinedDir splits an AGI v3 combined directory file into its four
// sections using the 8-byte header of section offsets, then parses each
// section with the same entry format as v2.
func parseCombinedDir(data []byte) (map[string][]Entry, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("combined dir file too short: %d bytes", len(data))
	}
	offsets := make([]int, 4)
	for i := range offsets {
		offsets[i] = int(data[i*2]) | int(data[i*2+1])<<8
	}

	entries := map[string][]Entry{}
	for i, rt := range ResourceTypes {
		start := offsets[i]
		end := len(data)
		if i+1 < len(offsets) {
			end = offsets[i+1]
		}
		if start < 0 || end > len(data) || start > end {
			return nil, fmt.Errorf("bad section bounds for %s: [%d:%d] (len %d)", rt.Label, start, end, len(data))
		}
		entries[rt.Label] = ParseDir(data[start:end])
	}
	return entries, nil
}

// ParseDir decodes an AGI resource directory section (e.g. LOGDIR, or one
// section of a v3 combined dir) into its resource entries. Each entry is 3
// bytes: the high nibble of the first byte is the VOL number, the
// remaining 20 bits are the byte offset into that VOL file. A volume
// nibble of 0xF marks an unused/empty slot.
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

// resourceHeader describes a decoded VOL resource header, uniform across
// v2 (uncompressed only) and v3 (optionally compressed) formats.
type resourceHeader struct {
	uncompressedSize int
	compressedSize   int
	isPicture        bool
	dataOffset       int
}

// readResourceHeader reads the header for the given entry from its VOL
// file. AGI v2 headers are 5 bytes (signature, vol number, one size field,
// implying uncompressed data). AGI v3 headers are 7 bytes (signature, vol
// number/picture-flag byte, uncompressed size, compressed size); the
// resource is stored uncompressed when the two sizes match.
func readResourceHeader(gameDir string, entry Entry, version int) (resourceHeader, error) {
	volPath, err := FindVolFile(gameDir, entry.Volume)
	if err != nil {
		return resourceHeader{}, err
	}

	f, err := os.Open(volPath)
	if err != nil {
		return resourceHeader{}, err
	}
	defer f.Close()

	headerLen := 5
	if version == 3 {
		headerLen = 7
	}
	header := make([]byte, headerLen)
	if _, err := f.ReadAt(header, int64(entry.Offset)); err != nil {
		return resourceHeader{}, fmt.Errorf("reading header at offset %d: %w", entry.Offset, err)
	}

	if header[0] != 0x12 || header[1] != 0x34 {
		return resourceHeader{}, fmt.Errorf("bad signature at offset %d: %02x%02x", entry.Offset, header[0], header[1])
	}

	h := resourceHeader{dataOffset: entry.Offset + headerLen}
	if version == 3 {
		h.isPicture = header[2]&0x80 != 0
		h.uncompressedSize = int(header[3]) | int(header[4])<<8
		h.compressedSize = int(header[5]) | int(header[6])<<8
	} else {
		h.uncompressedSize = int(header[3]) | int(header[4])<<8
		h.compressedSize = h.uncompressedSize
	}
	return h, nil
}

// ResourceInfo describes how a resource is stored in its VOL file.
type ResourceInfo struct {
	UncompressedSize int
	CompressedSize   int
	// IsPicture reports the v3 header flag marking a resource that uses
	// picture nibble-packing rather than LZW. Always false for v2.
	IsPicture bool
	// Compressed reports whether the stored bytes differ from the
	// decompressed ones. Always false for v2, which has no compression.
	Compressed bool
}

// ReadInfo reports how the given resource is stored, without decoding it.
// This matters beyond bookkeeping: AGI v3 leaves a LOGIC resource's message
// text as plaintext when the resource is LZW-compressed, and XOR-encrypts it
// only when the resource is stored uncompressed (v2 always encrypts).
func ReadInfo(gameDir string, entry Entry, version int) (ResourceInfo, error) {
	h, err := readResourceHeader(gameDir, entry, version)
	if err != nil {
		return ResourceInfo{}, err
	}
	return ResourceInfo{
		UncompressedSize: h.uncompressedSize,
		CompressedSize:   h.compressedSize,
		IsPicture:        h.isPicture,
		Compressed:       h.compressedSize != h.uncompressedSize,
	}, nil
}

// ReadHeader opens the VOL file for the given entry and reads the resource
// header at its offset. It returns the declared uncompressed resource size
// and the offset immediately following the header (where the resource data
// begins). This assumes AGI v2 (uncompressed) headers; use ReadData for v3
// games, which handles compression transparently.
func ReadHeader(gameDir string, entry Entry) (size, dataOffset int, err error) {
	h, err := readResourceHeader(gameDir, entry, 2)
	if err != nil {
		return 0, 0, err
	}
	return h.uncompressedSize, h.dataOffset, nil
}

// ReadData returns the decompressed resource bytes for the given entry.
// version must be 2 or 3, matching the format reported by LoadDirectories.
func ReadData(gameDir string, entry Entry, version int) ([]byte, error) {
	h, err := readResourceHeader(gameDir, entry, version)
	if err != nil {
		return nil, err
	}

	volPath, err := FindVolFile(gameDir, entry.Volume)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(volPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	raw := make([]byte, h.compressedSize)
	if _, err := f.ReadAt(raw, int64(h.dataOffset)); err != nil {
		return nil, fmt.Errorf("reading %d bytes at offset %d: %w", h.compressedSize, h.dataOffset, err)
	}

	if h.compressedSize == h.uncompressedSize {
		return raw, nil
	}
	if h.isPicture {
		return expandPicture(raw, h.uncompressedSize), nil
	}
	return lzwExpand(raw, h.uncompressedSize), nil
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

// ReadRaw returns the unparsed bytes of the VOL file starting at the given
// entry's offset, running to the end of the file. Unlike ReadData, it makes
// no assumptions about the resource header being valid or present, so it
// still returns something to inspect for entries that fail ReadData's
// signature or bounds checks (e.g. a corrupt directory entry).
func ReadRaw(gameDir string, entry Entry) ([]byte, error) {
	volPath, err := FindVolFile(gameDir, entry.Volume)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(volPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if int64(entry.Offset) >= info.Size() {
		return nil, fmt.Errorf("offset %d beyond end of file (size %d)", entry.Offset, info.Size())
	}

	raw := make([]byte, info.Size()-int64(entry.Offset))
	if _, err := f.ReadAt(raw, int64(entry.Offset)); err != nil {
		return nil, err
	}
	return raw, nil
}

// FindVolFile locates the VOL file for the given volume number. AGI v2
// games name these "vol.N"; AGI v3 games prefix them with the game's short
// ID (e.g. "Mhvol.0", "mh2vol.3"), so this matches on the "vol.N" suffix
// rather than an exact name.
func FindVolFile(dir string, volNum int) (string, error) {
	suffix := "vol." + strconv.Itoa(volNum)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.Name()), suffix) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("no vol file for volume %d found in %s", volNum, dir)
}
