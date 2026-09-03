package contract

import "fmt"

// Vehicle freezes the JSON-authored player and police chassis contract.
// Speed values are in px/s, acceleration and braking are in px/s^2, and turn rate is in rad/s.
type Vehicle struct {
	SchemaVersion  string          `json:"schema_version"`
	ID             ID              `json:"id"`
	Role           VehicleRole     `json:"role"`
	Tags           []Tag           `json:"tags,omitempty"`
	SpriteSet      ID              `json:"sprite_set"`
	Collider       VehicleCollider `json:"collider"`
	Mass           float64         `json:"mass"`
	TopSpeed       float64         `json:"top_speed"`
	ReverseSpeed   float64         `json:"reverse_speed"`
	Acceleration   float64         `json:"acceleration"`
	Braking        float64         `json:"braking"`
	TurnRate       float64         `json:"turn_rate"`
	Traction       float64         `json:"traction"`
	Armor          VehicleArmor    `json:"armor"`
	EngineHeatRate EngineHeatRate  `json:"engine_heat_rate"`
	Crush          CrushRule       `json:"crush"`
	Parts          []VehiclePart   `json:"parts"`
}

type VehicleRole string

const (
	VehicleRoleDozer   VehicleRole = "dozer"
	VehicleRoleCruiser VehicleRole = "cruiser"
)

func (r *VehicleRole) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "vehicle role", string(VehicleRoleDozer), string(VehicleRoleCruiser))
	if err != nil {
		return err
	}
	*r = VehicleRole(value)
	return nil
}

type ColliderShape string

const (
	ColliderCircle ColliderShape = "circle"
	ColliderBox    ColliderShape = "box"
)

func (s *ColliderShape) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "collider shape", string(ColliderCircle), string(ColliderBox))
	if err != nil {
		return err
	}
	*s = ColliderShape(value)
	return nil
}

type VehicleCollider struct {
	Shape  ColliderShape `json:"shape"`
	Radius float64       `json:"radius,omitempty"` // world px, for circle colliders
	Width  float64       `json:"width,omitempty"`  // world px, for box colliders
	Height float64       `json:"height,omitempty"` // world px, for box colliders
}

type VehicleArmor struct {
	Max                  float64 `json:"max"`
	CollisionDamageScale float64 `json:"collision_damage_scale"`
}

type EngineHeatRate struct {
	Load       float64 `json:"load"`        // heat units per second while loaded
	Stalled    float64 `json:"stalled"`     // heat units per second while stalled
	CoolMoving float64 `json:"cool_moving"` // heat units per second while moving
	CoolIdle   float64 `json:"cool_idle"`   // heat units per second while idling
	Max        float64 `json:"max"`
}

type CrushMode string

const (
	CrushNone    CrushMode = "none"
	CrushRam     CrushMode = "ram"
	CrushOverrun CrushMode = "overrun"
)

func (m *CrushMode) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "crush mode", string(CrushNone), string(CrushRam), string(CrushOverrun))
	if err != nil {
		return err
	}
	*m = CrushMode(value)
	return nil
}

type CrushRule struct {
	Mode         CrushMode `json:"mode"`
	Power        float64   `json:"power"`
	MinimumSpeed float64   `json:"minimum_speed"` // px/s
	TargetTags   []Tag     `json:"target_tags"`
}

type VehiclePartKind string

const (
	VehiclePartBlade VehiclePartKind = "blade"
)

func (k *VehiclePartKind) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "vehicle part kind", string(VehiclePartBlade))
	if err != nil {
		return err
	}
	*k = VehiclePartKind(value)
	return nil
}

type VehiclePart struct {
	ID                ID              `json:"id"`
	Kind              VehiclePartKind `json:"kind"`
	Width             float64         `json:"width"`              // world px
	Reach             float64         `json:"reach"`              // world px forward from chassis center
	RaisedSpeedScale  float64         `json:"raised_speed_scale"` // multiplier
	LoweredSpeedScale float64         `json:"lowered_speed_scale"`
	RaisedTurnScale   float64         `json:"raised_turn_scale"` // multiplier
	LoweredTurnScale  float64         `json:"lowered_turn_scale"`
	DamageRate        float64         `json:"damage_rate"` // health units per second
}

func ParseVehicleJSON(b []byte) (Vehicle, error) {
	var doc Vehicle
	if err := decodeStrict(b, &doc); err != nil {
		return Vehicle{}, err
	}
	if err := doc.Validate(); err != nil {
		return Vehicle{}, err
	}
	return doc, nil
}

