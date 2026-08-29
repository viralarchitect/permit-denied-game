package audio

import (
	"bytes"
	"math"
	"sync"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

// Audio is two channels: chase loop + crunch one-shot. Milestone 6, not a blocker.
type Audio struct {
	mu      sync.Mutex
	ctx     *ebitenaudio.Context
	chase   *ebitenaudio.Player
	crunchB []byte
	ready   bool
}

func (a *Audio) ensure() {
	if a == nil || a.ready {
		return
	}
	defer func() {
		_ = recover()
		a.ready = true
	}()
	ctx := ebitenaudio.CurrentContext()
	if ctx == nil {
		ctx = ebitenaudio.NewContext(44100)
	}
	a.ctx = ctx
	a.crunchB = genCrunch(44100)
	loop := genChase(44100)
	src := ebitenaudio.NewInfiniteLoop(bytes.NewReader(loop), int64(len(loop)))
	p, err := a.ctx.NewPlayer(src)
	if err == nil {
		p.SetVolume(0.35)
		a.chase = p
	}
}

func (a *Audio) StartChase() {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensure()
	if a.chase != nil {
		a.chase.SetVolume(0.35)
		_ = a.chase.Rewind()
		a.chase.Play()
	}
}

func (a *Audio) Duck(tally bool) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.chase == nil {
		return
	}
	if tally {
		a.chase.SetVolume(0.12)
	} else {
		a.chase.SetVolume(0.35)
	}
}

func (a *Audio) Crunch() {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensure()
	if a.ctx == nil || len(a.crunchB) == 0 {
		return
	}
	p, err := a.ctx.NewPlayer(bytes.NewReader(a.crunchB))
	if err != nil {
		return
	}
	p.SetVolume(0.55)
	p.Play()
}

func (a *Audio) Stop() {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.chase != nil {
		a.chase.Pause()
	}
}

func genChase(rate int) []byte {
	// 8-bar square/triangle, ~140 BPM, 16-bit stereo LE.
	const bpm = 140.0
	beats := 32.0 // 8 bars of 4/4
	sec := beats * 60.0 / bpm
	n := int(float64(rate) * sec)
	notes := []float64{110, 110, 146.8, 110, 130.8, 110, 98, 146.8}
	buf := make([]byte, n*4)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(rate)
		beat := t * bpm / 60.0
		ni := int(beat) % len(notes)
		freq := notes[ni]
		phase := t * freq
		sq := 0.12
		if math.Mod(phase, 1) < 0.5 {
			sq = -0.12
		}
		tri := 0.08 * (2*math.Abs(2*(phase-math.Floor(phase+0.5))) - 1)
		// kick-ish click on beats
		frac := beat - math.Floor(beat)
		kick := 0.0
		if frac < 0.08 {
			kick = -0.18 * (1 - frac/0.08)
		}
		v := sq + tri + kick
		s := int16(clamp32(v) * 32767)
		buf[i*4] = byte(s)
		buf[i*4+1] = byte(s >> 8)
		buf[i*4+2] = byte(s)
		buf[i*4+3] = byte(s >> 8)
	}
	return buf
}

func genCrunch(rate int) []byte {
	n := rate * 80 / 1000
	buf := make([]byte, n*4)
	seed := uint32(0xA341316C)
	for i := 0; i < n; i++ {
		seed = seed*1664525 + 1013904223
		noise := float64(int32(seed)>>16) / 32768.0
		env := 1 - float64(i)/float64(n)
		v := noise * 0.55 * env * env
		s := int16(clamp32(v) * 32767)
		buf[i*4] = byte(s)
		buf[i*4+1] = byte(s >> 8)
		buf[i*4+2] = byte(s)
		buf[i*4+3] = byte(s >> 8)
	}
	return buf
}

func clamp32(v float64) float64 {
	if v < -1 {
		return -1
	}
	if v > 1 {
		return 1
	}
	return v
}
