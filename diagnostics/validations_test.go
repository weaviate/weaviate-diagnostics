package diagnostics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/schema"
	"github.com/weaviate/weaviate/entities/models"
)

func TestBadVectorConfig(t *testing.T) {
	var dump = schema.Dump{}
	dump.Classes = []*models.Class{
		{
			Class: "Test",
			VectorIndexConfig: map[string]interface{}{
				"efConstruction":        8.0,
				"maxConnections":        4.0,
				"vectorCacheMaxObjects": 1e12,
			},
		},
	}
	assumed := []Validation{
		{
			Message: "efConstruction=8 for class Test is too low",
		},
		{
			Message: "maxConnections=4 for class Test is too low",
		},
	}
	validations := validateBadVectorIndexConfig(&dump)
	assert.Equal(t, assumed, validations)

}
