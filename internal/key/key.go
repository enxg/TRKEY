package key

type Key struct {
	ISO  string `toml:"-"`
	Name string `toml:"name"`

	WinScancode string `toml:"win_scancode"`
	WinVK       string `toml:"win_vk"`
	OSXKeycode  string `toml:"osx_keycode"`
	XKB         string `toml:"xkb"`

	/*Base        string
	Shift       string
	CapsAffects bool*/

	Base  Layer  `toml:"base"`
	AltGr *Layer `toml:"altgr,omitempty"`
}

type Layer struct {
	Normal string `toml:"normal"`
	Shift  string `toml:"shift"`
	Caps   bool   `toml:"caps"`
}
