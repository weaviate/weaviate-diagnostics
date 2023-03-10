package diagnostics

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/weaviate/weaviate-go-client/v4/weaviate/schema"
)

type Validation struct {
	Message string
}

func validateEnvironmentVariables() []Validation {
	var validations []Validation

	if os.Getenv("GOMEMLIMIT") == "" {
		validations = append(validations, Validation{
			Message: "<code>GOMEMLIMIT</code> is not set",
		})
	}

	if os.Getenv("QUERY_MAXIMUM_RESULTS") != "" {
		max_results, err := strconv.ParseInt(os.Getenv("QUERY_MAXIMUM_RESULTS"), 10, 64)
		if err != nil {
			validations = append(validations, Validation{
				Message: fmt.Sprintf("QUERY_MAXIMUM_RESULTS is not a number: %s", err),
			})
		} else if max_results > 10000 {
			validations = append(validations, Validation{
				Message: fmt.Sprintf("QUERY_MAXIMUM_RESULTS is set high: %d", max_results),
			})
		}
	}

	if os.Getenv("GOGC") != "" {
		max_results, err := strconv.ParseInt(os.Getenv("GOGC"), 10, 64)
		if err != nil {
			validations = append(validations, Validation{
				Message: fmt.Sprintf("GOGC is not a number: %s", err),
			})
		} else if max_results > 100 {
			validations = append(validations, Validation{
				Message: fmt.Sprintf("GOGC is set high: %d", max_results),
			})
		}
	}

	if strings.ToLower(os.Getenv("REINDEX_VECTOR_DIMENSIONS_AT_STARTUP")) == "true" || os.Getenv("REINDEX_VECTOR_DIMENSIONS_AT_STARTUP") == "1" {
		validations = append(validations, Validation{
			Message: "<code>REINDEX_VECTOR_DIMENSIONS_AT_STARTUP</code> is set to true. This is likely not needed if running on a recent version of Weaviate.",
		})
	}

	return validations
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

func validate(schema *schema.Dump) []Validation {
	var validations []Validation

	validations = append(validations, validateBadVectorIndexConfig(schema)...)
	validations = append(validations, validateEnvironmentVariables()...)

	return validations
}
