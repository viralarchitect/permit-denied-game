package game

import (
	"math"

	"permitdenied/internal/lot"
	"permitdenied/internal/threats"
)

type aabb struct {
	x, y, w, h float64
}

func (g *Game) stepPlay(in Input) {
	// 1. Input already read.
	// 2. Clock.
	g.run.Tick++
	if g.run.Tick >= g.buzzerTick() {
		g.endRun("buzzer")
		return
	}

	if g.dozer.IFrames > 0 {
		g.dozer.IFrames--
	}

	// 3. Beat chart (once flags). Time() includes PressureBonus (§11).
	if g.scene == ScenePlay {
		g.fireBeats()
	} else {
		g.fireCampaignBeats()
	}

	// 4. Chopper linger → PressureBonus += 0.35 * dt while inside spot.
	if g.chopper.Active {
		dx := g.dozer.X - g.chopper.X
		dy := g.dozer.Y - g.chopper.Y
		if dx*dx+dy*dy <= g.chopper.SpotR*g.chopper.SpotR {
			g.run.PressureBonus += PressureRate * Dt
		}
	}

	// 5. Steer / throttle / integrate candidate.
	g.integrateDozer(in)

	// 6. Blade toggle.
	if in.BladeToggle {
		g.dozer.BladeDown = !g.dozer.BladeDown
	}

	solids := g.collectSolids()

	// 7. Resolve dozer vs solids. Slide along; do not tunnel.
	stalled, overlapping := g.resolveDozer(solids)

	bx, by, bw, bh := bladeAABBW(g.dozer.X, g.dozer.Y, g.dozer.Heading, g.dozer.BladeDown, g.bladeWidth())
	bladeHitSolid := false
	deepRubble := false
	for _, s := range solids {
		if aabbOverlap(bx, by, bw, bh, s.x, s.y, s.w, s.h) {
			bladeHitSolid = true
		}
	}
	for i := range g.lot.Rubble {
		r := g.lot.Rubble[i]
		if r.W*r.H >= DeepRubbleA {
			if _, _, _, hit := CircleAABB(g.dozer.X, g.dozer.Y, DozerBodyR, r.X, r.Y, r.W, r.H); hit {
				deepRubble = true
			}
		}
	}

	// 8. Blade-down wreck.
	if g.dozer.BladeDown {
		g.glanceLatch = false
		eating := g.wreckWithBlade(bx, by, bw, bh)
		g.audio.DuckWreck(eating)
	} else {
		g.jerseyLatch = map[int]struct{}{}
		g.glanceBuildings()
		g.audio.DuckWreck(false)
	}

	// 10. Heat.
	g.stepHeat(stalled, overlapping, bladeHitSolid, deepRubble)

	// 11. Cruisers seek, collide.
	g.stepCruisers()

	// 12. Excavator boom if arrived.
	g.stepExcavator(bx, by, bw, bh)

	// 13. Chopper follows lazily (max 70 px/s).
	g.stepChopper()

	// 14. Peds wander; flatten garnish.
	g.stepPeds()

	g.stepHeavies(bx, by, bw, bh)
	g.collectPickups()
	if g.scene == SceneCapitol {
		g.tower.MarkReached(g.dozer.X, g.dozer.Y, g.lot)
		if g.tower.CoreDown(g.lot) && !g.tower.CoreCollapsed {
			g.collapseTower()
		}
	}

	// 15. Death.
	if g.checkDeath() {
		return
	}
	g.stepImmobilize(stalled, overlapping)
	if g.run.Over {
		return
	}
	if g.mapCleared() {
		g.advanceMap()
		return
	}

	// 16. Decay shake, dollar lives.
	g.fx.Step(Dt, ShakeDecay)
}

func (g *Game) integrateDozer(in Input) {
	turn := TurnRateBladeUp
	if g.dozer.BladeDown {
		turn = TurnRateBladeDown
	}
	g.dozer.Heading = wrapHeading(g.dozer.Heading + in.Steer*turn*Dt)

	max := SpeedFwdUp
	if in.Throttle < 0 {
		max = SpeedRevUp
		if g.dozer.BladeDown {
			max = SpeedRevDown
		}
	} else if g.dozer.BladeDown {
		max = SpeedFwdDown
		if g.progress.EngineTier >= 2 {
			max *= MetaEngineDown
		}
	}
	target := max * in.Throttle
	g.dozer.Speed = g.approachSpeed(g.dozer.Speed, target)

	fx, fy := Forward(g.dozer.Heading)
	g.dozer.X += fx * g.dozer.Speed * Dt
	g.dozer.Y += fy * g.dozer.Speed * Dt
}

