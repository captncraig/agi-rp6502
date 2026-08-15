package main

// parseMessages parses the message section of a logic resource (the bytes
// starting at msgPtr, as sliced off by decompile). Layout:
//
//	byte    numMessages
//	word    endOfText (offset to end of text block, LE, unused here)
//	word[N] offsets, LE, relative to byte 1 (the start of the endOfText
//	        word); 0 means the message is absent
//	[]byte  text, NUL-terminated
//
// The text arrives as plaintext: AGI's "Avis Durgan" XOR obfuscation is
// undone by cmd/unpack, so nothing downstream of the unpacked resources
// needs to know about it.
//
// Messages are 1-indexed by AGI convention; the returned map is keyed that way.
func parseMessages(in []byte) map[byte]string {
	if len(in) == 0 {
		return nil
	}
	num := int(in[0])
	msgs := make(map[byte]string, num)
	if num == 0 {
		return msgs
	}
	const tableStart = 3
	for i := 0; i < num; i++ {
		off := int(in[tableStart+i*2]) | int(in[tableStart+i*2+1])<<8
		if off == 0 {
			continue
		}
		pos := 1 + off
		if pos < 0 || pos >= len(in) {
			continue
		}
		buf := []byte{}
		for j := 0; pos+j < len(in); j++ {
			c := in[pos+j]
			if c == 0 {
				break
			}
			buf = append(buf, c)
		}
		msgs[byte(i+1)] = string(buf)
	}
	return msgs
}
