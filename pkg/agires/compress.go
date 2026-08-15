package agires

// AGI v3 introduces two independent compression schemes: an adaptive LZW
// used for LOGIC, VIEW, and SOUND resources, and a much simpler nibble-packing
// scheme used only for PICTURE resources.

// bitReader pulls LSB-first, variable-width codes out of a byte slice.
type bitReader struct {
	data []byte
	pos  int // bit offset
}

func (r *bitReader) readCode(bits int) (int, bool) {
	if r.pos+bits > len(r.data)*8 {
		return 0, false
	}
	code := 0
	for i := 0; i < bits; i++ {
		byteIdx := (r.pos + i) / 8
		bitIdx := uint((r.pos + i) % 8)
		bit := (r.data[byteIdx] >> bitIdx) & 1
		code |= int(bit) << uint(i)
	}
	r.pos += bits
	return code, true
}

const (
	lzwResetCode = 256
	lzwEndCode   = 257
	// lzwTableSize oversizes the dictionary arrays relative to the 4096
	// codes reachable in a well-formed 12-bit stream, matching the
	// reference decoder's own generously-sized buffer — dictNext can spill
	// a little past 4096 on malformed input without indexing out of range.
	lzwTableSize = 18041
	lzwStartBits = 9
	lzwMaxBits   = 12
)

// lzwExpand decompresses an AGI v3 LZW-compressed resource, matching Lance
// Ewing's reference decoder (used in ScummVM's AGI engine) bit for bit: the
// dictionary index counter starts at 257 (its first slot, 257, is written
// but never referenced — the first genuinely reachable entry lands at 258,
// as the format spec describes), the very first code read primes the
// decoder but is never itself emitted, and the code width widens one slot
// early, when the about-to-be-assigned index exceeds (1<<bits)-2.
func lzwExpand(data []byte, expectedLen int) []byte {
	br := &bitReader{data: data}
	prefix := make([]int, lzwTableSize)
	suffix := make([]byte, lzwTableSize)

	bits := lzwStartBits
	maxCode := (1 << uint(bits)) - 2 // widen one slot before the dictionary is actually full

	readCode := func() (int, bool) {
		return br.readCode(bits)
	}

	// A corrupt stream can produce a circular prefix chain, so bail out
	// rather than spinning forever, as the reference decoder does.
	decodeString := func(code int) []byte {
		var stack []byte
		for i := 0; code > 255; i++ {
			if i >= 4000 {
				return nil
			}
			stack = append(stack, suffix[code])
			code = prefix[code]
		}
		stack = append(stack, byte(code))
		for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
			stack[i], stack[j] = stack[j], stack[i]
		}
		return stack
	}

	out := make([]byte, 0, expectedLen)

	// dictNext is the next dictionary slot to define; it starts at 257
	// (that first slot is written but never read back — the format's
	// first genuinely reachable entry lands at 258, matching the spec).
	dictNext := 257

	oldCode, ok := readCode()
	if !ok {
		return out
	}
	c := byte(oldCode) // primes the decoder; never itself emitted
	newCode, ok := readCode()
	if !ok {
		return out
	}

	for len(out) < expectedLen && newCode != lzwEndCode {
		if newCode == lzwResetCode {
			dictNext = 258
			bits = lzwStartBits
			maxCode = (1 << uint(bits)) - 2
			oc, ok := readCode()
			if !ok {
				break
			}
			oldCode = oc
			c = byte(oc)
			out = append(out, c)
			nc, ok := readCode()
			if !ok {
				break
			}
			newCode = nc
			continue
		}

		var entry []byte
		if newCode >= dictNext {
			prev := decodeString(oldCode)
			if prev == nil {
				break
			}
			entry = append(prev, c)
		} else {
			entry = decodeString(newCode)
			if entry == nil {
				break
			}
		}
		c = entry[0]
		out = append(out, entry...)

		// The reference decoder's setBits() refuses a request for MAXBITS
		// outright, so the code width saturates at 11 and never reaches 12;
		// past that point the dictionary keeps growing but codes stay
		// 11 bits wide. Long resources decode to garbage if we widen here.
		if dictNext > maxCode && bits+1 < lzwMaxBits {
			bits++
			maxCode = (1 << uint(bits)) - 2
		}
		if dictNext >= lzwTableSize {
			break
		}
		prefix[dictNext] = oldCode
		suffix[dictNext] = c
		dictNext++

		oldCode = newCode
		nc, ok := readCode()
		if !ok {
			break
		}
		newCode = nc
	}

	if len(out) > expectedLen {
		out = out[:expectedLen]
	}
	return out
}

// logicMessageKey is the fixed XOR key AGI uses to obfuscate the message
// text inside a LOGIC resource.
var logicMessageKey = []byte("Avis Durgan")

// DecryptLogicMessages returns data with the message text of a LOGIC
// resource decrypted in place, so that everything downstream can treat
// message text as plaintext regardless of the game's AGI version.
//
// The resource layout is a 2-byte offset to the message section (relative
// to the byte after it), then within that section a 1-byte message count, a
// 2-byte end-of-text offset, a table of 2-byte message offsets, and finally
// the text block. Only the text block is encrypted, with the key cycling
// continuously from the start of that block.
//
// Malformed or truncated input is returned unchanged rather than mangled.
func DecryptLogicMessages(data []byte) []byte {
	if len(data) < 2 {
		return data
	}
	msgStart := 2 + (int(data[0]) | int(data[1])<<8)
	if msgStart < 2 || msgStart >= len(data) {
		return data
	}
	sec := data[msgStart:]
	if len(sec) < 3 {
		return data
	}
	textStart := 3 + int(sec[0])*2
	if textStart >= len(sec) {
		return data
	}

	out := make([]byte, len(data))
	copy(out, data)
	text := out[msgStart+textStart:]
	for i := range text {
		text[i] ^= logicMessageKey[i%len(logicMessageKey)]
	}
	return out
}

// expandPicture reverses AGI v3's PICTURE compression: every occurrence of
// the set-visual-color (0xF0) or set-priority-color (0xF2) opcode is
// followed by only a 4-bit color nibble instead of a full byte, since
// there are only 16 colors. Everything else passes through unchanged, one
// byte at a time, at whatever nibble alignment the stream is currently at.
func expandPicture(data []byte, expectedLen int) []byte {
	totalNibbles := len(data) * 2

	nibble := func(idx int) byte {
		b := data[idx/2]
		if idx%2 == 0 {
			return b >> 4
		}
		return b & 0x0F
	}

	out := make([]byte, 0, expectedLen)
	cursor := 0
	for len(out) < expectedLen && cursor+1 < totalNibbles {
		b := nibble(cursor)<<4 | nibble(cursor+1)
		out = append(out, b)
		cursor += 2
		if b == 0xF0 || b == 0xF2 {
			if cursor >= totalNibbles {
				break
			}
			out = append(out, nibble(cursor))
			cursor++
		}
	}

	if len(out) > expectedLen {
		out = out[:expectedLen]
	}
	return out
}
