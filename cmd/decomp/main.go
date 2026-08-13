package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const gamesDir = "games"

var messages map[byte]string
var totalLen int

// jumpTargets holds every address (base+idx) that a goto jumps to. It's
// populated by a throwaway first pass over the logic before the real one, so
// that by the time the real pass reaches a target address (whether the goto
// that referenced it appeared earlier or later in the byte stream) it knows
// to print a label there.
var jumpTargets = map[int]bool{}

func labelName(addr int) string {
	return fmt.Sprintf("l%d", addr)
}

func decompile(in []byte, name string) ([]byte, error) {
	fmt.Printf("%s: (%d)\n", name, len(in))
	// msgPtr is relative to the byte right after this 2-byte header
	msgPtr := uint16(in[0]) | uint16(in[1])<<8

	in = in[2:]
	msgs := in[msgPtr:]
	messages = parseMessages(msgs)
	in = in[:msgPtr]
	totalLen = len(in)

	jumpTargets = map[int]bool{}
	// Errors here are ignored: this pass only exists to collect jump targets
	// for the real pass below, and whatever targets it found before failing
	// are still useful. The real error (if any) comes from the second pass.
	decompileStmts(in, 0, io.Discard, 2)

	buf := bytes.Buffer{}
	err := decompileStmts(in, 0, &buf, 2)
	return buf.Bytes(), err
}
func writeLine(indent int, w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "%s", strings.Repeat("  ", indent))
	fmt.Fprintf(w, format, args...)
}

var knownVars = map[byte]string{
	0:  "CurrentRoom",
	1:  "PrevRoom",
	2:  "WhichEgoBorder",
	3:  "CurrentScore",
	4:  "BorderObj",
	5:  "WhichObjBorder",
	6:  "EgoDir",
	7:  "MaxScore",
	8:  "FreePages",
	9:  "UnknownWord",
	10: "CycleTime",
	11: "TimeSeconds",
	12: "TimeMinutes",
	13: "TimeHours",
	14: "TimeDays",
	15: "JoystickSensitivity",
	16: "EgoView",
	17: "ErrorCode",
	18: "ErrorCodeParam",
	19: "KeyPressed",
	20: "ComputerType",
	21: "MessageTimer",
	22: "SoundType",
	23: "SoundVolume",
	24: "InputBufferSize",
	25: "StatusItem",
	26: "MonitorType",
}

func varName(num byte) string {
	if knownVars[num] != "" {
		return fmt.Sprintf("v%s", knownVars[num])
	}
	return fmt.Sprintf("v%d", num)
}

func objName(num byte) string {
	return fmt.Sprintf("o%d", num)
}
func colorName(num byte) string {
	return fmt.Sprintf("c%d", num)
}

var knownFlags = map[byte]string{
	0: "EgoOnWater",
	1: "EgoObscured",
	2: "EnteredCommand",
	3: "TriggerLine",
	4: "SaidAccepted",
	5: "NewRoom",
	6: "RestartGame",
}

func flagName(num byte) string {
	if knownFlags[num] != "" {
		return fmt.Sprintf("f%s", knownFlags[num])
	}
	return fmt.Sprintf("f%d", num)
}

func ctlName(num byte) string {
	return fmt.Sprintf("c%d", num)
}

