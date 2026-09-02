// Command stamp-asset blits one master PNG into one reserved atlas/tileset rect.
// Run from the repo root:
//
//	go run ./cmd/stamp-asset -frame dozer_up_00
//	go run ./cmd/stamp-asset -tile 5
//	go run ./cmd/stamp-asset -reserve name=x,y,w,h
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	frame := flag.String("frame", "", "sprite frame name in sprites.json")
	tileStr := flag.String("tile", "", "tileset cell id")
	reserve := flag.String("reserve", "", "append frame: name=x,y,w,h")
	flag.Parse()

	paths := DefaultPaths()
	var err error
	switch {
	case *reserve != "":
		name, r, e := ParseReserve(*reserve)
		if e != nil {
			err = e
			break
		}
		err = ReserveFrame(paths, name, r)
	case *frame != "":
		err = StampFrame(paths, *frame)
	case *tileStr != "":
		var id int
		if _, e := fmt.Sscanf(*tileStr, "%d", &id); e != nil {
			err = fmt.Errorf("bad -tile %q", *tileStr)
			break
		}
		err = StampTile(paths, id)
	default:
		fmt.Fprintf(os.Stderr, "usage:\n  go run ./cmd/stamp-asset -frame NAME\n  go run ./cmd/stamp-asset -tile ID\n  go run ./cmd/stamp-asset -reserve name=x,y,w,h\n")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "stamp-asset: %v\n", err)
		os.Exit(1)
	}
}