func (v Vehicle) Validate() error {
	if err := validateSchemaVersion("schema_version", v.SchemaVersion, VehicleSchemaVersion); err != nil {
		return err
	}
	if err := v.ID.Validate("id"); err != nil {
		return err
	}
	if err := v.SpriteSet.Validate("sprite_set"); err != nil {
		return err
	}
	switch v.Role {
	case VehicleRoleDozer, VehicleRoleCruiser:
	default:
		return fmt.Errorf("role is required")
	}
	if err := validateUniqueTags("tags", v.Tags); err != nil {
		return err
	}
	for i, tag := range v.Tags {
		if err := tag.Validate(fmt.Sprintf("tags[%d]", i)); err != nil {
			return err
		}
	}

	switch v.Collider.Shape {
	case ColliderCircle:
		if err := validatePositiveFloat("collider.radius", v.Collider.Radius); err != nil {
			return err
		}
	case ColliderBox:
		if err := validatePositiveFloat("collider.width", v.Collider.Width); err != nil {
			return err
		}
		if err := validatePositiveFloat("collider.height", v.Collider.Height); err != nil {
			return err
		}
	default:
		return fmt.Errorf("collider.shape is required")
	}

	if err := validatePositiveFloat("mass", v.Mass); err != nil {
		return err
	}
	if err := validatePositiveFloat("top_speed", v.TopSpeed); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("reverse_speed", v.ReverseSpeed); err != nil {
		return err
	}
	if err := validatePositiveFloat("acceleration", v.Acceleration); err != nil {
		return err
	}
	if err := validatePositiveFloat("braking", v.Braking); err != nil {
		return err
	}
	if err := validatePositiveFloat("turn_rate", v.TurnRate); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("traction", v.Traction); err != nil {
		return err
	}
	if err := validatePositiveFloat("armor.max", v.Armor.Max); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("armor.collision_damage_scale", v.Armor.CollisionDamageScale); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("engine_heat_rate.load", v.EngineHeatRate.Load); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("engine_heat_rate.stalled", v.EngineHeatRate.Stalled); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("engine_heat_rate.cool_moving", v.EngineHeatRate.CoolMoving); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("engine_heat_rate.cool_idle", v.EngineHeatRate.CoolIdle); err != nil {
		return err
	}
	if err := validatePositiveFloat("engine_heat_rate.max", v.EngineHeatRate.Max); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("crush.power", v.Crush.Power); err != nil {
		return err
	}
	if err := validateNonNegativeFloat("crush.minimum_speed", v.Crush.MinimumSpeed); err != nil {
		return err
	}
	switch v.Crush.Mode {
	case CrushNone, CrushRam, CrushOverrun:
	default:
		return fmt.Errorf("crush.mode is required")
	}
	if err := validateUniqueTags("crush.target_tags", v.Crush.TargetTags); err != nil {
		return err
	}
	for i, tag := range v.Crush.TargetTags {
		if err := tag.Validate(fmt.Sprintf("crush.target_tags[%d]", i)); err != nil {
			return err
		}
	}

	partIDs := make([]ID, 0, len(v.Parts))
	for i, part := range v.Parts {
		field := fmt.Sprintf("parts[%d]", i)
		if err := part.ID.Validate(field + ".id"); err != nil {
			return err
		}
		partIDs = append(partIDs, part.ID)
		switch part.Kind {
		case VehiclePartBlade:
		default:
			return fmt.Errorf("%s.kind is required", field)
		}
		if err := validatePositiveFloat(field+".width", part.Width); err != nil {
			return err
		}
		if err := validateNonNegativeFloat(field+".reach", part.Reach); err != nil {
			return err
		}
		if err := validateNonNegativeFloat(field+".raised_speed_scale", part.RaisedSpeedScale); err != nil {
			return err
		}
		if err := validateNonNegativeFloat(field+".lowered_speed_scale", part.LoweredSpeedScale); err != nil {
			return err
		}
		if err := validateNonNegativeFloat(field+".raised_turn_scale", part.RaisedTurnScale); err != nil {
			return err
		}
		if err := validateNonNegativeFloat(field+".lowered_turn_scale", part.LoweredTurnScale); err != nil {
			return err
		}
		if err := validateNonNegativeFloat(field+".damage_rate", part.DamageRate); err != nil {
			return err
		}
	}
	return validateUniqueIDs("parts", partIDs)
}
