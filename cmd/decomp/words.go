package main

import "os"

// words maps a dictionary word to its word-group number (multiple words can
// share a group number as synonyms, e.g. "airlock" and "airlock door").
var words map[string]int

// wordsByNum is the reverse lookup, group number -> all words in that group.
// Used to annotate said() word-number arguments with their actual text.
var wordsByNum map[int][]string

// loadWords parses a WORDS.TOK file and populates words/wordsByNum. It should
// be called once per game, before decompiling that game's logics.
//
// Format: https://wiki.scummvm.org/index.php?title=AGI/Specifications/Data#Words
//
//	word[26]  offsets (BE), one per letter a-z, to that letter's first entry
//	entries   for each word: byte prefixLen (chars reused from previous word),
//	          then chars XORed with 0x7f, the last char additionally OR'd with
//	          0x80 to mark end-of-word, then a 2-byte (BE) word-group number
func loadWords(path string) error {
	in, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	words = make(map[string]int)
	wordsByNum = make(map[int][]string)

	pos := 52 // past the 26 letter offsets
	prev := ""
	for pos < len(in) {
		prefixLen := int(in[pos])
		pos++
		word := prev[:prefixLen]
		for pos < len(in) {
			b := in[pos]
			pos++
			last := b&0x80 != 0
			c := (b & 0x7f) ^ 0x7f
			word += string(rune(c))
			if last {
				break
			}
		}
		if pos+1 >= len(in) {
			break
		}
		num := int(in[pos])<<8 | int(in[pos+1])
		pos += 2

		words[word] = num
		wordsByNum[num] = append(wordsByNum[num], word)
		prev = word
	}
	return nil
}
