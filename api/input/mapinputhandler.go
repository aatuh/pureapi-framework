package input

import (
	"fmt"
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/aatuh/pureapi-framework/defaults"
	"github.com/aatuh/pureapi-util/objectpicker"
	"github.com/aatuh/pureapi-util/urlencoder"
	"github.com/aatuh/pureapi-util/validate"
	"github.com/mitchellh/mapstructure"
)

// allowedKeyRegex defines the allowed pattern for query parameter keys.
// It allows only alphanumeric, underscore '_', dot '.', and hyphen '-'
// characters.
var allowedKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// FieldError represents a field-level validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrValidationData contains a list of field-level validation errors
type ErrValidationData struct {
	Errors []FieldError `json:"errors"`
}

// MapInputHandler handles the input of a request as a map.
type MapInputHandler[Input any] struct {
	apiFields      APIFields
	conversionMap  map[string]func(any) any
	customRules    map[string]func(any) error
	inputFactoryFn func() *Input
}

// NewMapInputHandler creates a new MapInputHandler.
//
// Parameters:
//   - apiFields: The APIFields to use for validation.
//   - conversionMap: A map of conversion functions for fields.
//   - customRules: A map of custom validation rules for fields.
//   - inputFactoryFn: A function that returns a new Input.
//
// Returns:
//   - *MapInputHandler: The new MapInputHandler.
func NewMapInputHandler[Input any](
	apiFields APIFields,
	conversionMap map[string]func(any) any,
	customRules map[string]func(any) error,
	inputFactoryFn func() *Input,
) *MapInputHandler[Input] {
	return &MapInputHandler[Input]{
		apiFields:      apiFields,
		conversionMap:  conversionMap,
		customRules:    customRules,
		inputFactoryFn: inputFactoryFn,
	}
}

// ValidateAPIFields validates the provided APIFields. It returns the
// MapInputHandler if they are valid.
//
// Returns:
//   - *MapInputHandler: The MapInputHandler.
//   - error: Any error that occurred during validation.
func (h *MapInputHandler[Input]) ValidateAPIFields() (*MapInputHandler[Input], error) {
	err := h.validateRules(h.apiFields)
	if err != nil {
		return nil, err
	}
	_, err = h.mapFieldConfigFromAPIFields(h.apiFields)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// MustValidateAPIFields validates the provided APIFields. It panics if an error
// occurs.
//
// Returns:
//   - *MapInputHandler: The MapInputHandler.
func (h *MapInputHandler[Input]) MustValidateAPIFields() *MapInputHandler[Input] {
	_, err := h.ValidateAPIFields()
	if err != nil {
		panic(err)
	}
	return h
}

// Handle processes the request input by creating a map presentation from it and
// validating it.
//
// Parameters:
//   - w: The HTTP response writer.
//   - r: The HTTP request.
//
// Returns:
//   - *Input: The map presentation of the input.
//   - error: Any error that occurred during processing.
func (h *MapInputHandler[Input]) Handle(
	w http.ResponseWriter, r *http.Request,
) (*Input, error) {
	// Pick input as map.
	dataMap, err := h.pickMap(r, h.apiFields)
	if err != nil {
		defaults.CtxLogger(r.Context()).Tracef("Error picking input: %v", err)
		return nil, ErrInvalidInput
	}
	// Validate the map.
	if err := h.validate(dataMap, h.apiFields); err != nil {
		return nil, ErrValidation.WithData(ErrValidationData{
			Errors: []FieldError{
				{
					Field:   "input",
					Message: err.Error(),
				},
			},
		})
	}
	// Convert the map to the object.
	input, err := h.mapToObject(dataMap, h.inputFactoryFn())
	if err != nil {
		return nil, err
	}
	return input, nil
}

// mapToObject decodes a map into the provided object.
func (h *MapInputHandler[Input]) mapToObject(
	value map[string]any, obj *Input,
) (*Input, error) {
	cfg := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           obj,
		TagName:          "json",
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return nil, fmt.Errorf("mapToObject: error creating decoder: %v", err)
	}
	if err := decoder.Decode(value); err != nil {
		return nil, ErrInputDecoding.WithMessage(
			fmt.Sprintf("mapToObject: error decoding input: %v", err),
		)
	}
	return obj, nil
}

