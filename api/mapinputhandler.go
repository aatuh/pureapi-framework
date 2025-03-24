package api

import (
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/pureapi/pureapi-framework/input"
)

// FieldError represents a field-level validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrValidationData contains a list of field-level validation errors
type ErrValidationData struct {
	Errors []FieldError `json:"errors"`
}

// mapInputHandler handles the input of a request as a map.
type mapInputHandler[Input any] struct {
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
//
// Returns:
//   - *MapInputHandler: The new MapInputHandler.
func NewMapInputHandler[Input any](
	apiFields APIFields,
	conversionMap map[string]func(any) any,
	customRules map[string]func(any) error,
	inputFactoryFn func() *Input,
) *mapInputHandler[Input] {
	inputHandler := &mapInputHandler[Input]{
		apiFields:      apiFields,
		conversionMap:  conversionMap,
		customRules:    customRules,
		inputFactoryFn: inputFactoryFn,
	}

	// Validate APIFields.
	// TODO: Could cache results from these calls.
	err := inputHandler.testValidationRules(apiFields)
	if err != nil {
		panic(err)
	}
	_, err = inputHandler.mapFieldConfigFromAPIFields(apiFields)
	if err != nil {
		panic(err)
	}

	return inputHandler
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
func (h *mapInputHandler[Input]) Handle(
	w http.ResponseWriter, r *http.Request,
) (*Input, error) {
	// Pick input as map.
	dataMap, err := h.pickMap(r, h.apiFields)
	if err != nil {
		return nil, err
	}
	// Validate the map.
	if err := h.validateMap(dataMap, h.apiFields); err != nil {
		return nil, input.ErrValidation.WithData(ErrValidationData{
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
func (h *mapInputHandler[Input]) mapToObject(
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
		return nil, input.ErrInputDecoding.WithMessage(err.Error())
	}
	return obj, nil
}

// pickMap picks a map from the request using the provided APIFields.
func (h *mapInputHandler[Input]) pickMap(
	r *http.Request, apiFields APIFields,
) (map[string]any, error) {
	// Convert the APIFields to a MapFieldConfig.
	mapFieldConfig, err := h.mapFieldConfigFromAPIFields(apiFields)
	if err != nil {
		return nil, err
	}
	// Pick the map.
	return input.NewObjectPicker(input.NewURLEncoder(), h.conversionMap).
		PickMap(r, mapFieldConfig)
}

// getValidator returns a new instance of the getValidator.
func (h *mapInputHandler[Input]) getValidator() *input.Validate {
	return input.NewValidate(h.customRules)
}

// mapFieldConfigFromAPIFields converts an APIFields to a MapFieldConfig.
func (h *mapInputHandler[Input]) mapFieldConfigFromAPIFields(
	apiFields APIFields,
) (*input.MapFieldConfig, error) {
	cfg := &input.MapFieldConfig{
		Fields: make(map[string]*input.MapFieldConfig),
	}
	// Convert each field to a MapFieldConfig.
	for _, field := range apiFields {
		// Convert field to MapFieldConfig.
		// var expectedType string
		// if len(field.Validate) != 0 {
		// 	expectedType = field.Type
		// 	if expectedType == "" {
		// 		return nil, fmt.Errorf(
		// 			"type must be set for field %q", field.APIName,
		// 		)
		// 	}
		// }
		fieldCfg := &input.MapFieldConfig{
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

// TODO: Bug if same named fields (in nested fields?).
// validateMap validates an input map against the provided APIFields.
func (h *mapInputHandler[Input]) validateMap(
	input map[string]any, apiFields APIFields,
) error {
	// Ensure required fields are present and validate each value.
	for _, field := range apiFields {
		val, exists := input[field.APIName]
		if !exists {
			if field.Required {
				return fmt.Errorf("field %q is required", field.APIName)
			}
			continue
		}
		// If the field is nested, validate recursively.
		if field.Nested != nil {
			switch v := val.(type) {
			case map[string]any:
				if err := h.validateMap(v, field.Nested); err != nil {
					return fmt.Errorf("field %q: %w", field.APIName, err)
				}
			case []any:
				for _, item := range v {
					err := h.validateMap(item.(map[string]any), field.Nested)
					if err != nil {
						return fmt.Errorf("field %q: %w", field.APIName, err)
					}
				}
			default:
				return fmt.Errorf(
					"field %q is not an object or an array", field.APIName,
				)
			}
		}
		// For non-object fields, run the validation function.
		if field.Validate != nil {
			validate, err := h.getValidator().FromRules(field.Validate)
			if err != nil {
				return fmt.Errorf(
					"validation rule error for field %q: %w",
					field.APIName,
					err,
				)
			}
			if err := validate(val); err != nil {
				return fmt.Errorf(
					"validation error for field %q: %w", field.APIName, err,
				)
			}
		}
	}
	return nil
}

// testValidationRules tests that the validation rules are valid.
func (h *mapInputHandler[Input]) testValidationRules(
	apiFields APIFields,
) error {
	for _, field := range apiFields {
		if field.Nested != nil {
			if err := h.testValidationRules(field.Nested); err != nil {
				return fmt.Errorf("field %q: %w", field.APIName, err)
			}
		}
		if field.Validate != nil {
			_, err := h.getValidator().FromRules(field.Validate)
			if err != nil {
				return fmt.Errorf(
					"validation rule error for field %q: %w",
					field.APIName,
					err,
				)
			}
		}
	}
	return nil
}
