package certificate

import (
	"testing"

	"github.com/altinity/terraform-provider-altinitycloud/internal/provider/schematest"
)

func TestCertificateResourceModelMatchesSchema(t *testing.T) {
	schematest.AssertResourceModelMatchesSchema(t, &CertificateResource{}, &CertificateResourceModel{})
}
