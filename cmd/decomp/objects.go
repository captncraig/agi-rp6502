package main

import "os"

// itemNames maps an inventory item number to its display name, as read from
// the game's OBJECT file. Used to annotate get/drop/put's argItem arguments.
var itemNames map[byte]string

// loadObjects parses an OBJECT file and populates itemNames. It should be
// called once per game, before decompiling that game's logics.
//
// The file is XOR-obfuscated with the same key as logic messages. Layout
// (all multi-byte values little-endian):
//
//	word       numItems*3
//	[numItems] { byte startingRoom; word textOffset }
//	[]byte     names, NUL-terminated, at absolute position textOffset+3
//
// Item 0's "startingRoom" byte is repurposed by the original Sierra
// interpreter to store the max-animated-objects count, not an actual room.
func loadObjects(path string) error {
	in, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	itemNames = map[byte]string{}
	if len(in) < 2 {
		return nil
	}
	key := agiMessageKey
	dec := make([]byte, len(in))
	for i, b := range in {
		dec[i] = b ^ key[i%len(key)]
	}

	numItems := (int(dec[0]) | int(dec[1])<<8) / 3
	for i := 0; i < numItems && i < 256; i++ {
		base := 2 + i*3
		if base+2 >= len(dec) {
			break
		}
		off := int(dec[base+1]) | int(dec[base+2])<<8
		pos := off + 3
		if pos < 0 || pos >= len(dec) {
			continue
		}
		end := pos
		for end < len(dec) && dec[end] != 0 {
			end++
		}
		itemNames[byte(i)] = string(dec[pos:end])
	}
	return nil
}