// PC XT scan codes for the keys AGI logic scripts commonly bind via set.key.
var scanCodeNames = map[byte]string{
	1:   "Esc",
	2:   "1",
	3:   "2",
	4:   "3",
	5:   "4",
	6:   "5",
	7:   "6",
	8:   "7",
	9:   "8",
	10:  "9",
	11:  "0",
	12:  "-",
	13:  "=",
	14:  "Backspace",
	15:  "Tab",
	16:  "Q",
	17:  "W",
	18:  "E",
	19:  "R",
	20:  "T",
	21:  "Y",
	22:  "U",
	23:  "I",
	24:  "O",
	25:  "P",
	26:  "[",
	27:  "]",
	28:  "Enter",
	29:  "Ctrl",
	30:  "A",
	31:  "S",
	32:  "D",
	33:  "F",
	34:  "G",
	35:  "H",
	36:  "J",
	37:  "K",
	38:  "L",
	39:  ";",
	40:  "'",
	41:  "`",
	42:  "LShift",
	43:  "\\",
	44:  "Z",
	45:  "X",
	46:  "C",
	47:  "V",
	48:  "B",
	49:  "N",
	50:  "M",
	51:  ",",
	52:  ".",
	53:  "/",
	54:  "RShift",
	56:  "Alt",
	57:  "Space",
	58:  "CapsLock",
	59:  "F1",
	60:  "F2",
	61:  "F3",
	62:  "F4",
	63:  "F5",
	64:  "F6",
	65:  "F7",
	66:  "F8",
	67:  "F9",
	68:  "F10",
	71:  "Home",
	72:  "Up",
	73:  "PgUp",
	75:  "Left",
	77:  "Right",
	79:  "End",
	80:  "Down",
	81:  "PgDn",
	82:  "Ins",
	83:  "Del",
	120: "Alt+1",
	121: "Alt+2",
	122: "Alt+3",
	123: "Alt+4",
	124: "Alt+5",
	125: "Alt+6",
	126: "Alt+7",
	127: "Alt+8",
	128: "Alt+9",
	129: "Alt+0",
	130: "Alt+-",
	131: "Alt+=",
}

// asciiNames covers control characters and other non-printable ASCII codes
// that show up as key bindings (e.g. Ctrl+S saves as ascii 0x13) but render
// unreadably via %q.
var asciiNames = map[byte]string{
	8:  "Backspace",
	9:  "Tab",
	13: "Enter",
	27: "Esc",
}

// keyName renders a set.key(ascii, scan, ...) pair as a single readable
// token. ascii and scan are mutually exclusive in practice: a binding is
// made by one or the other, with the unused half left as 0.
func keyName(ascii, scan byte) string {
	if ascii != 0 {
		if name, ok := asciiNames[ascii]; ok {
			return name
		}
		if ascii >= 1 && ascii <= 26 {
			return fmt.Sprintf("Ctrl+%c", 'A'+ascii-1)
		}
		return fmt.Sprintf("%q", string(rune(ascii)))
	}
	if name, ok := scanCodeNames[scan]; ok {
		return name
	}
	return fmt.Sprintf("scan%d", scan)
}

func msgName(num byte) string {
	if text, ok := messages[num]; ok {
		return fmt.Sprintf("%q", text)
	}
	return fmt.Sprintf("m%d", num)
}
func itemName(num byte) string {
	if name, ok := itemNames[num]; ok && name != "" {
		return fmt.Sprintf("%q", name)
	}
	return fmt.Sprintf("i%d", num)
}
func stringName(num byte) string {
	return fmt.Sprintf("str%d", num)
}
func wordNum(num uint16) string {
	if ws, ok := wordsByNum[int(num)]; ok {
		return fmt.Sprintf("%q", ws[0])
	}
	return fmt.Sprintf("w%d", num)
}

