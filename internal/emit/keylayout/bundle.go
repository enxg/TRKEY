package keylayout

import (
	"fmt"
	"io"
)

const bundleID = "com.github.enxg.trkey"

// InfoPlist writes the bundle's Contents/Info.plist. name must match the .keylayout name.
func InfoPlist(w io.Writer, name string) error {
	_, err := fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>%[1]s</string>
	<key>CFBundleName</key>
	<string>TRKEY</string>
	<key>CFBundleVersion</key>
	<string>1.0</string>
	<key>KLInfo_%[2]s</key>
	<dict>
		<key>TISInputSourceID</key>
		<string>%[1]s.us</string>
		<key>TISIntendedLanguage</key>
		<string>tr</string>
	</dict>
</dict>
</plist>
`, bundleID, attr(name))
	return err
}

// InfoPlistStrings writes the bundle's English.lproj/InfoPlist.strings (display name).
func InfoPlistStrings(w io.Writer, name string) error {
	_, err := fmt.Fprintf(w, "\"%[1]s\" = \"%[1]s\";\n", name)
	return err
}
