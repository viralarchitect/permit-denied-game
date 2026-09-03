package contract

import (
	"encoding/json"
	"fmt"
)

// Map is the visual map contract.
// Layers are row-major visual-only grids. All collision lives on objects.
type Map struct {
	SchemaVersion string      `json:"schema_version"`
	ID            ID          `json:"id"`
	TileSize      int         `json:"tile_size"` // world px per tile
	Width         int         `json:"width"`     // tiles
	Height        int         `json:"height"`    // tiles
	Tileset       string      `json:"tileset"`
	Layers        MapLayers   `json:"layers"`
	Objects       []MapObject `json:"objects"`
}

type MapLayers struct {
	Ground []int `json:"ground"` // row-major tile ids, width*height entries
	Decal  []int `json:"decal"`  // row-major tile ids, width*height entries
}

type MapObjectType string

const (
	MapObjectSpawn    MapObjectType = "spawn"
	MapObjectBuilding MapObjectType = "building"
	MapObjectBlocker  MapObjectType = "blocker"
	MapObjectMarker   MapObjectType = "marker"
	MapObjectTrigger  MapObjectType = "trigger"
)

func (t *MapObjectType) UnmarshalJSON(data []byte) error {
	value, err := decodeEnum(data, "map object type",
		string(MapObjectSpawn),
		string(MapObjectBuilding),
		string(MapObjectBlocker),
		string(MapObjectMarker),
		string(MapObjectTrigger),
	)
	if err != nil {
		return err
	}
	*t = MapObjectType(value)
	return nil
}

type MapObject struct {
	ID         ID                         `json:"id"`
	Type       MapObjectType              `json:"type"`
	X          float64                    `json:"x"`       // world px
	Y          float64                    `json:"y"`       // world px
	W          float64                    `json:"w"`       // world px
	H          float64                    `json:"h"`       // world px
	Heading    float64                    `json:"heading"` // radians, 0 = north
	Properties map[string]json.RawMessage `json:"properties"`
}

func ParseMapJSON(b []byte) (Map, error) {
	var doc Map
	if err := decodeStrict(b, &doc); err != nil {
		return Map{}, err
	}
	if err := doc.Validate(); err != nil {
		return Map{}, err
	}
	return doc, nil
}

func (m Map) Validate() error {
	if err := validateSchemaVersion("schema_version", m.SchemaVersion, MapSchemaVersion); err != nil {
		return err
	}
	if err := m.ID.Validate("id"); err != nil {
		return err
	}
	if err := validatePositiveInt("tile_size", m.TileSize); err != nil {
		return err
	}
	if err := validatePositiveInt("width", m.Width); err != nil {
		return err
	}
	if err := validatePositiveInt("height", m.Height); err != nil {
		return err
	}
	if m.Tileset == "" {
		return fmt.Errorf("tileset must not be empty")
	}

	wantTiles := m.Width * m.Height
	if len(m.Layers.Ground) != wantTiles {
		return fmt.Errorf("layers.ground must contain %d row-major entries, got %d", wantTiles, len(m.Layers.Ground))
	}
	if len(m.Layers.Decal) != wantTiles {
		return fmt.Errorf("layers.decal must contain %d row-major entries, got %d", wantTiles, len(m.Layers.Decal))
	}

	ids := make([]ID, 0, len(m.Objects))
	for i, obj := range m.Objects {
		field := fmt.Sprintf("objects[%d]", i)
		if err := obj.ID.Validate(field + ".id"); err != nil {
			return err
		}
		ids = append(ids, obj.ID)
		if err := validateNonNegativeFloat(field+".w", obj.W); err != nil {
			return err
		}
		if err := validateNonNegativeFloat(field+".h", obj.H); err != nil {
			return err
		}
		if obj.Properties == nil {
			obj.Properties = map[string]json.RawMessage{}
		}
		switch obj.Type {
		case MapObjectBuilding:
			if _, ok := obj.Properties["destructible_id"]; !ok {
				return fmt.Errorf("%s.properties.destructible_id is required for building objects", field)
			}
			if _, _, err := obj.PropertyID("destructible_id"); err != nil {
				return err
			}
		case MapObjectSpawn, MapObjectBlocker, MapObjectMarker, MapObjectTrigger:
		default:
			return fmt.Errorf("%s.type is invalid", field)
		}
	}
	return validateUniqueIDs("objects", ids)
}

func (m Map) ObjectByID(id ID) (MapObject, bool) {
	for _, obj := range m.Objects {
		if obj.ID == id {
			return obj, true
		}
	}
	return MapObject{}, false
}

func (o MapObject) PropertyID(name string) (ID, bool, error) {
	raw, ok := o.Properties[name]
	if !ok {
		return "", false, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", true, fmt.Errorf("object %q property %q must be a string id: %w", o.ID, name, err)
	}
	id := ID(value)
	if err := id.Validate(fmt.Sprintf("object %q property %q", o.ID, name)); err != nil {
		return "", true, err
	}
	return id, true, nil
}
