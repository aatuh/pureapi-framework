package defaults

import (
	"fmt"
)

// InputConversionRules returns the default input conversion rules.
//
// Returns:
//   - map[string]func(any) any: The input conversion rules.
func InputConversionRules() map[string]func(any) any {
	return map[string]func(any) any{
		"uuid": func(a any) any {
			strVal, ok := a.(string)
			if !ok {
				return nil
			}
			uuidVal, err := NewUUIDGen().FromString(strVal)
			if err != nil {
				return nil
			}
			return uuidVal
		},
	}
}

// ValidationRules returns the default validation rules.
//
// Returns:
//   - map[string]func(any) error: The validation rules.
func ValidationRules() map[string]func(any) error {
	return map[string]func(any) error{
		"uuid": func(a any) error {
			_, err := NewUUIDGen().FromString(fmt.Sprintf("%v", a))
			if err != nil {
				return fmt.Errorf("invalid uuid")
			}
			return nil
		},
	}
}
