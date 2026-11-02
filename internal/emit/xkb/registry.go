package xkb

import (
	"fmt"
	"io"
	"strings"
)

// Registry writes an evdev.xml rules registry listing the layout, so layout pickers can show it.
// install.sh merges its <layout> block into the user's ~/.config/xkb/rules/evdev.xml.
func Registry(w io.Writer, name string) error {
	_, err := fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE xkbConfigRegistry SYSTEM "xkb.dtd">
<xkbConfigRegistry version="1.1">
  <layoutList>
    <layout>
      <configItem>
        <name>trkey</name>
        <shortDescription>tk</shortDescription>
        <description>%s</description>
        <languageList>
          <iso639Id>tur</iso639Id>
        </languageList>
      </configItem>
    </layout>
  </layoutList>
</xkbConfigRegistry>
`, xmlText(name))
	return err
}

func xmlText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}
