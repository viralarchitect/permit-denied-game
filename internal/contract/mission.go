package contract

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
)

// Mission freezes the JSON-authored slice rules: objectives, failure, edge-triggered events,
// wanted tuning, and wanted-tier spawn tables.
type Mission struct {
	SchemaVersion  string                 `json:"schema_version"`
	ID             ID                     `json:"id"`
	MapID          ID                     `json:"map_id"`
	PlayerSpawnID  ID                     `json:"player_spawn_id"`
	InitialState   ID                     `json:"initial_state"`
	Variables      map[string]ScalarValue `json:"variables"`
	Objectives     []Objective            `json:"objectives"`
	FailConditions []FailCondition        `json:"fail_conditions"`
	Triggers       []Trigger              `json:"triggers"`
	Wanted         WantedConfig           `json:"wanted"`
	SpawnTables    SpawnTables            `json:"spawn_tables"`
}

type ObjectiveType string

const (
	ObjectiveDollarsAtLeast ObjectiveType = "dollars_at_least"
	ObjectiveDestroyAll     ObjectiveType = "destroy_all"
	ObjectiveDestroyCount   ObjectiveType = "destroy_count"
	ObjectiveSurviveSeconds ObjectiveType = "survive_seconds"
	ObjectiveAreaEntered    ObjectiveType = "area_entered"
)

func (t *ObjectiveType) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "objective type",
		string(ObjectiveDollarsAtLeast),
		string(ObjectiveDestroyAll),
		string(ObjectiveDestroyCount),
		string(ObjectiveSurviveSeconds),
		string(ObjectiveAreaEntered),
	)
	if err != nil {
		return err
	}
	*t = ObjectiveType(value)
	return nil
}

type Objective struct {
	Type            ObjectiveType `json:"type"`
	DollarsAtLeast  int           `json:"dollars_at_least,omitempty"`
	DestructibleIDs []ID          `json:"destructible_ids,omitempty"`
	TargetTags      []Tag         `json:"target_tags,omitempty"`
	Count           int           `json:"count,omitempty"`
	Seconds         float64       `json:"seconds,omitempty"` // seconds
	AreaObjectID    ID            `json:"area_object_id,omitempty"`
}

type FailConditionType string

const (
	FailEngineHeatAtLeast FailConditionType = "engine_heat_at_least"
	FailArmorAtMost       FailConditionType = "armor_at_most"
	FailPinnedSeconds     FailConditionType = "pinned_seconds"
	FailBuriedSeconds     FailConditionType = "buried_seconds"
	FailTimerExpired      FailConditionType = "timer_expired"
)

func (t *FailConditionType) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "fail condition type",
		string(FailEngineHeatAtLeast),
		string(FailArmorAtMost),
		string(FailPinnedSeconds),
		string(FailBuriedSeconds),
		string(FailTimerExpired),
	)
	if err != nil {
		return err
	}
	*t = FailConditionType(value)
	return nil
}

type FailCondition struct {
	Type  FailConditionType `json:"type"`
	Value float64           `json:"value"` // heat units, armor units, or seconds by condition type
}

type TriggerEvent string

const (
	TriggerOnDestroy   TriggerEvent = "on_destroy"
	TriggerOnAreaEnter TriggerEvent = "on_area_enter"
	TriggerTimerTick   TriggerEvent = "timer_tick"
	TriggerOnPin       TriggerEvent = "on_pin"
)

func (e *TriggerEvent) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "trigger event",
		string(TriggerOnDestroy),
		string(TriggerOnAreaEnter),
		string(TriggerTimerTick),
		string(TriggerOnPin),
	)
	if err != nil {
		return err
	}
	*e = TriggerEvent(value)
	return nil
}

type Trigger struct {
	ID              ID              `json:"id"`
	Enabled         bool            `json:"enabled"`
	Once            bool            `json:"once"`
	Priority        int             `json:"priority"`
	Filter          TriggerFilter   `json:"filter"`
	Actions         []TriggerAction `json:"actions"`
	CooldownSeconds float64         `json:"cooldown_seconds"` // seconds
}

// TriggerFilter describes edge-triggered events only: rubble-entry, outside-to-inside area entry,
// scheduled timer ticks, and the first pin threshold crossing.
type TriggerFilter struct {
	Event                TriggerEvent `json:"event"`
	DestructibleIDs      []ID         `json:"destructible_ids,omitempty"`
	TargetTags           []Tag        `json:"target_tags,omitempty"`
	AreaObjectID         ID           `json:"area_object_id,omitempty"`
	TimerSeconds         float64      `json:"timer_seconds,omitempty"`           // seconds
	RepeatSeconds        float64      `json:"repeat_seconds,omitempty"`          // 0 = one-shot
	PinnedSecondsAtLeast float64      `json:"pinned_seconds_at_least,omitempty"` // seconds
}

