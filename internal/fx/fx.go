package fx

import "math"

type Dollar struct {
	X, Y float64
	Amt  int
	Life float64
}

type FX struct {
	HitStop int
	Shake   float64
	Dollars []Dollar
	Bursts  []Burst
	Banner  string
	BannerT float64
	TallyT  float64
}

func (f *FX) SpawnDollar(x, y float64, amt int, life float64) {
	f.Dollars = append(f.Dollars, Dollar{X: x, Y: y, Amt: amt, Life: life})
}

func (f *FX) SetBanner(s string, life float64) {
	f.Banner = s
	f.BannerT = life
}

func (f *FX) Step(dt, shakeDecay float64) {
	if f.Shake > 0 {
		f.Shake -= shakeDecay * dt
		if f.Shake < 0 {
			f.Shake = 0
		}
	}
	if f.BannerT > 0 {
		f.BannerT -= dt
		if f.BannerT <= 0 {
			f.BannerT = 0
			f.Banner = ""
		}
	}
	n := 0
	for i := range f.Dollars {
		d := f.Dollars[i]
		d.Life -= dt
		if d.Life <= 0 {
			continue
		}
		f.Dollars[n] = d
		n++
	}
	f.Dollars = f.Dollars[:n]
	bn := 0
	for i := range f.Bursts {
		b := f.Bursts[i]
		b.Age++
		if b.Dead() {
			continue
		}
		f.Bursts[bn] = b
		bn++
	}
	f.Bursts = f.Bursts[:bn]
}

func (f *FX) Offsets(tick int) (sx, sy float64) {
	if f.Shake <= 0 {
		return 0, 0
	}
	t := float64(tick)
	return f.Shake * math.Sin(t*1.7), f.Shake * math.Cos(t*1.9)
}
