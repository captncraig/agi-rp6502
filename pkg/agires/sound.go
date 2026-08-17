package agires

// agiClockHz is the PCjr/Tandy sound chip's frequency constant: the 10-bit
// divisor stored in a note is divided into 1/32 of the 3.579545 MHz system
// clock to get Hz.
const agiClockHz = 111860.0

// psgFreqScale converts Hz to the RP6502 PSG's freq register units of 1/3 Hz.
const psgFreqScale = 3.0

// ConvertSoundToRia rewrites an AGI PCjr-format SOUND resource's frequency
// divisors into RP6502 PSG frequency registers (Hz * 3), leaving duration,
// attenuation, and note layout untouched. The fourth ("noise") voice doesn't
// encode a tone frequency divisor in its note bytes, so it's copied through
// unmodified rather than guessed at.
func ConvertSoundToRia(data []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)
	if len(data) < 8 {
		return out
	}

	for voice := range 3 {
		pos := int(data[voice*2]) | int(data[voice*2+1])<<8
		for pos >= 0 && pos+5 <= len(data) {
			if data[pos] == 0xFF && data[pos+1] == 0xFF {
				break
			}
			divisor := int(data[pos+2]&0x3F)<<4 | int(data[pos+3]&0x0F)
			if divisor > 0 {
				reg := min(int(agiClockHz*psgFreqScale/float64(divisor)+0.5), 0xFFFF)
				out[pos+2] = byte(reg)
				out[pos+3] = byte(reg >> 8)
			}
			pos += 5
		}
	}
	return out
}
