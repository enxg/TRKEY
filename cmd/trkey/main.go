package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/enxg/TRKEY/internal/emit/keylayout"
	"github.com/enxg/TRKEY/internal/emit/klc"
	"github.com/enxg/TRKEY/internal/emit/xkb"
	"github.com/enxg/TRKEY/internal/layout"
)

var versionRe = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func main() {
	usPath := flag.String("us", "./us.toml", "base layout")
	trPath := flag.String("tr", "./tr.toml", "altgr overlay")
	outDir := flag.String("out", "./out", "output directory")
	name := flag.String("name", "TRKEY (US)", "layout display name")
	company := flag.String("company", "Enes Genç", "company shown by the Windows installer")
	copyright := flag.String("copyright", "", "copyright shown by the Windows installer")
	version := flag.String("version", "0.0.0-dev", "release version, major.minor.patch with an optional -suffix")
	flag.Parse()

	if *copyright == "" {
		*copyright = fmt.Sprintf("(c) %d %s", time.Now().Year(), *company)
	}

	// platforms want a plain numeric version; drop any -suffix (0.0.0-dev -> 0.0.0)
	numVersion, _, _ := strings.Cut(*version, "-")
	if !versionRe.MatchString(numVersion) {
		panic(fmt.Sprintf("version %q must look like 1.2.3 or 1.2.3-suffix", *version))
	}

	us, err := layout.Load(*usPath)
	if err != nil {
		panic(err)
	}

	tr, err := layout.Load(*trPath)
	if err != nil {
		panic(err)
	}

	keys, err := layout.Overlay(us, tr)
	if err != nil {
		panic(err)
	}

	writeFile(filepath.Join(*outDir, "xkb", "symbols", "trkey"), func(w io.Writer) error {
		return xkb.Emit(w, keys, *name)
	})

	writeFile(filepath.Join(*outDir, "xkb", "rules", "evdev.xml"), func(w io.Writer) error {
		return xkb.Registry(w, *name)
	})

	bundle := filepath.Join(*outDir, "macos", "TRKEY.bundle", "Contents")

	writeFile(filepath.Join(bundle, "Info.plist"), func(w io.Writer) error {
		return keylayout.InfoPlist(w, *name, numVersion)
	})

	writeFile(filepath.Join(bundle, "Resources", *name+".keylayout"), func(w io.Writer) error {
		return keylayout.Emit(w, keys, *name)
	})

	writeFile(filepath.Join(bundle, "Resources", "English.lproj", "InfoPlist.strings"), func(w io.Writer) error {
		return keylayout.InfoPlistStrings(w, *name)
	})

	writeFile(filepath.Join(*outDir, "windows", "trkey.klc"), func(w io.Writer) error {
		return klc.Emit(w, keys, klc.Meta{
			ID:          "trkey",
			Description: *name,
			Company:     *company,
			Copyright:   *copyright,
			Version:     numVersion,
		})
	})
}

func writeFile(path string, emit func(io.Writer) error) {
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = emit(f)
	if err != nil {
		panic(err)
	}
}
