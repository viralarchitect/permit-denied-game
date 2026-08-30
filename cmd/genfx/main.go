// Command genfx writes boom/spark frames into assets/usable/sprites.png
// and sprites.json, plus PCM WAVs under assets/usable/sfx/.
// Run from the repo root: go run ./cmd/genfx
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

const (
	spritesPNG  = "assets/usable/sprites.png"
	spritesJSON = "assets/usable/sprites.json"
	sfxDir      = "assets/usable/sfx"
	boomSize    = 48
	sparkSize   = 16
	boomY       = 208
	sampleRate  = 44100
)

var ramp = []color.RGBA{
	{0xFF, 0xFF, 0xFF, 0xFF},
	{0xF8, 0xE0, 0x70, 0xFF},
	{0xE0, 0x80, 0x30, 0xFF},
	{0xC0, 0x40, 0x40, 0xFF},
	{0x5A, 0x2A, 0x20, 0xFF},
	{0x2A, 0x2A, 0x28, 0xFF},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "genfx: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := stampSprites(); err != nil {
		return err
	}
	if err := os.MkdirAll(sfxDir, 0o755); err != nil {
		return err
	}
	if err := writeWAV(filepath.Join(sfxDir, "wreck.wav"), wreckPCM()); err != nil {
		return err
	}
	if err := writeWAV(filepath.Join(sfxDir, "peel.wav"), peelPCM()); err != nil {
		return err
	}
	if err := writeWAV(filepath.Join(sfxDir, "burst.wav"), burstPCM()); err != nil {
		return err
	}
	return nil
}

func stampSprites() error {
	f, err := os.Open(spritesPNG)
	if err != nil {
		return err
	}
	src, err := png.Decode(f)
	f.Close()
	if err != nil {
		return err
	}
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)

	for i := 0; i < 6; i++ {
		x0 := i * boomSize
		paintBoom(dst, x0, boomY, i)
	}
	sparkY := boomY
	sparkX0 := 6 * boomSize
	for i := 0; i < 4; i++ {
		paintSpark(dst, sparkX0+i*sparkSize, sparkY, i)
	}

	pngTmp, err := os.CreateTemp(filepath.Dir(spritesPNG), ".sprites-*.png")
	if err != nil {
		return err
	}
	pngName := pngTmp.Name()
	defer os.Remove(pngName)
	if err := png.Encode(pngTmp, dst); err != nil {
		pngTmp.Close()
		return err
	}
	if err := pngTmp.Close(); err != nil {
		return err
	}

	jsonTmp, err := os.CreateTemp(filepath.Dir(spritesJSON), ".sprites-*.json")
	if err != nil {
		return err
	}
	jsonName := jsonTmp.Name()
	defer os.Remove(jsonName)
	if err := patchJSON(jsonName); err != nil {
		jsonTmp.Close()
		return err
	}
	if err := jsonTmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(pngName, spritesPNG); err != nil {
		return err
	}
	return os.Rename(jsonName, spritesJSON)
}

func paintBoom(img *image.NRGBA, x0, y0, frame int) {
	cx, cy := 23.5, 23.5
	clearRect(img, x0, y0, boomSize, boomSize)
	if frame == 5 {
		return
	}
	for y := 0; y < boomSize; y++ {
		for x := 0; x < boomSize; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			r := math.Hypot(dx, dy)
			ang := math.Atan2(dy, dx)
			c, ok := boomPixel(frame, r, ang, x, y)
			if !ok {
				continue
			}
			img.SetNRGBA(x0+x, y0+y, c)
		}
	}
}

