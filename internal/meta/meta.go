package meta

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Eight unlocks. Each changes collapse or stuck.
const (
	Engine1   = "engine_1"
	Engine2   = "engine_2"
	Armor1    = "armor_1"
	Armor2    = "armor_2"
	Ripper    = "ripper"
	Wide      = "wide_blade"
	Ball      = "wrecking_ball"
	MapPool   = "map_pool"
	UnlockCap = 8
)

var Order = []string{Engine1, Armor1, Ripper, Wide, Engine2, Ball, Armor2, MapPool}

type Save struct {
	EngineTier  int      `json:"engine_tier"`
	ArmorTier   int      `json:"armor_tier"`
	Unlocks     []string `json:"unlocks"`
	BestCash    int      `json:"best_cash"`
	HighestTier int      `json:"highest_tier"` // 0 none, 1 county, 2 town, 3 city, 4 capitol
}

func (s Save) Has(id string) bool {
	for _, u := range s.Unlocks {
		if u == id {
			return true
		}
	}
	return false
}

func (s *Save) Grant(id string) bool {
	if s.Has(id) {
		return false
	}
	if len(s.Unlocks) >= UnlockCap {
		return false
	}
	s.Unlocks = append(s.Unlocks, id)
	switch id {
	case Engine1:
		if s.EngineTier < 1 {
			s.EngineTier = 1
		}
	case Engine2:
		if s.EngineTier < 2 {
			s.EngineTier = 2
		}
	case Armor1:
		if s.ArmorTier < 1 {
			s.ArmorTier = 1
		}
	case Armor2:
		if s.ArmorTier < 2 {
			s.ArmorTier = 2
		}
	}
	return true
}

func (s Save) Label(id string) string {
	switch id {
	case Engine1:
		return "ENGINE I — STALL EXTRACT"
	case Engine2:
		return "ENGINE II — BLADE PUSH"
	case Armor1:
		return "ARMOR I — EXTRA PLATE"
	case Armor2:
		return "ARMOR II — SLOWER COOK"
	case Ripper:
		return "START RIPPER"
	case Wide:
		return "START WIDE BLADE"
	case Ball:
		return "START WRECKING BALL"
	case MapPool:
		return "ALT CIVIC LAYOUT"
	default:
		return id
	}
}

func DefaultPath() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "permitdenied", "meta.json")
	}
	return "permitdenied-meta.json"
}

func Load(path string) (Save, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Save{}, nil
		}
		return Save{}, err
	}
	var s Save
	if err := json.Unmarshal(b, &s); err != nil {
		return Save{}, err
	}
	if s.Unlocks == nil {
		s.Unlocks = []string{}
	}
	return s, nil
}

func SaveFile(path string, s Save) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !os.IsExist(err) {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func (s *Save) NoteCash(total int) {
	if total > s.BestCash {
		s.BestCash = total
	}
}

func (s *Save) NoteTier(tier int) {
	if tier > s.HighestTier {
		s.HighestTier = tier
	}
}

func (s *Save) GrantNext() string {
	for _, id := range Order {
		if s.Grant(id) {
			return id
		}
	}
	return ""
}
