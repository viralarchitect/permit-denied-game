package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

const (
	MapSchemaVersion          = "permitdenied.map.v1"
	VehicleSchemaVersion      = "permitdenied.vehicle.v1"
	DestructibleSchemaVersion = "permitdenied.destructible.v1"
	MissionSchemaVersion      = "permitdenied.mission.v1"
)

var namespacedIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*:[a-z0-9]+(?:[._/-][a-z0-9]+)*$`)

// ID is a runtime-authored namespaced identifier like permitdenied:dozer.
type ID string

func (id ID) String() string {
	return string(id)
}

func (id ID) Validate(field string) error {
	if !namespacedIDPattern.MatchString(string(id)) {
		return fmt.Errorf("%s must be a lowercase namespaced id like permitdenied:example", field)
	}
	return nil
}

// Tag is a namespaced match label used by crush rules and destructible lookup.
type Tag string

func (tag Tag) String() string {
	return string(tag)
}

func (tag Tag) Validate(field string) error {
	if !namespacedIDPattern.MatchString(string(tag)) {
		return fmt.Errorf("%s must be a lowercase namespaced tag like permitdenied:example", field)
	}
	return nil
}

// ScalarValue is the supported mission variable payload: string, number, or bool.
type ScalarValue struct {
	Value any
}

func (v *ScalarValue) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch raw.(type) {
	case bool, string, float64:
		v.Value = raw
		return nil
	default:
		return fmt.Errorf("scalar values must be string, number, or bool")
	}
}

func decodeStrict[T any](b []byte, dst *T) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("expected a single JSON document")
	}
	return nil
}

func decodeEnum(data []byte, field string, allowed ...string) (string, error) {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return "", fmt.Errorf("%s must be a JSON string: %w", field, err)
	}
	for _, candidate := range allowed {
		if value == candidate {
			return value, nil
		}
	}
	sorted := append([]string(nil), allowed...)
	sort.Strings(sorted)
	return "", fmt.Errorf("%s %q is not one of [%s]", field, value, strings.Join(sorted, ", "))
}

func validateSchemaVersion(field, got, want string) error {
	if got != want {
		return fmt.Errorf("%s must be %q, got %q", field, want, got)
	}
	return nil
}

func validatePositiveInt(field string, value int) error {
	if value <= 0 {
		return fmt.Errorf("%s must be > 0", field)
	}
	return nil
}

func validateNonNegativeInt(field string, value int) error {
	if value < 0 {
		return fmt.Errorf("%s must be >= 0", field)
	}
	return nil
}

func validatePositiveFloat(field string, value float64) error {
	if value <= 0 {
		return fmt.Errorf("%s must be > 0", field)
	}
	return nil
}

func validateNonNegativeFloat(field string, value float64) error {
	if value < 0 {
		return fmt.Errorf("%s must be >= 0", field)
	}
	return nil
}

func validateFraction(field string, value float64) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("%s must be in [0, 1]", field)
	}
	return nil
}

func validateUniqueIDs(field string, ids []ID) error {
	seen := make(map[ID]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			return fmt.Errorf("%s contains duplicate id %q", field, id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateUniqueTags(field string, tags []Tag) error {
	seen := make(map[Tag]struct{}, len(tags))
	for _, tag := range tags {
		if _, ok := seen[tag]; ok {
			return fmt.Errorf("%s contains duplicate tag %q", field, tag)
		}
		seen[tag] = struct{}{}
	}
	return nil
}