func approachSpeed(speed, target float64) float64 {
	return approachSpeedAccel(speed, target, AccelFwd)
}

func (g *Game) approachSpeed(speed, target float64) float64 {
	a := AccelFwd
	if g.progress.EngineTier >= 1 {
		a *= MetaEngineAccel
	}
	return approachSpeedAccel(speed, target, a)
}

func approachSpeedAccel(speed, target, accelFwd float64) float64 {
	if speed == target {
		return speed
	}
	var a float64
	braking := (speed > 0 && target < speed) || (speed < 0 && target > speed)
	if braking {
		a = AccelBrake
	} else if target < 0 || speed < 0 {
		a = AccelRev
	} else {
		a = accelFwd
	}
	if speed < target {
		speed += a * Dt
		if speed > target {
			speed = target
		}
	} else {
		speed -= a * Dt
		if speed < target {
			speed = target
		}
	}
	return speed
}

func (g *Game) collectSolids() []aabb {
	out := make([]aabb, 0, 32)
	for i := range g.lot.Buildings {
		b := g.lot.Buildings[i]
		if b.State == lot.InRubble {
			continue
		}
		out = append(out, aabb{b.X, b.Y, b.W, b.H})
	}
	for i := range g.lot.Rubble {
		r := g.lot.Rubble[i]
		if r.Ramp {
			continue
		}
		out = append(out, aabb{r.X, r.Y, r.W, r.H})
	}
	for i := range g.heavies {
		h := g.heavies[i]
		if !h.Alive || !h.Solid {
			continue
		}
		x, y, w, hh := h.Body()
		out = append(out, aabb{x, y, w, hh})
	}
	for i := range g.blockers {
		b := g.blockers[i]
		if !b.Alive || !b.Solid {
			continue
		}
		out = append(out, aabb{b.X, b.Y, b.W, b.H})
	}
	if g.excavator.Arrived && g.excavator.Alive {
		x, y, w, h := g.excavator.Body()
		out = append(out, aabb{x, y, w, h})
	}
	return out
}

func (g *Game) resolveDozer(solids []aabb) (stalled, overlapping bool) {
	for iter := 0; iter < 3; iter++ {
		for _, s := range solids {
			nx, ny, pen, hit := CircleAABB(g.dozer.X, g.dozer.Y, DozerBodyR, s.x, s.y, s.w, s.h)
			if !hit {
				continue
			}
			overlapping = true
			g.dozer.X += nx * pen
			g.dozer.Y += ny * pen
			fx, fy := Forward(g.dozer.Heading)
			vx, vy := fx*g.dozer.Speed, fy*g.dozer.Speed
			if vx*nx+vy*ny < 0 {
				if g.dozer.Speed > 0 {
					g.dozer.Speed = math.Min(g.dozer.Speed, 0)
				} else {
					g.dozer.Speed = math.Max(g.dozer.Speed, 0)
				}
			}
		}
	}
	g.dozer.X = clamp(g.dozer.X, DozerBodyR, g.worldW()-DozerBodyR)
	g.dozer.Y = clamp(g.dozer.Y, DozerBodyR, g.worldH()-DozerBodyR)
	stalled = overlapping && math.Abs(g.dozer.Speed) < StallSpeed
	return stalled, overlapping
}