func decompileStmts(in []byte, indent int, w io.Writer, base int) error {
	idx := 0
	for idx < len(in) {
		if jumpTargets[idx+base] {
			fmt.Fprintf(w, "%s:\n", labelName(idx+base))
		}
		n, err := decompileStmt(in, idx, indent, w, base)
		if err != nil {
			if indent == 0 {
				fmt.Fprintf(w, "\n!!!\n%s\n", err)
			}
			return err
		}
		idx += n
	}
	return nil
}
func printFunction(indent int, w io.Writer, fDef fDef, get func() byte) {
	writeLine(indent, w, "%s(", fDef.name)
	args := []string{}
	for _, a := range fDef.args {
		v := ""
		switch a {
		case argNum:
			v = fmt.Sprint(get())
		case argFlag:
			v = flagName(get())
		case argVar:
			v = varName(get())
		case argCtrl:
			v = ctlName(get())
		case argMsg:
			v = msgName(get())
		case argObj:
			v = objName(get())
		case argItem:
			v = itemName(get())
		case argString:
			v = stringName(get())
		case argColor:
			v = colorName(get())
		}
		args = append(args, v)
	}
	fmt.Fprint(w, strings.Join(args, ","))
	fmt.Fprint(w, ")")
}
func decompileStmt(in []byte, idx int, indent int, w io.Writer, base int) (int, error) {
	count := 0
	get := func() byte {
		v := in[idx+count]
		count++
		return v
	}
	op := get()
	switch op {
	case 0x00:
		writeLine(indent, w, "return()\n")
	case 0x01:
		writeLine(indent, w, "%s++\n", varName(get()))
	case 0x02:
		writeLine(indent, w, "%s--\n", varName(get()))
	case 0x03:
		v := get()
		writeLine(indent, w, "%s = %s\n", varName(v), getConst(v, get()))
	case 0x04:
		writeLine(indent, w, "%s = %s\n", varName(get()), varName(get()))
	case 0x05:
		writeLine(indent, w, "%s += %d\n", varName(get()), get())
	case 0x06:
		writeLine(indent, w, "%s += %s\n", varName(get()), varName(get()))
	case 0x07:
		writeLine(indent, w, "%s -= %d\n", varName(get()), get())
	case 0x08:
		writeLine(indent, w, "%s -= %s\n", varName(get()), varName(get()))
	case 0xfe:
		offset := int16(uint16(get()) | uint16(get())<<8)
		target := base + idx + count + int(offset)
		jumpTargets[target] = true
		writeLine(indent, w, "goto(%s)\n", labelName(target))
	case 0xff:
		// IF
		addr := base + idx
		writeLine(indent, w, "if (")
		n, err := decompConditionList(in, idx+count, w)
		if err != nil {
			return 0, err
		}
		count += n
		writeLine(0, w, ") { //%04x\n", addr)
		size := int(get()) | int(get())<<8
		blockBase := base + idx + count
		block := make([]byte, 0, size)
		for len(block) < size {
			block = append(block, get())
		}
		// No else-block detection: a true-branch that ends in a forward goto
		// (an "if (cond) goto X;" with no else, or an if/else) is ambiguous
		// from local bytes alone, since both compile to identical bytecode.
		// Just decompile the block as-is and let goto/label markers show the
		// jump; else blocks show up as a goto followed by fallthrough code.
		if err = decompileStmts(block, indent+1, w, blockBase); err != nil {
			return count, err
		}
		writeLine(indent, w, "}\n")

	case 0x79:
		ascii, scan := get(), get()
		writeLine(indent, w, "set.key(%s, %s)\n", keyName(ascii, scan), ctlName(get()))

	default:
		if fDef, ok := functions[op]; ok {
			printFunction(indent, w, fDef, get)
			fmt.Fprintln(w, "")
		} else {
			fmt.Printf("%.2f%% done\n", float64(base+idx)*100/float64(totalLen))
			return 0, fmt.Errorf("unknown opcode: 0x%02x", op)
		}
	}
	return count, nil
}

var (
	constantMaps = map[byte]map[byte]string{
		// computer types
		20: {
			0:  "IBM_PC",
			4:  "ATARI_ST",
			5:  "AMIGA",
			7:  "APPLE_2_GS",
			20: "AMIGA_SQ1",
		},
		// sound types
		22: {
			0: "IBM_PC",
			3: "TANDY",
			8: "APPLE_2_GS",
		},
		// monitor types
		26: {
			0: "CGA",
			2: "HERCULES",
			3: "EGA",
		},
	}
)

func getConst(v byte, c byte) string {
	if cMap, ok := constantMaps[v]; ok {
		if v, ok := cMap[c]; ok {
			return v
		}
	}
	return fmt.Sprint(c)
}