type TriggerActionType string

const (
	ActionSetState           TriggerActionType = "set_state"
	ActionSetTriggerEnabled  TriggerActionType = "set_trigger_enabled"
	ActionSetVariable        TriggerActionType = "set_variable"
	ActionSetWantedLevel     TriggerActionType = "set_wanted_level"
	ActionSetWantedAwareness TriggerActionType = "set_wanted_awareness"
	ActionSpawnFromTable     TriggerActionType = "spawn_from_table"
	ActionEndMission         TriggerActionType = "end_mission"
)

func (t *TriggerActionType) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "trigger action type",
		string(ActionSetState),
		string(ActionSetTriggerEnabled),
		string(ActionSetVariable),
		string(ActionSetWantedLevel),
		string(ActionSetWantedAwareness),
		string(ActionSpawnFromTable),
		string(ActionEndMission),
	)
	if err != nil {
		return err
	}
	*t = TriggerActionType(value)
	return nil
}

type MissionOutcome string

const (
	OutcomeSuccess MissionOutcome = "success"
	OutcomeFailure MissionOutcome = "failure"
)

func (o *MissionOutcome) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "mission outcome", string(OutcomeSuccess), string(OutcomeFailure))
	if err != nil {
		return err
	}
	*o = MissionOutcome(value)
	return nil
}

type TriggerAction struct {
	Type            TriggerActionType `json:"type"`
	StateID         ID                `json:"state_id,omitempty"`
	TriggerID       ID                `json:"trigger_id,omitempty"`
	Enabled         *bool             `json:"enabled,omitempty"`
	VariableID      ID                `json:"variable_id,omitempty"`
	Value           *ScalarValue      `json:"value,omitempty"`
	WantedLevel     int               `json:"wanted_level,omitempty"`
	WantedAwareness WantedAwareness   `json:"wanted_awareness,omitempty"`
	SpawnWantedTier int               `json:"spawn_wanted_tier,omitempty"`
	Outcome         MissionOutcome    `json:"outcome,omitempty"`
}

type WantedAwareness string

const (
	AwarenessUnaware   WantedAwareness = "unaware"
	AwarenessSearching WantedAwareness = "searching"
	AwarenessLocated   WantedAwareness = "located"
)

func (a *WantedAwareness) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "wanted awareness",
		string(AwarenessUnaware),
		string(AwarenessSearching),
		string(AwarenessLocated),
	)
	if err != nil {
		return err
	}
	*a = WantedAwareness(value)
	return nil
}

type PoliceState string

const (
	PolicePatrol   PoliceState = "patrol"
	PolicePursuit  PoliceState = "pursuit"
	PoliceRam      PoliceState = "ram"
	PoliceBlockade PoliceState = "blockade"
	PoliceBust     PoliceState = "bust"
)

func (s *PoliceState) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "police state",
		string(PolicePatrol),
		string(PolicePursuit),
		string(PoliceRam),
		string(PoliceBlockade),
		string(PoliceBust),
	)
	if err != nil {
		return err
	}
	*s = PoliceState(value)
	return nil
}

type WantedCrime string

const (
	CrimePropertyDamage WantedCrime = "property_damage"
	CrimeHitPolice      WantedCrime = "hit_police"
	CrimeDestroyPolice  WantedCrime = "destroy_police"
)

func (c *WantedCrime) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "wanted crime",
		string(CrimePropertyDamage),
		string(CrimeHitPolice),
		string(CrimeDestroyPolice),
	)
	if err != nil {
		return err
	}
	*c = WantedCrime(value)
	return nil
}

type WantedConfig struct {
	Min            int                   `json:"min"`
	Max            int                   `json:"max"`
	Initial        int                   `json:"initial"`
	CrimeAwards    []WantedCrimeAward    `json:"crime_awards"`
	TierThresholds []WantedTierThreshold `json:"tier_thresholds"`
	Decay          WantedDecay           `json:"decay"`
	Awareness      WantedAwareness       `json:"awareness"`
}

type WantedCrimeAward struct {
	Crime       WantedCrime `json:"crime"`
	WantedDelta int         `json:"wanted_delta"`
}