func (g *Game) stepCruisers() {
	fx, fy := Forward(g.dozer.Heading)
	rx, ry := Right(g.dozer.Heading)
	for i := range g.cruisers {
		c := &g.cruisers[i]
		if !c.Alive {
			continue
		}
		side := 1.0
		if i%2 == 0 {
			side = -1
		}
		tx := g.dozer.X - fx*CruiserOffsetF + rx*side*CruiserOffsetS
		ty := g.dozer.Y - fy*CruiserOffsetF + ry*side*CruiserOffsetS
		dx, dy := tx-c.X, ty-c.Y
		dist := hypot(dx, dy)
		if dist > 1 {
			c.X += dx / dist * CruiserSpeed * Dt
			c.Y += dy / dist * CruiserSpeed * Dt
			c.Heading = math.Atan2(dx, -dy)
		}
		c.X = clamp(c.X, CruiserRadius, g.worldW()-CruiserRadius)
		c.Y = clamp(c.Y, CruiserRadius, g.worldH()-CruiserRadius)

		ddx, ddy := c.X-g.dozer.X, c.Y-g.dozer.Y
		if ddx*ddx+ddy*ddy >= (DozerBodyR+CruiserRadius)*(DozerBodyR+CruiserRadius) {
			continue
		}
		along := ddx*fx + ddy*fy
		front := along > FrontAlong
		if front && g.dozer.BladeDown {
			c.Alive = false
			g.run.VehicleCash += CruiserKillCash
			g.fx.SpawnDollar(c.X, c.Y, CruiserKillCash, DollarLife)
			g.fx.SpawnBoom(c.X, c.Y)
			g.fx.Shake += ShakeWreck * 0.4
			g.audio.Burst()
			continue
		}
		if !front && g.dozer.BladeDown == false && g.dozer.IFrames == 0 && g.run.CruiserPIT {
			g.peel("cruiser")
		}
		// bounce
		n := hypot(ddx, ddy)
		if n < 1e-6 {
			n = 1
			ddx, ddy = 0, -1
		}
		c.X += ddx / n * 18
		c.Y += ddy / n * 18
	}
}

func (g *Game) stepExcavator(bx, by, bw, bh float64) {
	ex := &g.excavator
	if !ex.Arrived || !ex.Alive {
		return
	}
	ex.Swinging = true
	// buried: rubble overlapping boom origin + right*20
	rrx, rry := Right(ex.Heading)
	originX := ex.X + rrx*20
	originY := ex.Y + rry*20
	buried := g.lot.RubbleBlocks(originX-8, originY-8, 16, 16)

	if !buried {
		ex.BoomPhase += Dt / BoomPeriod
		if ex.BoomPhase >= 1 {
			ex.BoomPhase -= 1
		}
		ang := (ex.BoomPhase - 0.5) * BoomSweep
		hdg := wrapHeading(ex.Heading + ang)
		fx, fy := Forward(hdg)
		x0, y0 := ex.X, ex.Y
		x1, y1 := ex.X+fx*BoomLen, ex.Y+fy*BoomLen
		if distPointSeg(g.dozer.X, g.dozer.Y, x0, y0, x1, y1) < DozerBodyR+BoomHalfW {
			if g.dozer.IFrames == 0 {
				g.dozer.Plates--
				g.dozer.IFrames = IFramesBoom
				g.fx.Shake += ShakePeel
				g.audio.Peel()
				if g.dozer.Plates <= 0 {
					g.fx.SpawnBoom(g.dozer.X, g.dozer.Y)
				}
			}
		}
	}

	// Blade-down body ram: excavator HP 40, you cook (heat handled via solid overlap).
	if g.dozer.BladeDown {
		exx, exy, exw, exh := ex.Body()
		if aabbOverlap(bx, by, bw, bh, exx, exy, exw, exh) {
			ex.HP -= WreckRateDown * Dt
			if ex.HP <= 0 {
				ex.Alive = false
				g.run.VehicleCash += ExcavatorKillCash
				g.fx.SpawnDollar(ex.X, ex.Y, ExcavatorKillCash, DollarLife)
				g.fx.SpawnBoom(ex.X, ex.Y)
				g.fx.HitStop = HitStopTicks
				g.fx.Shake += ShakePeel
				g.audio.Burst()
			}
		}
	}
}

func (g *Game) stepChopper() {
	if !g.chopper.Active {
		return
	}
	dx := g.dozer.X - g.chopper.X
	dy := g.dozer.Y - g.chopper.Y
	dist := hypot(dx, dy)
	if dist > 1 {
		step := ChopperSpeed * Dt
		if step > dist {
			step = dist
		}
		g.chopper.X += dx / dist * step
		g.chopper.Y += dy / dist * step
	}
}

