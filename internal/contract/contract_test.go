package contract

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidRuntimeContracts(t *testing.T) {
	gameMap, err := ParseMapJSON([]byte(validMapJSON()))
	if err != nil {
		t.Fatalf("parse map: %v", err)
	}
	dozer, err := ParseVehicleJSON([]byte(validDozerJSON()))
	if err != nil {
		t.Fatalf("parse dozer: %v", err)
	}
	cruiser, err := ParseVehicleJSON([]byte(validCruiserJSON()))
	if err != nil {
		t.Fatalf("parse cruiser: %v", err)
	}
	destructible, err := ParseDestructibleJSON([]byte(validDestructibleJSON()))
	if err != nil {
		t.Fatalf("parse destructible: %v", err)
	}
	mission, err := ParseMissionJSON([]byte(validMissionJSON()))
	if err != nil {
		t.Fatalf("parse mission: %v", err)
	}

	if dozer.Crush.Mode != CrushOverrun {
		t.Fatalf("dozer crush mode = %q, want %q", dozer.Crush.Mode, CrushOverrun)
	}
	if len(dozer.Parts) != 1 || dozer.Parts[0].Kind != VehiclePartBlade {
		t.Fatalf("dozer parts = %+v, want one blade part", dozer.Parts)
	}
	if cruiser.Crush.Mode != CrushRam {
		t.Fatalf("cruiser crush mode = %q, want %q", cruiser.Crush.Mode, CrushRam)
	}

	bundle := Bundle{
		Map:           gameMap,
		Vehicles:      []Vehicle{dozer, cruiser},
		Destructibles: []Destructible{destructible},
		Mission:       mission,
	}
	if err := bundle.Validate(); err != nil {
		t.Fatalf("validate bundle: %v", err)
	}
}

func TestParseVehicleJSONRejectsUnknownEnum(t *testing.T) {
	_, err := ParseVehicleJSON([]byte(stringsReplace(validDozerJSON(), `"mode": "overrun"`, `"mode": "flatten"`)))
	if err == nil {
		t.Fatal("expected unknown crush mode error")
	}
}

func TestBundleRejectsMissingAndUnknownReferences(t *testing.T) {
	t.Run("missing building destructible id", func(t *testing.T) {
		_, err := ParseMapJSON([]byte(stringsReplace(validMapJSON(), `"destructible_id": "permitdenied:destructible.sheriff",`, ``)))
		if err == nil {
			t.Fatal("expected missing destructible id error")
		}
	})

	t.Run("unknown player spawn id", func(t *testing.T) {
		gameMap, err := ParseMapJSON([]byte(validMapJSON()))
		if err != nil {
			t.Fatalf("parse map: %v", err)
		}
		dozer, _ := ParseVehicleJSON([]byte(validDozerJSON()))
		cruiser, _ := ParseVehicleJSON([]byte(validCruiserJSON()))
		destructible, _ := ParseDestructibleJSON([]byte(validDestructibleJSON()))
		mission, err := ParseMissionJSON([]byte(stringsReplace(validMissionJSON(), `"player_spawn_id": "permitdenied:spawn.player",`, `"player_spawn_id": "permitdenied:spawn.ghost",`)))
		if err != nil {
			t.Fatalf("parse mission: %v", err)
		}
		err = Bundle{
			Map:           gameMap,
			Vehicles:      []Vehicle{dozer, cruiser},
			Destructibles: []Destructible{destructible},
			Mission:       mission,
		}.Validate()
		if err == nil {
			t.Fatal("expected missing player spawn reference error")
		}
	})
}

func TestDestructibleTransitionRubbleEdgeIsOneShot(t *testing.T) {
	doc, err := ParseDestructibleJSON([]byte(validDestructibleJSON()))
	if err != nil {
		t.Fatalf("parse destructible: %v", err)
	}

	next, entered, err := doc.Transition(StateCracked, 0)
	if err != nil {
		t.Fatalf("first transition: %v", err)
	}
	if next != StateRubble || !entered {
		t.Fatalf("first transition = (%q, %v), want (%q, true)", next, entered, StateRubble)
	}

	next, entered, err = doc.Transition(StateRubble, 0)
	if err != nil {
		t.Fatalf("second transition: %v", err)
	}
	if next != StateRubble || entered {
		t.Fatalf("second transition = (%q, %v), want (%q, false)", next, entered, StateRubble)
	}
}