type WantedTierThreshold struct {
	Tier          int `json:"tier"`
	WantedAtLeast int `json:"wanted_at_least"`
}

type WantedDecay struct {
	DelaySeconds    float64 `json:"delay_seconds"`     // seconds
	WantedPerSecond float64 `json:"wanted_per_second"` // wanted units per second
}

// SpawnTables are keyed to the exact wanted tier.
// They never silently inherit lower tiers. GTA2-style tier-5/tier-6 replacement must be authored explicitly.
type SpawnTables map[int][]SpawnTableEntry

func (s *SpawnTables) UnmarshalJSON(data []byte) error {
	var raw map[string][]SpawnTableEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	out := make(SpawnTables, len(raw))
	for key, entries := range raw {
		tier, err := strconv.Atoi(key)
		if err != nil {
			return fmt.Errorf("spawn_tables keys must be integer wanted tiers, got %q", key)
		}
		out[tier] = entries
	}
	*s = out
	return nil
}

func (s SpawnTables) EntriesForWantedTier(tier int) []SpawnTableEntry {
	return slices.Clone(s[tier])
}

type SpawnTableEntry struct {
	UnitID          ID            `json:"unit_id"`
	Weight          int           `json:"weight"`
	Count           int           `json:"count"`
	Cap             int           `json:"cap"`
	SpawnMarkerID   ID            `json:"spawn_marker_id"`
	CooldownSeconds float64       `json:"cooldown_seconds"` // seconds
	AllowedStates   []PoliceState `json:"allowed_states"`
}

func ParseMissionJSON(b []byte) (Mission, error) {
	var doc Mission
	if err := decodeStrict(b, &doc); err != nil {
		return Mission{}, err
	}
	if err := doc.Validate(); err != nil {
		return Mission{}, err
	}
	return doc, nil
}

func (m Mission) Validate() error {
	if err := validateSchemaVersion("schema_version", m.SchemaVersion, MissionSchemaVersion); err != nil {
		return err
	}
	if err := m.ID.Validate("id"); err != nil {
		return err
	}
	if err := m.MapID.Validate("map_id"); err != nil {
		return err
	}
	if err := m.PlayerSpawnID.Validate("player_spawn_id"); err != nil {
		return err
	}
	if err := m.InitialState.Validate("initial_state"); err != nil {
		return err
	}
	for key := range m.Variables {
		if err := ID(key).Validate("variables key"); err != nil {
			return err
		}
	}
	switch m.Wanted.Awareness {
	case AwarenessUnaware, AwarenessSearching, AwarenessLocated:
	default:
		return fmt.Errorf("wanted.awareness is required")
	}
	if len(m.Objectives) == 0 {
		return fmt.Errorf("objectives must not be empty")
	}
	for i, objective := range m.Objectives {
		if err := objective.Validate(fmt.Sprintf("objectives[%d]", i)); err != nil {
			return err
		}
	}
	if len(m.FailConditions) == 0 {
		return fmt.Errorf("fail_conditions must not be empty")
	}
	for i, condition := range m.FailConditions {
		if err := condition.Validate(fmt.Sprintf("fail_conditions[%d]", i)); err != nil {
			return err
		}
	}

	triggerIDs := make([]ID, 0, len(m.Triggers))
	for i, trigger := range m.Triggers {
		field := fmt.Sprintf("triggers[%d]", i)
		if err := trigger.ID.Validate(field + ".id"); err != nil {
			return err
		}
		triggerIDs = append(triggerIDs, trigger.ID)
		if err := validateNonNegativeFloat(field+".cooldown_seconds", trigger.CooldownSeconds); err != nil {
			return err
		}
		if err := trigger.Filter.Validate(field + ".filter"); err != nil {
			return err
		}
		if len(trigger.Actions) == 0 {
			return fmt.Errorf("%s.actions must not be empty", field)
		}
	}
	if err := validateUniqueIDs("triggers", triggerIDs); err != nil {
		return err
	}
	triggerSet := make(map[ID]struct{}, len(triggerIDs))
	for _, id := range triggerIDs {
		triggerSet[id] = struct{}{}
	}
	for i, trigger := range m.Triggers {
		for j, action := range trigger.Actions {
			if err := action.Validate(fmt.Sprintf("triggers[%d].actions[%d]", i, j), triggerSet, m); err != nil {
				return err
			}
		}
	}
	if err := m.Wanted.Validate(); err != nil {
		return err
	}
	return m.SpawnTables.Validate(m.Wanted)
}