func boomPixel(frame int, r, ang float64, x, y int) (color.NRGBA, bool) {
	switch frame {
	case 0:
		if r < 10 {
			return nrgba(ramp[0]), true
		}
		if r < 14 {
			return nrgba(ramp[1]), true
		}
		if r < 16 && dither(x, y) {
			return nrgba(ramp[5]), true
		}
	case 1:
		if r < 8 {
			return nrgba(ramp[0]), true
		}
		if r < 16 {
			return nrgba(ramp[2]), true
		}
		if r < 20 {
			return nrgba(ramp[3]), true
		}
		if r < 22 {
			return nrgba(ramp[5]), true
		}
	case 2:
		petal := 1 + 0.35*math.Abs(math.Sin(ang*4))
		rr := r / petal
		if rr < 6 {
			return nrgba(ramp[1]), true
		}
		if rr < 14 {
			return nrgba(ramp[2]), true
		}
		if rr < 18 {
			return nrgba(ramp[3]), true
		}
		if rr < 20 {
			return nrgba(ramp[5]), true
		}
		if rr < 22 && dither(x, y) {
			return nrgba(ramp[4]), true
		}
	case 3:
		if r > 10 && r < 18 {
			if dither(x, y) {
				return nrgba(ramp[5]), true
			}
			return nrgba(ramp[4]), true
		}
		if r > 8 && r < 20 && (int(r)+x+y)%3 == 0 {
			return nrgba(ramp[5]), true
		}
	case 4:
		puffs := [][2]float64{{-8, -6}, {9, -4}, {-4, 9}, {7, 8}, {0, -11}}
		for _, p := range puffs {
			if math.Hypot(float64(x)-23.5-p[0], float64(y)-23.5-p[1]) < 3.2 {
				if dither(x, y) {
					return nrgba(ramp[5]), true
				}
				return nrgba(ramp[4]), true
			}
		}
	}
	return color.NRGBA{}, false
}

func paintSpark(img *image.NRGBA, x0, y0, frame int) {
	clearRect(img, x0, y0, sparkSize, sparkSize)
	cx, cy := 7.5, 7.5
	for y := 0; y < sparkSize; y++ {
		for x := 0; x < sparkSize; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			r := math.Hypot(dx, dy)
			c, ok := sparkPixel(frame, r, x, y)
			if !ok {
				continue
			}
			img.SetNRGBA(x0+x, y0+y, c)
		}
	}
}

func sparkPixel(frame int, r float64, x, y int) (color.NRGBA, bool) {
	switch frame {
	case 0:
		if r < 2.2 {
			return nrgba(ramp[0]), true
		}
		if r < 4 {
			return nrgba(ramp[1]), true
		}
	case 1:
		if r < 3 {
			return nrgba(ramp[2]), true
		}
		if r < 5 {
			return nrgba(ramp[3]), true
		}
	case 2:
		if r < 2 {
			return nrgba(ramp[2]), true
		}
		if math.Abs(float64(x)-7.5) < 1 && math.Abs(float64(y)-7.5) < 6 {
			return nrgba(ramp[1]), true
		}
		if math.Abs(float64(y)-7.5) < 1 && math.Abs(float64(x)-7.5) < 6 {
			return nrgba(ramp[1]), true
		}
	case 3:
		if r < 2.5 && dither(x, y) {
			return nrgba(ramp[5]), true
		}
	}
	return color.NRGBA{}, false
}

func clearRect(img *image.NRGBA, x0, y0, w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x0+x, y0+y, color.NRGBA{})
		}
	}
}

func nrgba(c color.RGBA) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

func dither(x, y int) bool {
	return (x+y*3)&1 == 0
}

func patchJSON(path string) error {
	b, err := os.ReadFile(spritesJSON)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	frames, _ := doc["frames"].(map[string]any)
	if frames == nil {
		frames = map[string]any{}
		doc["frames"] = frames
	}
	if _, ok := frames["boom_00"]; ok {
		return os.WriteFile(path, b, 0o644)
	}
	for i := 0; i < 6; i++ {
		name := fmt.Sprintf("boom_%02d", i)
		frames[name] = map[string]any{
			"x": i * boomSize, "y": boomY, "w": boomSize, "h": boomSize,
			"anchor": []int{boomSize / 2, boomSize / 2},
		}
	}
	sparkX0 := 6 * boomSize
	for i := 0; i < 4; i++ {
		name := fmt.Sprintf("spark_%02d", i)
		frames[name] = map[string]any{
			"x": sparkX0 + i*sparkSize, "y": boomY, "w": sparkSize, "h": sparkSize,
			"anchor": []int{sparkSize / 2, sparkSize / 2},
		}
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}

func wreckPCM() []int16 {
	n := sampleRate * 220 / 1000
	buf := make([]int16, n*2)
	seed := uint32(0xC0FFEE01)
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		seed = seed*1664525 + 1013904223
		noise := float64(int32(seed)>>16) / 32768.0
		click := 0.0
		if t < 0.008 {
			click = 0.95 * (1 - t/0.008)
			if int(t*2200)%2 == 0 {
				click = -click
			}
		}
		body := 0.0
		if t >= 0.006 && t < 0.11 {
			env := 1 - (t-0.006)/0.104
			body = (0.35*math.Sin(2*math.Pi*90*t) + 0.25*noise) * env
		}
		tail := 0.0
		if t >= 0.09 {
			env := 1 - (t-0.09)/0.13
			if env < 0 {
				env = 0
			}
			tail = 0.12 * noise * env * env
		}
		s := clamp16(click + body + tail)
		buf[i*2] = s
		buf[i*2+1] = s
	}
	return buf
}

