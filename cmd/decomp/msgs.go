package main

// agiMessageKey is the fixed XOR key AGI uses to obfuscate message text.
var agiMessageKey = []byte("Avis Durgan")

// parseMessages parses the message section of a logic resource (the bytes
// starting at msgPtr, as sliced off by decompile). Layout:
//
//	byte    numMessages
//	word    endOfText (offset to end of text block, LE, unused here)
//	word[N] offsets, LE, relative to byte 1 (the start of the endOfText
//	        word); 0 means the message is absent
//	[]byte  text, each message XOR-encrypted (key cycling continuously
//	        across the whole text block, phase keyed off distance from the
//	        start of the text block) and NUL-terminated
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
	textStart := tableStart + num*2
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
			phase := ((pos+j-textStart)%len(agiMessageKey) + len(agiMessageKey)) % len(agiMessageKey)
			c := in[pos+j] ^ agiMessageKey[phase]
			if c == 0 {
				break
			}
			buf = append(buf, c)
		}
		msgs[byte(i+1)] = string(buf)
		//fmt.Println(i+1, string(buf))
	}
	return msgs
}