func (o Objective) Validate(field string) error {
	switch o.Type {
	case ObjectiveDollarsAtLeast:
		return validatePositiveInt(field+".dollars_at_least", o.DollarsAtLeast)
	case ObjectiveDestroyAll:
		if len(o.DestructibleIDs) == 0 && len(o.TargetTags) == 0 {
			return fmt.Errorf("%s requires destructible_ids or target_tags", field)
		}
	case ObjectiveDestroyCount:
		if len(o.DestructibleIDs) == 0 && len(o.TargetTags) == 0 {
			return fmt.Errorf("%s requires destructible_ids or target_tags", field)
		}
		if err := validatePositiveInt(field+".count", o.Count); err != nil {
			return err
		}
	case ObjectiveSurviveSeconds:
		return validatePositiveFloat(field+".seconds", o.Seconds)
	case ObjectiveAreaEntered:
		return o.AreaObjectID.Validate(field + ".area_object_id")
	default:
		return fmt.Errorf("%s.type is invalid", field)
	}
	for i, id := range o.DestructibleIDs {
		if err := id.Validate(fmt.Sprintf("%s.destructible_ids[%d]", field, i)); err != nil {
			return err
		}
	}
	if err := validateUniqueIDs(field+".destructible_ids", o.DestructibleIDs); err != nil {
		return err
	}
	for i, tag := range o.TargetTags {
		if err := tag.Validate(fmt.Sprintf("%s.target_tags[%d]", field, i)); err != nil {
			return err
		}
	}
	return validateUniqueTags(field+".target_tags", o.TargetTags)
}

func (f FailCondition) Validate(field string) error {
	switch f.Type {
	case FailEngineHeatAtLeast, FailArmorAtMost, FailPinnedSeconds, FailBuriedSeconds, FailTimerExpired:
	default:
		return fmt.Errorf("%s.type is required", field)
	}
	return validatePositiveFloat(field+".value", f.Value)
}

func (f TriggerFilter) Validate(field string) error {
	switch f.Event {
	case TriggerOnDestroy:
		if len(f.DestructibleIDs) == 0 && len(f.TargetTags) == 0 {
			return fmt.Errorf("%s requires destructible_ids or target_tags for on_destroy", field)
		}
	case TriggerOnAreaEnter:
		if err := f.AreaObjectID.Validate(field + ".area_object_id"); err != nil {
			return err
		}
	case TriggerTimerTick:
		if err := validateNonNegativeFloat(field+".timer_seconds", f.TimerSeconds); err != nil {
			return err
		}
		if err := validateNonNegativeFloat(field+".repeat_seconds", f.RepeatSeconds); err != nil {
			return err
		}
	case TriggerOnPin:
		if err := validatePositiveFloat(field+".pinned_seconds_at_least", f.PinnedSecondsAtLeast); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s.event is invalid", field)
	}
	for i, id := range f.DestructibleIDs {
		if err := id.Validate(fmt.Sprintf("%s.destructible_ids[%d]", field, i)); err != nil {
			return err
		}
	}
	if err := validateUniqueIDs(field+".destructible_ids", f.DestructibleIDs); err != nil {
		return err
	}
	for i, tag := range f.TargetTags {
		if err := tag.Validate(fmt.Sprintf("%s.target_tags[%d]", field, i)); err != nil {
			return err
		}
	}
	return validateUniqueTags(field+".target_tags", f.TargetTags)
}

func (a TriggerAction) Validate(field string, triggerSet map[ID]struct{}, mission Mission) error {
	switch a.Type {
	case ActionSetState:
		return a.StateID.Validate(field + ".state_id")
	case ActionSetTriggerEnabled:
		if err := a.TriggerID.Validate(field + ".trigger_id"); err != nil {
			return err
		}
		if _, ok := triggerSet[a.TriggerID]; !ok {
			return fmt.Errorf("%s.trigger_id %q does not match any trigger id", field, a.TriggerID)
		}
		if a.Enabled == nil {
			return fmt.Errorf("%s.enabled is required", field)
		}
	case ActionSetVariable:
		if err := a.VariableID.Validate(field + ".variable_id"); err != nil {
			return err
		}
		if a.Value == nil {
			return fmt.Errorf("%s.value is required", field)
		}
	case ActionSetWantedLevel:
		if a.WantedLevel < mission.Wanted.Min || a.WantedLevel > mission.Wanted.Max {
			return fmt.Errorf("%s.wanted_level must be within wanted min/max", field)
		}
	case ActionSetWantedAwareness:
		switch a.WantedAwareness {
		case AwarenessUnaware, AwarenessSearching, AwarenessLocated:
		default:
			return fmt.Errorf("%s.wanted_awareness is required", field)
		}
	case ActionSpawnFromTable:
		if _, ok := mission.SpawnTables[a.SpawnWantedTier]; !ok {
			return fmt.Errorf("%s.spawn_wanted_tier %d has no spawn table", field, a.SpawnWantedTier)
		}
	case ActionEndMission:
		switch a.Outcome {
		case OutcomeSuccess, OutcomeFailure:
		default:
			return fmt.Errorf("%s.outcome is required", field)
		}
	default:
		return fmt.Errorf("%s.type is invalid", field)
	}
	return nil
}

