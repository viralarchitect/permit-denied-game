package contract

import "fmt"

// Destructible freezes the JSON-authored building and rubble contract.
// Yield dollars and on_destroy fire on the first transition into the rubble state only.
type Destructible struct {
	SchemaVersion string               `json:"schema_version"`
	ID            ID                   `json:"id"`
	Tags          []Tag                `json:"tags"`
	Material      DestructibleMaterial `json:"material"`
	Health        float64              `json:"health"`
	YieldDollars  int                  `json:"yield_dollars"`
	States        []DestructibleState  `json:"states"`
	RubbleRule    RubbleRule           `json:"rubble_rule"`
}

type DestructibleMaterial string

const (
	MaterialWood     DestructibleMaterial = "wood"
	MaterialBrick    DestructibleMaterial = "brick"
	MaterialConcrete DestructibleMaterial = "concrete"
	MaterialSteel    DestructibleMaterial = "steel"
)

func (m *DestructibleMaterial) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "destructible material",
		string(MaterialWood),
		string(MaterialBrick),
		string(MaterialConcrete),
		string(MaterialSteel),
	)
	if err != nil {
		return err
	}
	*m = DestructibleMaterial(value)
	return nil
}

type DestructibleStateID string

const (
	StateIntact  DestructibleStateID = "intact"
	StateCracked DestructibleStateID = "cracked"
	StateRubble  DestructibleStateID = "rubble"
)

func (id *DestructibleStateID) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "destructible state id",
		string(StateIntact),
		string(StateCracked),
		string(StateRubble),
	)
	if err != nil {
		return err
	}
	*id = DestructibleStateID(value)
	return nil
}

type CollisionMode string

const (
	CollisionSolid    CollisionMode = "solid"
	CollisionPassable CollisionMode = "passable"
	CollisionRamp     CollisionMode = "ramp"
)

func (m *CollisionMode) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "collision mode",
		string(CollisionSolid),
		string(CollisionPassable),
		string(CollisionRamp),
	)
	if err != nil {
		return err
	}
	*m = CollisionMode(value)
	return nil
}

type DestructibleState struct {
	ID                    DestructibleStateID `json:"id"`
	EnterAtHealthFraction float64             `json:"enter_at_health_fraction"`
	Sprite                ID                  `json:"sprite"`
	Collision             CollisionMode       `json:"collision"`
}

type RubbleRule struct {
	SpawnCollision   bool    `json:"spawn_collision"`
	Inset            float64 `json:"inset"` // world px
	Mass             float64 `json:"mass"`  // simulation mass units
	Ramp             bool    `json:"ramp"`
	CountsTowardBury bool    `json:"counts_toward_bury"`
	Persist          bool    `json:"persist"`
}

// Transition reports the current state for a health fraction and whether this step hit the
// rubble-entry edge. Callers should award dollars and fire on_destroy only when enteredRubble is true.
// Once a destructible is already rubble, further damage is a no-op from the contract's perspective.
func (d Destructible) Transition(currentState DestructibleStateID, healthFraction float64) (nextState DestructibleStateID, enteredRubble bool, err error) {
	if err := validateFraction("health fraction", healthFraction); err != nil {
		return "", false, err
	}
	next, err := d.StateForHealthFraction(healthFraction)
	if err != nil {
		return "", false, err
	}
	if currentState == StateRubble {
		return StateRubble, false, nil
	}
	return next.ID, currentState != StateRubble && next.ID == StateRubble, nil
}

func ParseDestructibleJSON(b []byte) (Destructible, error) {
	var doc Destructible
	if err := decodeStrict(b, &doc); err != nil {
		return Destructible{}, err
	}
	if err := doc.Validate(); err != nil {
		return Destructible{}, err
	}
	return doc, nil
}

func (d Destructible) Validate() error {
	if err := validateSchemaVersion("schema_version", d.SchemaVersion, DestructibleSchemaVersion); err != nil {
		return err
	}
	if err := d.ID.Validate("id"); err != nil {
		return err
	}
	if err := validateUniqueTags("tags", d.Tags); err != nil {
		return err
	}
	for i, tag := range d.Tags {
		if err := tag.Validate(fmt.Sprintf("tags[%d]", i)); err != nil {
			return err
		}
	}
	switch d.Material {
	case MaterialWood, MaterialBrick, MaterialConcrete, MaterialSteel:
	default:
		return fmt.Errorf("material is required")
	}
	if err := validatePositiveFloat("health", d.Health); err != nil {
		return err
	}
	if err := validateNonNegativeInt("yield_dollars", d.YieldDollars); err != nil {
		return err
	}
	if len(d.States) != 3 {
		return fmt.Errorf("states must contain exactly the frozen ids intact, cracked, and rubble")
	}

	found := make(map[DestructibleStateID]DestructibleState, len(d.States))
	for i, state := range d.States {
		field := fmt.Sprintf("states[%d]", i)
		if _, ok := found[state.ID]; ok {
			return fmt.Errorf("%s.id duplicates %q", field, state.ID)
		}
		found[state.ID] = state
		if err := validateFraction(field+".enter_at_health_fraction", state.EnterAtHealthFraction); err != nil {
			return err
		}
		if err := state.Sprite.Validate(field + ".sprite"); err != nil {
			return err
		}
		switch state.Collision {
		case CollisionSolid, CollisionPassable, CollisionRamp:
		default:
			return fmt.Errorf("%s.collision is required", field)
		}
	}
	intact, ok := found[StateIntact]
	if !ok {
		return fmt.Errorf("states must include id %q", StateIntact)
	}
	cracked, ok := found[StateCracked]
	if !ok {
		return fmt.Errorf("states must include id %q", StateCracked)
	}
	rubble, ok := found[StateRubble]
	if !ok {
		return fmt.Errorf("states must include id %q", StateRubble)
	}
	if intact.EnterAtHealthFraction != 1 {
		return fmt.Errorf("states.intact.enter_at_health_fraction must be 1")
	}
	if rubble.EnterAtHealthFraction != 0 {
		return fmt.Errorf("states.rubble.enter_at_health_fraction must be 0")
	}
	if cracked.EnterAtHealthFraction <= rubble.EnterAtHealthFraction || cracked.EnterAtHealthFraction >= intact.EnterAtHealthFraction {
		return fmt.Errorf("states.cracked.enter_at_health_fraction must fall strictly between intact and rubble")
	}

	if err := validateNonNegativeFloat("rubble_rule.inset", d.RubbleRule.Inset); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("rubble_rule.mass", d.RubbleRule.Mass); err != nil {
		return err
	}
	return nil
}

func (d Destructible) StateForHealthFraction(healthFraction float64) (DestructibleState, error) {
	if err := validateFraction("health fraction", healthFraction); err != nil {
		return DestructibleState{}, err
	}

	var best DestructibleState
	found := false
	for _, state := range d.States {
		if healthFraction > state.EnterAtHealthFraction {
			continue
		}
		if !found || state.EnterAtHealthFraction < best.EnterAtHealthFraction {
			best = state
			found = true
		}
	}
	if !found {
		return DestructibleState{}, fmt.Errorf("no destructible state matches health fraction %v", healthFraction)
	}
	return best, nil
}