// decompTest parses a single test opcode (optionally negated via 0xfd) and
// writes its rendering to w. It does not consume list structure (0xfc/0xff).
func decompTest(in []byte, idx int, w io.Writer) (int, error) {
	count := 0
	get := func() byte {
		v := in[idx+count]
		count++
		return v
	}
	code := get()
	switch code {
	case 0x01:
		v := get()
		writeLine(0, w, "%s == %s", varName(v), getConst(v, get()))
	case 0x02:
		v := get()
		writeLine(0, w, "%s == %s", varName(v), varName(get()))
	case 0x03:
		v := get()
		writeLine(0, w, "%s < %s", varName(v), getConst(v, get()))
	case 0x04:
		v := get()
		writeLine(0, w, "%s < %s", varName(v), varName(get()))
	case 0x05:
		v := get()
		writeLine(0, w, "%s > %s", varName(v), getConst(v, get()))
	case 0x06:
		v := get()
		writeLine(0, w, "%s > %s", varName(v), varName(get()))
	case 0x07:
		writeLine(0, w, "%s", flagName(get()))
	case 0x08:
		writeLine(0, w, "%s", varName(get()))
	case 0x0e:
		writeLine(0, w, "said(")
		nWords := get()
		args := []string{}
		for i := byte(0); i < nWords; i++ {
			num := uint16(get()) | uint16(get())<<8
			args = append(args, wordNum(num))
		}
		writeLine(0, w, "%s", strings.Join(args, ", "))
		writeLine(0, w, ")")
	case 0xfd:
		// special case: !(x == y) reads better as x != y
		switch in[idx+count] {
		case 0x01:
			count++
			v := get()
			writeLine(0, w, "%s != %s", varName(v), getConst(v, get()))
		case 0x02:
			count++
			writeLine(0, w, "%s != %s", varName(get()), varName(get()))
		case 0x07:
			count++
			writeLine(0, w, "!%s", flagName(get()))
		case 0x08:
			count++
			writeLine(0, w, "!%s", varName(get()))
		default:
			writeLine(0, w, "!(")
			n, err := decompTest(in, idx+count, w)
			if err != nil {
				return 0, err
			}
			count += n
			writeLine(0, w, ")")
		}
	default:
		if fDef, ok := testFunctions[code]; ok {
			printFunction(0, w, fDef, get)
		} else {
			return 0, fmt.Errorf("unknown condition: 0x%02x", code)
		}

	}
	return count, nil
}

// decompConditionList parses the top-level test list of an if-statement:
// a sequence of tests ANDed together, terminated by 0xff. A 0xfc marker
// starts (and another 0xfc ends) a group of tests that are ORed together.
func decompConditionList(in []byte, idx int, w io.Writer) (int, error) {
	count := 0
	var parts []string
	for {
		b := in[idx+count]
		if b == 0xff {
			count++
			break
		}
		if b == 0xfc {
			count++
			var buf bytes.Buffer
			n, err := decompOrGroup(in, idx+count, &buf)
			if err != nil {
				return 0, err
			}
			count += n
			parts = append(parts, "("+buf.String()+")")
			continue
		}
		var buf bytes.Buffer
		n, err := decompTest(in, idx+count, &buf)
		if err != nil {
			return 0, err
		}
		count += n
		parts = append(parts, buf.String())
	}
	writeLine(0, w, "%s", strings.Join(parts, " && "))
	return count, nil
}

// decompOrGroup parses tests ORed together until the closing 0xfc marker,
// which it consumes.
func decompOrGroup(in []byte, idx int, w io.Writer) (int, error) {
	count := 0
	var parts []string
	for {
		b := in[idx+count]
		if b == 0xfc {
			count++
			break
		}
		var buf bytes.Buffer
		n, err := decompTest(in, idx+count, &buf)
		if err != nil {
			return 0, err
		}
		count += n
		parts = append(parts, buf.String())
	}
	writeLine(0, w, "%s", strings.Join(parts, " || "))
	return count, nil
}

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
		if err := decompileGame(e.Name(), gameDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", e.Name(), err)
			return
		}
	}
}

func decompileGame(name, dir string) error {
	if err := loadWords(filepath.Join(dir, "WORDS.TOK")); err != nil {
		return fmt.Errorf("loading words: %w", err)
	}
	if err := loadObjects(filepath.Join(dir, "OBJECT")); err != nil {
		return fmt.Errorf("loading objects: %w", err)
	}

	logicDir := filepath.Join(dir, "unpacked", "logic")
	entries, err := os.ReadDir(logicDir)
	if err != nil {
		return err
	}

	outDir := filepath.Join(dir, "decomp")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".bin") {
			continue
		}
		num := strings.TrimSuffix(e.Name(), ".bin")
		if _, err := strconv.Atoi(num); err != nil {
			continue
		}

		inPath := filepath.Join(logicDir, e.Name())
		data, err := os.ReadFile(inPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %s: %v\n", name, e.Name(), err)
			continue
		}

		out, decompErr := decompile(data, inPath)

		outPath := filepath.Join(outDir, num+".txt")
		if len(out) > 0 {
			if err := os.WriteFile(outPath, out, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", outPath, err)
			}
		}

		if decompErr != nil {
			fmt.Fprintf(os.Stderr, "  %s: %s: %v\n", name, e.Name(), decompErr)
			return decompErr
		}

	}

	return nil
}