func (w WantedConfig) Validate() error {
	if err := validateNonNegativeInt("wanted.min", w.Min); err != nil {
		return err
	}
	if w.Max < w.Min {
		return fmt.Errorf("wanted.max must be >= wanted.min")
	}
	if w.Initial < w.Min || w.Initial > w.Max {
		return fmt.Errorf("wanted.initial must be within wanted min/max")
	}
	crimes := make(map[WantedCrime]struct{}, len(w.CrimeAwards))
	for i, award := range w.CrimeAwards {
		field := fmt.Sprintf("wanted.crime_awards[%d]", i)
		if _, ok := crimes[award.Crime]; ok {
			return fmt.Errorf("%s duplicate crime %q", field, award.Crime)
		}
		crimes[award.Crime] = struct{}{}
		if err := validateNonNegativeInt(field+".wanted_delta", award.WantedDelta); err != nil {
			return err
		}
	}
	if len(w.TierThresholds) == 0 {
		return fmt.Errorf("wanted.tier_thresholds must not be empty")
	}
	lastTier := -1
	lastWanted := -1
	for i, threshold := range w.TierThresholds {
		field := fmt.Sprintf("wanted.tier_thresholds[%d]", i)
		if err := validateNonNegativeInt(field+".tier", threshold.Tier); err != nil {
			return err
		}
		if err := validateNonNegativeInt(field+".wanted_at_least", threshold.WantedAtLeast); err != nil {
			return err
		}
		if threshold.WantedAtLeast < w.Min || threshold.WantedAtLeast > w.Max {
			return fmt.Errorf("%s.wanted_at_least must be within wanted min/max", field)
		}
		if threshold.Tier <= lastTier {
			return fmt.Errorf("%s.tier must be strictly increasing", field)
		}
		if threshold.WantedAtLeast <= lastWanted {
			return fmt.Errorf("%s.wanted_at_least must be strictly increasing", field)
		}
		lastTier = threshold.Tier
		lastWanted = threshold.WantedAtLeast
	}
	if err := validateNonNegativeFloat("wanted.decay.delay_seconds", w.Decay.DelaySeconds); err != nil {
		return err
	}
	return validateNonNegativeFloat("wanted.decay.wanted_per_second", w.Decay.WantedPerSecond)
}

func (s SpawnTables) Validate(w WantedConfig) error {
	if len(s) == 0 {
		return fmt.Errorf("spawn_tables must not be empty")
	}
	knownTiers := make(map[int]struct{}, len(w.TierThresholds))
	for _, threshold := range w.TierThresholds {
		knownTiers[threshold.Tier] = struct{}{}
	}
	for tier, entries := range s {
		if _, ok := knownTiers[tier]; !ok {
			return fmt.Errorf("spawn_tables[%d] does not match any wanted tier threshold", tier)
		}
		if len(entries) == 0 {
			return fmt.Errorf("spawn_tables[%d] must not be empty", tier)
		}
		for i, entry := range entries {
			field := fmt.Sprintf("spawn_tables[%d][%d]", tier, i)
			if err := entry.UnitID.Validate(field + ".unit_id"); err != nil {
				return err
			}
			if err := entry.SpawnMarkerID.Validate(field + ".spawn_marker_id"); err != nil {
				return err
			}
			if err := validatePositiveInt(field+".weight", entry.Weight); err != nil {
				return err
			}
			if err := validatePositiveInt(field+".count", entry.Count); err != nil {
				return err
			}
			if err := validatePositiveInt(field+".cap", entry.Cap); err != nil {
				return err
			}
			if err := validateNonNegativeFloat(field+".cooldown_seconds", entry.CooldownSeconds); err != nil {
				return err
			}
			if len(entry.AllowedStates) == 0 {
				return fmt.Errorf("%s.allowed_states must not be empty", field)
			}
		}
	}
	return nil
}
