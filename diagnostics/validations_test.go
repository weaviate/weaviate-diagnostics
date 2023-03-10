package diagnostics

import (
	"os"
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

func TestEnvironmentVariables(t *testing.T) {

	err := os.Setenv("QUERY_MAXIMUM_RESULTS", "10001")
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("GOGC", "200")
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("REINDEX_VECTOR_DIMENSIONS_AT_STARTUP", "true")
	if err != nil {
		t.Fatal(err)
	}

	assumed := []Validation{
		{
			Message: "<code>GOMEMLIMIT</code> is not set",
		},
		{
			Message: "<code>QUERY_MAXIMUM_RESULTS</code> is set high: 10001",
		},
		{
			Message: "<code>GOGC</code> is set high: 200",
		},
		{
			Message: "<code>REINDEX_VECTOR_DIMENSIONS_AT_STARTUP</code> is set to true. This is likely not needed if running on a recent version of Weaviate.",
		},
	}
	validations := validateEnvironmentVariables()
	assert.Equal(t, assumed, validations)
}
