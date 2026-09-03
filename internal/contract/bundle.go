package contract

import "fmt"

// Bundle validates the four runtime contracts together without adding pack loading.
type Bundle struct {
	Map           Map
	Vehicles      []Vehicle
	Destructibles []Destructible
	Mission       Mission
}

func (b Bundle) Validate() error {
	if err := b.Map.Validate(); err != nil {
		return fmt.Errorf("map: %w", err)
	}
	if len(b.Vehicles) == 0 {
		return fmt.Errorf("vehicles must not be empty")
	}
	vehicleSet := make(map[ID]Vehicle, len(b.Vehicles))
	for i, vehicle := range b.Vehicles {
		if err := vehicle.Validate(); err != nil {
			return fmt.Errorf("vehicles[%d]: %w", i, err)
		}
		if _, ok := vehicleSet[vehicle.ID]; ok {
			return fmt.Errorf("vehicles contains duplicate id %q", vehicle.ID)
		}
		vehicleSet[vehicle.ID] = vehicle
	}
	if len(b.Destructibles) == 0 {
		return fmt.Errorf("destructibles must not be empty")
	}
	destructibleSet := make(map[ID]Destructible, len(b.Destructibles))
	for i, destructible := range b.Destructibles {
		if err := destructible.Validate(); err != nil {
			return fmt.Errorf("destructibles[%d]: %w", i, err)
		}
		if _, ok := destructibleSet[destructible.ID]; ok {
			return fmt.Errorf("destructibles contains duplicate id %q", destructible.ID)
		}
		destructibleSet[destructible.ID] = destructible
	}
	if err := b.Mission.Validate(); err != nil {
		return fmt.Errorf("mission: %w", err)
	}
	return b.validateReferences(vehicleSet, destructibleSet)
}

func (b Bundle) validateReferences(vehicleSet map[ID]Vehicle, destructibleSet map[ID]Destructible) error {
	if b.Mission.MapID != b.Map.ID {
		return fmt.Errorf("mission.map_id %q does not match map.id %q", b.Mission.MapID, b.Map.ID)
	}
	spawnObj, ok := b.Map.ObjectByID(b.Mission.PlayerSpawnID)
	if !ok {
		return fmt.Errorf("mission.player_spawn_id %q does not match any map object", b.Mission.PlayerSpawnID)
	}
	if spawnObj.Type != MapObjectSpawn {
		return fmt.Errorf("mission.player_spawn_id %q must point to a spawn object", b.Mission.PlayerSpawnID)
	}

	for _, object := range b.Map.Objects {
		if object.Type != MapObjectBuilding {
			continue
		}
		destructibleID, _, err := object.PropertyID("destructible_id")
		if err != nil {
			return err
		}
		if _, ok := destructibleSet[destructibleID]; !ok {
			return fmt.Errorf("map building %q references unknown destructible %q", object.ID, destructibleID)
		}
	}

	for i, objective := range b.Mission.Objectives {
		field := fmt.Sprintf("mission.objectives[%d]", i)
		if err := validateObjectiveRefs(field, objective, destructibleSet, b.Map); err != nil {
			return err
		}
	}
	for i, trigger := range b.Mission.Triggers {
		field := fmt.Sprintf("mission.triggers[%d]", i)
		if err := validateTriggerRefs(field, trigger, destructibleSet, b.Map, vehicleSet, b.Mission); err != nil {
			return err
		}
	}
	for tier, entries := range b.Mission.SpawnTables {
		for i, entry := range entries {
			field := fmt.Sprintf("mission.spawn_tables[%d][%d]", tier, i)
			if _, ok := vehicleSet[entry.UnitID]; !ok {
				return fmt.Errorf("%s.unit_id %q does not match any vehicle id", field, entry.UnitID)
			}
			obj, ok := b.Map.ObjectByID(entry.SpawnMarkerID)
			if !ok {
				return fmt.Errorf("%s.spawn_marker_id %q does not match any map object", field, entry.SpawnMarkerID)
			}
			if obj.Type != MapObjectMarker {
				return fmt.Errorf("%s.spawn_marker_id %q must point to a marker object", field, entry.SpawnMarkerID)
			}
		}
	}
	return nil
}

func validateObjectiveRefs(field string, objective Objective, destructibleSet map[ID]Destructible, gameMap Map) error {
	for _, id := range objective.DestructibleIDs {
		if _, ok := destructibleSet[id]; !ok {
			return fmt.Errorf("%s references unknown destructible id %q", field, id)
		}
	}
	if objective.Type == ObjectiveAreaEntered {
		obj, ok := gameMap.ObjectByID(objective.AreaObjectID)
		if !ok {
			return fmt.Errorf("%s.area_object_id %q does not match any map object", field, objective.AreaObjectID)
		}
		if obj.Type != MapObjectTrigger {
			return fmt.Errorf("%s.area_object_id %q must point to a trigger object", field, objective.AreaObjectID)
		}
	}
	return nil
}

func validateTriggerRefs(field string, trigger Trigger, destructibleSet map[ID]Destructible, gameMap Map, vehicleSet map[ID]Vehicle, mission Mission) error {
	for _, id := range trigger.Filter.DestructibleIDs {
		if _, ok := destructibleSet[id]; !ok {
			return fmt.Errorf("%s.filter references unknown destructible id %q", field, id)
		}
	}
	if trigger.Filter.Event == TriggerOnAreaEnter {
		obj, ok := gameMap.ObjectByID(trigger.Filter.AreaObjectID)
		if !ok {
			return fmt.Errorf("%s.filter.area_object_id %q does not match any map object", field, trigger.Filter.AreaObjectID)
		}
		if obj.Type != MapObjectTrigger {
			return fmt.Errorf("%s.filter.area_object_id %q must point to a trigger object", field, trigger.Filter.AreaObjectID)
		}
	}
	for i, action := range trigger.Actions {
		switch action.Type {
		case ActionSetVariable:
			if _, ok := mission.Variables[action.VariableID.String()]; !ok {
				return fmt.Errorf("%s.actions[%d].variable_id %q is not declared in mission.variables", field, i, action.VariableID)
			}
		case ActionSpawnFromTable:
			for _, entry := range mission.SpawnTables[action.SpawnWantedTier] {
				if _, ok := vehicleSet[entry.UnitID]; !ok {
					return fmt.Errorf("%s.actions[%d] spawn table references unknown vehicle %q", field, i, entry.UnitID)
				}
			}
		}
	}
	return nil
}
