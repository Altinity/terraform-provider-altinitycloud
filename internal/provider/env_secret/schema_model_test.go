package secret

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestSecretResourceModelMatchesSchema(t *testing.T) {
	schematest.AssertResourceModelMatchesSchema(t, &SecretResource{}, &SecretResourceModel{})
}
