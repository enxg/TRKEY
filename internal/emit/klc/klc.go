package klc

import (
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/enxg/TRKEY/internal/key"
	"github.com/enxg/TRKEY/internal/layout"
)

type Meta struct {
	ID          string // KBD name, also the DLL name; max 8 characters
	Description string
	Company     string
	Copyright   string
	Version     string // major.minor.patch; klc only keeps major.minor
}

// Emit writes an MSKLC .klc source file (UTF-16LE with BOM, CRLF).
func Emit(w io.Writer, keys layout.Keys, meta Meta) error {
	if len(meta.ID) == 0 || len(meta.ID) > 8 {
		return fmt.Errorf("klc: id %q must be 1-8 characters", meta.ID)
	}

	for _, s := range []string{meta.Description, meta.Company, meta.Copyright} {
		if strings.ContainsAny(s, "\"\n") {
			return fmt.Errorf("klc: %q can't contain quotes or newlines", s)
		}
	}

	var b strings.Builder

	fmt.Fprintf(&b, "KBD\t%s\t\"%s\"\n\n", meta.ID, meta.Description)
	fmt.Fprintf(&b, "COPYRIGHT\t\"%s\"\n\n", meta.Copyright)
	fmt.Fprintf(&b, "COMPANY\t\"%s\"\n\n", meta.Company)
	b.WriteString("LOCALENAME\t\"tr-TR\"\n\n")
	b.WriteString("LOCALEID\t\"0000041f\"\n\n")
	parts := strings.Split(meta.Version, ".")
	if len(parts) < 2 {
		return fmt.Errorf("klc: version %q must be major.minor[.patch]", meta.Version)
	}
	fmt.Fprintf(&b, "VERSION\t%s.%s\n\n", parts[0], parts[1])
	b.WriteString(shiftStates)

	b.WriteString("LAYOUT\t\t;an extra '@' at the end is a dead key\n\n")
	b.WriteString("//SC\tVK_\t\tCap\t0\t1\t2\t3\t6\t7\n")
	b.WriteString("//--\t----\t\t----\t----\t----\t----\t----\t----\t----\n\n")

	rows, err := layoutRows(keys)
	if err != nil {
		return err
	}
	b.WriteString(rows)
	b.WriteString("\n")

	b.WriteString(keyNames)
	fmt.Fprintf(&b, "DESCRIPTIONS\n\n0409\t%s\n\n", meta.Description)
	// keyed by UI language (0409 = English UI), value is the layout's language name
	b.WriteString("LANGUAGENAMES\n\n0409\tTurkish (Türkiye)\n\n")
	b.WriteString("ENDKBD\n")

	return writeUTF16(w, b.String())
}

func layoutRows(keys layout.Keys) (string, error) {
	type row struct {
		sc   int
		line string
	}

	rows := make([]row, 0, len(keys)+1)

	// numpad decimal, like MSKLC's own US layout
	rows = append(rows, row{0x53, "53\tDECIMAL\t\t0\t002e\t002e\t-1\t-1\t-1\t-1\n"})

	for iso, k := range keys {
		sc, err := strconv.Atoi(k.WinScancode)
		if err != nil {
			return "", fmt.Errorf("klc: %s: scancode %q: %w", iso, k.WinScancode, err)
		}

		line, err := keyLine(sc, k)
		if err != nil {
			return "", fmt.Errorf("klc: %s: %w", iso, err)
		}

		rows = append(rows, row{sc, line})
	}

	slices.SortFunc(rows, func(a, b row) int { return a.sc - b.sc })

	var b strings.Builder
	for _, r := range rows {
		b.WriteString(r.line)
	}

	return b.String(), nil
}

func keyLine(sc int, k key.Key) (string, error) {
	altgr := key.Layer{}
	if k.AltGr != nil {
		altgr = *k.AltGr
	}

	ctl := ctrlChars[k.ISO]
	chars := []string{k.Base.Normal, k.Base.Shift, ctl[0], ctl[1], altgr.Normal, altgr.Shift}

	cols := make([]string, len(chars))
	for i, c := range chars {
		s, err := char(c)
		if err != nil {
			return "", err
		}
		cols[i] = s
	}

	vk := k.WinVK
	if len(vk) < 8 {
		vk += "\t"
	}

	return fmt.Sprintf("%02x\t%s\t%d\t%s\n", sc, vk, capFlags(k), strings.Join(cols, "\t")), nil
}

