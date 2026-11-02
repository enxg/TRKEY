package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/enxg/TRKEY/internal/emit/xkb"
	"github.com/enxg/TRKEY/internal/layout"
)

func main() {
	usPath := flag.String("us", "./us.toml", "base layout")
	trPath := flag.String("tr", "./tr.toml", "altgr overlay")
	outDir := flag.String("out", "./out", "output directory")
	name := flag.String("name", "TRKEY (US)", "layout display name")
	flag.Parse()

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

	writeXKB(keys, filepath.Join(*outDir, "xkb", "symbols", "trkey"), *name)
}

func writeXKB(keys layout.Keys, path, name string) {
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = xkb.Emit(f, keys, name)
	if err != nil {
		panic(err)
	}
}
