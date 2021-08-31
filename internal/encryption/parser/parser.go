package parser

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"
)

type PasswordEntry struct {
	Label   string
	Secret  string
	Primary bool
}

func Parse(input string) ([]PasswordEntry, error) {
	if len(input) == 0 {
		return nil, nil
	}

	var r receiver
	if err := json.Unmarshal([]byte(input), &r); err != nil {
		return nil, fmt.Errorf("password configuration JSON error: %w", err)
	}

	if len(r) == 0 {
		return nil, nil
	}

	result := convert(r)

	if err := validate(result); err != nil {
		return nil, fmt.Errorf("password configuration error: %w", err)
	}

	return result, nil
}

type receiver []struct {
	Label    string `json:"label"`
	Primary  bool   `json:"primary"`
	Password struct {
		Secret string `json:"secret"`
	} `json:"password"`
}

func convert(r receiver) []PasswordEntry {
	var result []PasswordEntry
	for _, p := range r {
		result = append(result, PasswordEntry{
			Label:   p.Label,
			Secret:  p.Password.Secret,
			Primary: p.Primary,
		})
	}

	return result
}

func validate(passwordEntries []PasswordEntry) (errs *validation.FieldError) {
	labels := make(map[string]struct{})
	primaries := 0
	for i, p := range passwordEntries {
		if p.Primary {
			primaries++
		}
		errs = errs.Also(
			validation.ErrIfOutsideLength(p.Secret, "secret.password", 20, 1024).ViaIndex(i),
			validation.ErrIfOutsideLength(p.Label, "label", 5, 20).ViaIndex(i),
			validation.ErrIfDuplicate(p.Label, "label", labels).ViaIndex(i),
		)
	}

	switch primaries {
	case 0:
		return errs.Also(&validation.FieldError{
			Message: "expected exactly one primary, got none",
			Paths:   []string{"[].primary"},
		})
	case 1:
		return errs
	default:
		return errs.Also(&validation.FieldError{
			Message: "expected exactly one primary, got multiple",
			Paths:   []string{"[].primary"},
		})
	}
}