// capFlags: 1 = CAPLOK (caps affects base), 4 = CAPLOKALTGR (caps affects altgr).
func capFlags(k key.Key) int {
	f := 0
	if k.Base.Caps {
		f |= 1
	}
	if k.AltGr != nil && k.AltGr.Caps {
		f |= 4
	}
	return f
}

// char formats a slot the way MSKLC does: letters and digits as-is, the rest as hex.
func char(s string) (string, error) {
	if s == "" {
		return "-1", nil
	}

	r, size := utf8.DecodeRuneInString(s)
	if size != len(s) {
		return "", fmt.Errorf("%q is not a single character", s)
	}

	if r > 0xffff {
		return "", fmt.Errorf("%U is outside the BMP, klc can't represent it", r)
	}

	if r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
		return string(r), nil
	}

	return fmt.Sprintf("%04x", r), nil
}

func writeUTF16(w io.Writer, s string) error {
	s = strings.ReplaceAll(s, "\n", "\r\n")

	u := utf16.Encode([]rune(s))
	buf := make([]byte, 2+2*len(u))
	buf[0], buf[1] = 0xff, 0xfe // BOM
	for i, c := range u {
		binary.LittleEndian.PutUint16(buf[2+2*i:], c)
	}

	_, err := w.Write(buf)
	return err
}

// ctrlChars are Windows US control characters for Ctrl and Shift+Ctrl.
var ctrlChars = map[string][2]string{
	"D11": {"\x1b", ""}, // Ctrl+[  ESC
	"D12": {"\x1d", ""}, // Ctrl+]  GS
	"C12": {"\x1c", ""}, // Ctrl+\  FS
	"B00": {"\x1c", ""}, // Ctrl+\  FS (ISO key)
	"A03": {" ", ""},    // Ctrl+Space
	"E02": {"", "\x00"}, // Ctrl+Shift+2  NUL
	"E06": {"", "\x1e"}, // Ctrl+Shift+6  RS
	"E11": {"", "\x1f"}, // Ctrl+Shift+-  US
}

const shiftStates = `SHIFTSTATE

0	//Column 4
1	//Column 5 : Shft
2	//Column 6 :       Ctrl
3	//Column 7 : Shft  Ctrl
6	//Column 8 :       Ctrl Alt
7	//Column 9 : Shft  Ctrl Alt

`

const keyNames = `KEYNAME

01	Esc
0e	Backspace
0f	Tab
1c	Enter
1d	Ctrl
2a	Shift
36	"Right Shift"
37	"Num *"
38	Alt
39	Space
3a	"Caps Lock"
3b	F1
3c	F2
3d	F3
3e	F4
3f	F5
40	F6
41	F7
42	F8
43	F9
44	F10
45	Pause
46	"Scroll Lock"
47	"Num 7"
48	"Num 8"
49	"Num 9"
4a	"Num -"
4b	"Num 4"
4c	"Num 5"
4d	"Num 6"
4e	"Num +"
4f	"Num 1"
50	"Num 2"
51	"Num 3"
52	"Num 0"
53	"Num Del"
54	"Sys Req"
57	F11
58	F12
7c	F13
7d	F14
7e	F15
7f	F16
80	F17
81	F18
82	F19
83	F20
84	F21
85	F22
86	F23
87	F24

KEYNAME_EXT

1c	"Num Enter"
1d	"Right Ctrl"
35	"Num /"
37	"Prnt Scrn"
38	"Right Alt"
45	"Num Lock"
46	Break
47	Home
48	Up
49	"Page Up"
4b	Left
4d	Right
4f	End
50	Down
51	"Page Down"
52	Insert
53	Delete
54	<00>
56	Help
5b	"Left Windows"
5c	"Right Windows"
5d	Application

`
