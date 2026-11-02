package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/enxg/TRKEY/internal/emit/keylayout"
	"github.com/enxg/TRKEY/internal/emit/klc"
	"github.com/enxg/TRKEY/internal/emit/xkb"
	"github.com/enxg/TRKEY/internal/layout"
)

func main() {
	usPath := flag.String("us", "./us.toml", "base layout")
	trPath := flag.String("tr", "./tr.toml", "altgr overlay")
	outDir := flag.String("out", "./out", "output directory")
	name := flag.String("name", "TRKEY (US)", "layout display name")
	company := flag.String("company", "Enes Genç", "company shown by the Windows installer")
	copyright := flag.String("copyright", "", "copyright shown by the Windows installer")
	flag.Parse()

	if *copyright == "" {
		*copyright = fmt.Sprintf("(c) %d %s", time.Now().Year(), *company)
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

	bundle := filepath.Join(*outDir, "macos", "TRKEY.bundle", "Contents")

	writeFile(filepath.Join(bundle, "Info.plist"), func(w io.Writer) error {
		return keylayout.InfoPlist(w, *name)
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