func (g *Game) stepPeds() {
	for i := range g.peds {
		p := &g.peds[i]
		if !p.Alive {
			continue
		}
		p.Wait -= Dt
		if p.Wait <= 0 {
			p.Heading = wrapHeading(p.Heading + 0.9)
			p.Wait = 1.4 + float64(i)*0.17
		}
		fx, fy := Forward(p.Heading)
		p.X = clamp(p.X+fx*PedWanderSpeed*Dt, PedRadius, g.worldW()-PedRadius)
		p.Y = clamp(p.Y+fy*PedWanderSpeed*Dt, PedRadius, g.worldH()-PedRadius)
		dx, dy := p.X-g.dozer.X, p.Y-g.dozer.Y
		if dx*dx+dy*dy < (DozerBodyR+PedRadius)*(DozerBodyR+PedRadius) && math.Abs(g.dozer.Speed) > FlattenSpeed {
			p.Alive = false
			g.run.VehicleCash += PedCash
			g.fx.SpawnDollar(p.X, p.Y, PedCash, DollarLife)
		}
	}
}

func (g *Game) endRun(death string) {
	if g.run.Over {
		return
	}
	g.run.Over = true
	g.run.Death = death
	g.run.TimeAlive = math.Min(g.run.Time(), g.clockLimit())
	if g.scene == ScenePlay || g.run.Tier == 0 {
		g.run.RecountTargets()
	} else {
		down, _ := g.lot.NamedDown()
		g.run.Targets = down
	}
	g.scene = SceneTally
	g.fx.TallyT = 0
}

func (g *Game) fireCampaignBeats() {
	t := g.run.Time()
	if !g.wave.t30 && t >= 30 {
		g.wave.t30 = true
		g.spawnCruisers(2)
	}
	if !g.wave.t80 && t >= 80 {
		g.wave.t80 = true
		g.spawnCruisers(1)
		if g.scene == SceneCapitol && !g.excavator.Arrived {
			g.placeExcavatorAt(g.mapW/2, g.mapH*0.62)
		}
	}
	if !g.wave.t140 && t >= 140 {
		g.wave.t140 = true
		g.spawnCruisers(2)
	}
}

func (g *Game) stepHeavies(bx, by, bw, bh float64) {
	for i := range g.heavies {
		h := &g.heavies[i]
		if !h.Alive {
			continue
		}
		if !h.Parked {
			dx, dy := g.dozer.X-h.X, g.dozer.Y-h.Y
			dist := hypot(dx, dy)
			spd := HeavyFireSpeed
			if h.Kind == threats.HeavyWagon {
				spd = HeavyWagonSpeed
			}
			if dist > 1 {
				h.X += dx / dist * spd * Dt
				h.Y += dy / dist * spd * Dt
				h.Heading = math.Atan2(dx, -dy)
			}
			h.X = clamp(h.X, h.W/2, g.worldW()-h.W/2)
			h.Y = clamp(h.Y, h.H/2, g.worldH()-h.H/2)
		}
		if !g.dozer.BladeDown {
			continue
		}
		hx, hy, hw, hh := h.Body()
		if !aabbOverlap(bx, by, bw, bh, hx, hy, hw, hh) {
			continue
		}
		rate := WreckRateDown
		if g.kit.Driver {
			rate *= DriverMul
		}
		h.HP -= rate * Dt
		if h.HP > 0 {
			continue
		}
		h.Alive = false
		h.Solid = false
		cash := HeavyFireCash
		if h.Kind == threats.HeavyWagon {
			cash = HeavyWagonCash
		}
		g.run.VehicleCash += cash
		g.fx.SpawnDollar(h.X, h.Y, cash, DollarLife)
		g.fx.SpawnBoom(h.X, h.Y)
		g.fx.Shake += ShakeWreck * 0.5
		g.lot.AddRubble(hx, hy, hw, hh, RubbleInset)
	}
}

func (g *Game) collapseTower() {
	hit := g.tower.CollapseFromCore(&g.lot)
	for _, i := range hit {
		if i < 0 || i >= len(g.lot.Buildings) {
			continue
		}
		b := &g.lot.Buildings[i]
		g.spawnBuildingRubble(*b)
		g.run.StructCash += b.Value
		cx, cy := b.Center()
		g.fx.SpawnDollar(cx, cy, b.Value, DollarLife)
		g.fx.SpawnBoom(cx, cy)
	}
	if len(hit) > 0 {
		g.fx.HitStop = HitStopTicks
		g.fx.Shake += ShakeWreck * 2
		g.fx.SetBanner("PERMIT: TOWER DROPPED", BannerLife)
		g.audio.Wreck()
	}
}
