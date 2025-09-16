package crud

import (
	"github.com/aatuh/pureapi-framework/api/input"
	"github.com/aatuh/pureapi-framework/util/inpututil"
)

// CRUDEntitiesToOutputs converts a slice of entities to a slice of outputs.
//
// Parameters:
//   - entities: The entities to convert.
//   - apiFields: The API fields to use for mapping.
//   - tableName: The name of the table to use for mapping.
//
// Returns:
//   - []Output: The outputs.
func CRUDEntitiesToOutputs[Entity any, Output any](
	entities []Entity, apiFields input.APIFields, tableName string,
) []Output {
	anyEntities := []any{}
	for i := range entities {
		anyEntities = append(anyEntities, entities[i])
	}
	anyOutputs := inpututil.MustMapEntitiesToOutput(
		anyEntities,
		func() any {
			return new(Output)
		},
		inpututil.GetAPIToDBMap(apiFields, tableName),
	)
	outputs := []Output{}
	for i := range anyOutputs {
		outputs = append(outputs, *anyOutputs[i].(*Output))
	}
	return outputs
}

// CRUDEntitiesToOutput converts a slice of entities to a single output.
//
// Parameters:
//   - entities: The entities to convert.
//   - apiFields: The API fields to use for mapping.
//   - tableName: The name of the table to use for mapping.
//
// Returns:
//   - Output: The output.
func CRUDEntitiesToOutput[Entity any, Output any](
	entities []Entity, apiFields input.APIFields, tableName string,
) *Output {
	outputs := CRUDEntitiesToOutputs[Entity, Output](
		entities, apiFields, tableName,
	)
	if len(outputs) == 0 {
		return nil
	}
	return &outputs[0]
}