func TestSpawnTablesUseExactTierPolicy(t *testing.T) {
	mission, err := ParseMissionJSON([]byte(validMissionJSON()))
	if err != nil {
		t.Fatalf("parse mission: %v", err)
	}

	tierOne := mission.SpawnTables.EntriesForWantedTier(1)
	tierFive := mission.SpawnTables.EntriesForWantedTier(5)
	if len(tierOne) != 1 {
		t.Fatalf("tier 1 entries = %d, want 1", len(tierOne))
	}
	if len(tierFive) != 1 {
		t.Fatalf("tier 5 entries = %d, want 1", len(tierFive))
	}
	if tierFive[0].Count == tierOne[0].Count && tierFive[0].Cap == tierOne[0].Cap {
		t.Fatalf("tier 5 should be authored explicitly, got tier 1 values %+v", tierFive[0])
	}
	if got := mission.SpawnTables.EntriesForWantedTier(6); len(got) != 0 {
		t.Fatalf("tier 6 entries = %d, want 0 for exact-tier lookup", len(got))
	}
}

func validMapJSON() string {
	return fmt.Sprintf(`{
  "schema_version": %q,
  "id": "permitdenied:map.county_slice",
  "tile_size": 16,
  "width": 2,
  "height": 2,
  "tileset": "usable/tileset.png",
  "layers": {
    "ground": [1, 1, 1, 1],
    "decal": [0, 0, 0, 0]
  },
  "objects": [
    {
      "id": "permitdenied:spawn.player",
      "type": "spawn",
      "x": 16,
      "y": 24,
      "w": 0,
      "h": 0,
      "heading": 0,
      "properties": {}
    },
    {
      "id": "permitdenied:building.sheriff",
      "type": "building",
      "x": 16,
      "y": 16,
      "w": 32,
      "h": 32,
      "heading": 0,
      "properties": {
        "destructible_id": "permitdenied:destructible.sheriff",
        "role": "target"
      }
    },
    {
      "id": "permitdenied:marker.north_road",
      "type": "marker",
      "x": 32,
      "y": 0,
      "w": 16,
      "h": 16,
      "heading": 0,
      "properties": {}
    },
    {
      "id": "permitdenied:trigger.yard_gate",
      "type": "trigger",
      "x": 0,
      "y": 16,
      "w": 16,
      "h": 16,
      "heading": 0,
      "properties": {}
    }
  ]
}`, MapSchemaVersion)
}

func validDozerJSON() string {
	return fmt.Sprintf(`{
  "schema_version": %q,
  "id": "permitdenied:vehicle.dozer",
  "role": "dozer",
  "tags": ["permitdenied:vehicle.dozer"],
  "sprite_set": "permitdenied:sprites.dozer",
  "collider": {
    "shape": "circle",
    "radius": 14
  },
  "mass": 52,
  "top_speed": 110,
  "reverse_speed": 55,
  "acceleration": 180,
  "braking": 220,
  "turn_rate": 2.4,
  "traction": 1.1,
  "armor": {
    "max": 4,
    "collision_damage_scale": 0.75
  },
  "engine_heat_rate": {
    "load": 28,
    "stalled": 45,
    "cool_moving": 18,
    "cool_idle": 8,
    "max": 100
  },
  "crush": {
    "mode": "overrun",
    "power": 1.4,
    "minimum_speed": 20,
    "target_tags": ["permitdenied:target.overrun"]
  },
  "parts": [
    {
      "id": "permitdenied:part.dozer_blade",
      "kind": "blade",
      "width": 32,
      "reach": 18,
      "raised_speed_scale": 1,
      "lowered_speed_scale": 0.44,
      "raised_turn_scale": 1,
      "lowered_turn_scale": 0.46,
      "damage_rate": 26
    }
  ]
}`, VehicleSchemaVersion)
}

func validCruiserJSON() string {
	return fmt.Sprintf(`{
  "schema_version": %q,
  "id": "permitdenied:vehicle.cruiser",
  "role": "cruiser",
  "tags": ["permitdenied:vehicle.cruiser_heavy"],
  "sprite_set": "permitdenied:sprites.cruiser",
  "collider": {
    "shape": "circle",
    "radius": 8
  },
  "mass": 16,
  "top_speed": 95,
  "reverse_speed": 24,
  "acceleration": 160,
  "braking": 180,
  "turn_rate": 3.1,
  "traction": 0.9,
  "armor": {
    "max": 1,
    "collision_damage_scale": 1
  },
  "engine_heat_rate": {
    "load": 0,
    "stalled": 0,
    "cool_moving": 0,
    "cool_idle": 0,
    "max": 1
  },
  "crush": {
    "mode": "ram",
    "power": 0.6,
    "minimum_speed": 18,
    "target_tags": []
  },
  "parts": []
}`, VehicleSchemaVersion)
}