// isValidKey checks if a key is well-formed.
func isValidKey(key string) bool {
	// Check if key is valid UTF-8 and matches the allowed pattern.
	return utf8.ValidString(key) && allowedKeyRegex.MatchString(key)
}

// pickMap picks a map from the request using the provided APIFields. It filters
// out invalid keys.
func (h *MapInputHandler[Input]) pickMap(
	r *http.Request, apiFields APIFields,
) (map[string]any, error) {
	// Filter out invalid keys.
	cleanForm := make(map[string][]string)
	for key, vals := range r.Form {
		if isValidKey(key) {
			cleanForm[key] = vals
		}
	}

	// Obtain the MapFieldConfig from APIFields.
	mapFieldConfig, err := h.mapFieldConfigFromAPIFields(apiFields)
	if err != nil {
		return nil, err
	}

	// Otherwise, fall back to the standard objectpicker.
	return objectpicker.NewObjectPicker(
		urlencoder.NewURLEncoder(), h.conversionMap,
	).PickMap(r, mapFieldConfig)
}

// getValidator returns a new instance of the getValidator.
func (h *MapInputHandler[Input]) getValidator() *validate.Validate {
	return validate.NewValidate(h.customRules)
}

// mapFieldConfigFromAPIFields converts an APIFields to a MapFieldConfig.
func (h *MapInputHandler[Input]) mapFieldConfigFromAPIFields(
	apiFields APIFields,
) (*objectpicker.MapFieldConfig, error) {
	cfg := &objectpicker.MapFieldConfig{
		Fields: make(map[string]*objectpicker.MapFieldConfig),
	}
	// Convert each field to a MapFieldConfig.
	for _, field := range apiFields {
		fieldCfg := &objectpicker.MapFieldConfig{
			Source:       field.Source,
			ExpectedType: field.Type,
			DefaultValue: field.Default,
			Optional:     !field.Required,
		}
		// Convert any nested fields recursively.
		if len(field.Nested) > 0 {
			fields, err := h.mapFieldConfigFromAPIFields(field.Nested)
			if err != nil {
				return nil, err
			}
			fieldCfg.Fields = fields.Fields

		}
		cfg.Fields[field.APIName] = fieldCfg
	}

	return cfg, nil
}

// validate validates an input map against the provided APIFields.
func (h *MapInputHandler[Input]) validate(
	input map[string]any, apiFields APIFields,
) error {
	// Ensure required fields are present and validate each value.
	for _, field := range apiFields {
		val, exists := input[field.APIName]
		if !exists {
			if field.Required {
				return fmt.Errorf(
					"validate: field %q is required", field.APIName,
				)
			}
			continue
		}
		// If the field is nested, validate recursively.
		if field.Nested != nil {
			switch v := val.(type) {
			case map[string]any:
				if err := h.validate(v, field.Nested); err != nil {
					return fmt.Errorf(
						"validate: field %q: %w", field.APIName, err,
					)
				}
			case []any:
				for _, item := range v {
					err := h.validate(item.(map[string]any), field.Nested)
					if err != nil {
						return fmt.Errorf(
							"validate: field %q: %w", field.APIName, err,
						)
					}
				}
			default:
				return fmt.Errorf(
					"validate: field %q is not an object or an array",
					field.APIName,
				)
			}
		}
		// Validate non-object fields.
		if field.Validate != nil {
			validate, err := h.getValidator().FromRules(field.Validate)
			if err != nil {
				return fmt.Errorf(
					"validate: validation rule error for field %q: %w",
					field.APIName,
					err,
				)
			}
			if err := validate(val); err != nil {
				return fmt.Errorf(
					"validate: validation error for field %q: %w",
					field.APIName,
					err,
				)
			}
		}
	}
	return nil
}

// validateRules tests that the validation rules are valid.
func (h *MapInputHandler[Input]) validateRules(
	apiFields APIFields,
) error {
	for _, field := range apiFields {
		if field.Nested != nil {
			if err := h.validateRules(field.Nested); err != nil {
				return fmt.Errorf(
					"validateRules: field %s: %w", field.APIName, err,
				)
			}
		}
		if field.Validate != nil {
			_, err := h.getValidator().FromRules(field.Validate)
			if err != nil {
				return fmt.Errorf(
					"validateRules: validation rule error for field %s: %w",
					field.APIName,
					err,
				)
			}
		}
	}
	return nil
}
