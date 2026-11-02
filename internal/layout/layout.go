package layout

import (
	"fmt"
	"os"

	"github.com/enxg/TRKEY/internal/key"
	"github.com/pelletier/go-toml/v2"
)

type Keys map[string]key.Key

// Load reads a layout TOML file keyed by ISO position.
func Load(path string) (Keys, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var keys Keys
	if err := toml.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	for iso, k := range keys {
		k.ISO = iso
		keys[iso] = k
	}

	return keys, nil
}

// Overlay adds the AltGr layers from tr onto a copy of us.
// tr may only set altgr, and only on keys that exist in us.
func Overlay(us, tr Keys) (Keys, error) {
	out := make(Keys, len(us))
	for iso, k := range us {
		out[iso] = k
	}

	for iso, t := range tr {
		k, ok := out[iso]
		if !ok {
			return nil, fmt.Errorf("tr: %s is not a key in the base layout", iso)
		}

		if t.AltGr == nil {
			return nil, fmt.Errorf("tr: %s has no altgr layer", iso)
		}

		if t != (key.Key{ISO: iso, AltGr: t.AltGr}) {
			return nil, fmt.Errorf("tr: %s sets fields other than altgr", iso)
		}

		k.AltGr = t.AltGr
		out[iso] = k
	}

	return out, nil
}