func peelPCM() []int16 {
	n := sampleRate * 110 / 1000
	buf := make([]int16, n*2)
	seed := uint32(0xA5A5A5A5)
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		seed = seed*1664525 + 1013904223
		noise := float64(int32(seed)>>16) / 32768.0
		click := 0.0
		if t < 0.005 {
			click = 0.85 * (1 - t/0.005)
			if int(t*3400)%2 == 0 {
				click = -click
			}
		}
		body := 0.0
		if t >= 0.004 && t < 0.055 {
			env := 1 - (t-0.004)/0.051
			body = (0.28*math.Sin(2*math.Pi*420*t) + 0.18*noise) * env
		}
		tail := 0.0
		if t >= 0.04 {
			env := 1 - (t-0.04)/0.07
			if env < 0 {
				env = 0
			}
			tail = 0.1 * noise * env
		}
		s := clamp16(click + body + tail)
		buf[i*2] = s
		buf[i*2+1] = s
	}
	return buf
}

func burstPCM() []int16 {
	n := sampleRate * 180 / 1000
	buf := make([]int16, n*2)
	seed := uint32(0x51EDF00D)
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		seed = seed*1664525 + 1013904223
		noise := float64(int32(seed)>>16) / 32768.0
		click := 0.0
		if t < 0.007 {
			click = 0.9 * (1 - t/0.007)
			if int(t*1800)%2 == 0 {
				click = -click
			}
		}
		body := 0.0
		if t >= 0.005 && t < 0.09 {
			env := 1 - (t-0.005)/0.085
			body = (0.4*math.Sin(2*math.Pi*140*t) + 0.22*noise) * env
		}
		tail := 0.0
		if t >= 0.07 {
			env := 1 - (t-0.07)/0.11
			if env < 0 {
				env = 0
			}
			tail = 0.14 * noise * env * env
		}
		s := clamp16(click + body + tail)
		buf[i*2] = s
		buf[i*2+1] = s
	}
	return buf
}

func clamp16(v float64) int16 {
	if v > 1 {
		v = 1
	}
	if v < -1 {
		v = -1
	}
	return int16(v * 32767)
}

func writeWAV(path string, pcm []int16) error {
	dataBytes := len(pcm) * 2
	hdr := make([]byte, 44)
	copy(hdr[0:], []byte("RIFF"))
	binary.LittleEndian.PutUint32(hdr[4:], uint32(36+dataBytes))
	copy(hdr[8:], []byte("WAVE"))
	copy(hdr[12:], []byte("fmt "))
	binary.LittleEndian.PutUint32(hdr[16:], 16)
	binary.LittleEndian.PutUint16(hdr[20:], 1)
	binary.LittleEndian.PutUint16(hdr[22:], 2)
	binary.LittleEndian.PutUint32(hdr[24:], sampleRate)
	binary.LittleEndian.PutUint32(hdr[28:], sampleRate*2*2)
	binary.LittleEndian.PutUint16(hdr[32:], 4)
	binary.LittleEndian.PutUint16(hdr[34:], 16)
	copy(hdr[36:], []byte("data"))
	binary.LittleEndian.PutUint32(hdr[40:], uint32(dataBytes))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(hdr); err != nil {
		return err
	}
	return binary.Write(f, binary.LittleEndian, pcm)
}
