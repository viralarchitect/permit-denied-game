package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"permitdenied/internal/game"
)

type host struct {
	g           *game.Game
	out         string
	lines       []string
	i           int
	holdLeft    int
	holdTotal   int
	holdKeys    []string
	pendingShot string
	err         error
	done        bool
}

func (h *host) Layout(outsideWidth, outsideHeight int) (int, int) {
	return h.g.Layout(outsideWidth, outsideHeight)
}

func (h *host) Draw(screen *ebiten.Image) {
	h.g.Draw(screen)
	if h.pendingShot == "" {
		return
	}
	path := filepath.Join(h.out, filepath.FromSlash(h.pendingShot))
	if err := writeScreenPNG(screen, path); err != nil {
		h.err = err
	}
	h.pendingShot = ""
}

func (h *host) Update() error {
	if h.err != nil {
		return ebiten.Termination
	}
	if h.pendingShot != "" {
		return nil
	}
	if h.holdLeft > 0 {
		if err := h.stepHold(); err != nil {
			h.err = err
			return ebiten.Termination
		}
		return nil
	}
	for h.i < len(h.lines) {
		line := strings.TrimSpace(h.lines[h.i])
		h.i++
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		block, err := h.startCmd(line)
		if err != nil {
			h.err = fmt.Errorf("line %d: %w", h.i, err)
			return ebiten.Termination
		}
		if block {
			return nil
		}
	}
	h.done = true
	return ebiten.Termination
}

func (h *host) startCmd(line string) (block bool, err error) {
	cmd, rest, _ := strings.Cut(line, " ")
	rest = strings.TrimSpace(rest)
	switch cmd {
	case "tap":
		return true, h.beginHold(rest, 1)
	case "hold":
		keys, ticks, ok := strings.Cut(rest, " ")
		if !ok {
			return false, fmt.Errorf("hold KEYS TICKS")
		}
		n, err := strconv.Atoi(strings.TrimSpace(ticks))
		if err != nil || n < 1 {
			return false, fmt.Errorf("hold ticks")
		}
		return true, h.beginHold(strings.TrimSpace(keys), n)
	case "wait":
		n, err := strconv.Atoi(rest)
		if err != nil || n < 1 {
			return false, fmt.Errorf("wait TICKS")
		}
		return true, h.beginHold("", n)
	case "shot":
		if rest == "" {
			return false, fmt.Errorf("shot RELPATH")
		}
		h.pendingShot = rest
		return true, nil
	case "dump":
		if rest == "" {
			return false, fmt.Errorf("dump RELPATH")
		}
		return false, writeDump(h.g, filepath.Join(h.out, filepath.FromSlash(rest)))
	case "assert":
		return false, assertSnap(h.g.Snapshot(), rest)
	default:
		return false, fmt.Errorf("unknown command %s", cmd)
	}
}

func (h *host) beginHold(spec string, ticks int) error {
	var names []string
	if spec != "" {
		for _, p := range strings.Split(spec, ",") {
			p = strings.ToLower(strings.TrimSpace(p))
			if p == "" {
				continue
			}
			names = append(names, p)
		}
	}
	h.holdKeys = names
	h.holdTotal = ticks
	h.holdLeft = ticks
	return h.stepHold()
}

func (h *host) stepHold() error {
	edge := h.holdLeft == h.holdTotal
	in, keys, err := keysFor(h.holdKeys, edge)
	if err != nil {
		return err
	}
	if err := h.g.Drive(in, keys); err != nil {
		return err
	}
	h.holdLeft--
	return nil
}

func writeScreenPNG(screen *ebiten.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	w, ht := screen.Bounds().Dx(), screen.Bounds().Dy()
	pix := make([]byte, 4*w*ht)
	screen.ReadPixels(pix)
	nrgba := image.NewNRGBA(image.Rect(0, 0, w, ht))
	copy(nrgba.Pix, pix)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, nrgba)
}