func validDestructibleJSON() string {
	return fmt.Sprintf(`{
  "schema_version": %q,
  "id": "permitdenied:destructible.sheriff",
  "tags": ["permitdenied:target.overrun", "permitdenied:objective.county"],
  "material": "brick",
  "health": 60,
  "yield_dollars": 400,
  "states": [
    {
      "id": "intact",
      "enter_at_health_fraction": 1,
      "sprite": "permitdenied:sheriff.intact",
      "collision": "solid"
    },
    {
      "id": "cracked",
      "enter_at_health_fraction": 0.5,
      "sprite": "permitdenied:sheriff.cracked",
      "collision": "solid"
    },
    {
      "id": "rubble",
      "enter_at_health_fraction": 0,
      "sprite": "permitdenied:sheriff.rubble",
      "collision": "ramp"
    }
  ],
  "rubble_rule": {
    "spawn_collision": true,
    "inset": 2,
    "mass": 1024,
    "ramp": true,
    "counts_toward_bury": true,
    "persist": true
  }
}`, DestructibleSchemaVersion)
}

func validMissionJSON() string {
	return fmt.Sprintf(`{
  "schema_version": %q,
  "id": "permitdenied:mission.county_slice",
  "map_id": "permitdenied:map.county_slice",
  "player_spawn_id": "permitdenied:spawn.player",
  "initial_state": "permitdenied:state.live",
  "variables": {
    "permitdenied:var.alerted": false
  },
  "objectives": [
    {
      "type": "dollars_at_least",
      "dollars_at_least": 400
    },
    {
      "type": "area_entered",
      "area_object_id": "permitdenied:trigger.yard_gate"
    }
  ],
  "fail_conditions": [
    {
      "type": "engine_heat_at_least",
      "value": 100
    },
    {
      "type": "timer_expired",
      "value": 210
    }
  ],
  "triggers": [
    {
      "id": "permitdenied:trigger.first_rubble",
      "enabled": true,
      "once": true,
      "priority": 100,
      "filter": {
        "event": "on_destroy",
        "destructible_ids": ["permitdenied:destructible.sheriff"]
      },
      "actions": [
        {
          "type": "set_variable",
          "variable_id": "permitdenied:var.alerted",
          "value": true
        }
      ],
      "cooldown_seconds": 0
    },
    {
      "id": "permitdenied:trigger.timer_reinforce",
      "enabled": true,
      "once": false,
      "priority": 50,
      "filter": {
        "event": "timer_tick",
        "timer_seconds": 60,
        "repeat_seconds": 30
      },
      "actions": [
        {
          "type": "spawn_from_table",
          "spawn_wanted_tier": 5
        }
      ],
      "cooldown_seconds": 0
    }
  ],
  "wanted": {
    "min": 0,
    "max": 100,
    "initial": 0,
    "crime_awards": [
      {"crime": "property_damage", "wanted_delta": 5},
      {"crime": "hit_police", "wanted_delta": 20},
      {"crime": "destroy_police", "wanted_delta": 35}
    ],
    "tier_thresholds": [
      {"tier": 1, "wanted_at_least": 10},
      {"tier": 5, "wanted_at_least": 60}
    ],
    "decay": {
      "delay_seconds": 8,
      "wanted_per_second": 1
    },
    "awareness": "searching"
  },
  "spawn_tables": {
    "1": [
      {
        "unit_id": "permitdenied:vehicle.cruiser",
        "weight": 1,
        "count": 1,
        "cap": 2,
        "spawn_marker_id": "permitdenied:marker.north_road",
        "cooldown_seconds": 6,
        "allowed_states": ["patrol", "pursuit"]
      }
    ],
    "5": [
      {
        "unit_id": "permitdenied:vehicle.cruiser",
        "weight": 1,
        "count": 2,
        "cap": 4,
        "spawn_marker_id": "permitdenied:marker.north_road",
        "cooldown_seconds": 12,
        "allowed_states": ["ram", "blockade"]
      }
    ]
  }
}`, MissionSchemaVersion)
}

func stringsReplace(s, old, new string) string {
	return strings.Replace(s, old, new, 1)
}
