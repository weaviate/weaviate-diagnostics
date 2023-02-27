package diagnostics

import (
	"fmt"

	"github.com/weaviate/weaviate-go-client/v4/weaviate/schema"
)

type Validation struct {
	Message string
}

func validateBadVectorIndexConfig(schema *schema.Dump) []Validation {
	var validations []Validation

	for _, class := range schema.Classes {
		vectorIndexConfig, ok := class.VectorIndexConfig.(map[string]interface{})
		if !ok {
			continue
		}
		if vectorIndexConfig["efConstruction"].(float64) < 16 {
			validations = append(validations, Validation{
				Message: fmt.Sprintf("efConstruction=%.0f for class %s is too low", vectorIndexConfig["efConstruction"].(float64), class.Class),
			})
		}
		if vectorIndexConfig["maxConnections"].(float64) < 8 {
			validations = append(validations, Validation{
				Message: fmt.Sprintf("maxConnections=%.0f for class %s is too low", vectorIndexConfig["maxConnections"].(float64), class.Class),
			})
		}
		if vectorIndexConfig["vectorCacheMaxObjects"].(float64) != 1e12 {
			validations = append(validations, Validation{
				Message: fmt.Sprintf("vectorCacheMaxObjects=%.0f for class %s is not 1e12", vectorIndexConfig["vectorCacheMaxObjects"].(float64), class.Class),
			})
		}
	}

	return validations
}

func validateSchema(schema *schema.Dump) []Validation {
	var validations []Validation

	validations = append(validations, validateBadVectorIndexConfig(schema)...)

	return validations
}
