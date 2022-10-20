package openssf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Potential flaky test
func TestOsvExternalServiceApi(t *testing.T) {
	svc := NewOsvServiceAdapter(DefaultServiceAdapterConfig())
	vulns, err := svc.QueryPackage("Maven", "org.apache.logging.log4j:log4j-core", "2.16.0")

	assert.Nil(t, err)
	assert.True(t, len(*vulns.Vulns) > 0)

	var knownVuln *OsvVulnerability = nil
	for _, v := range *vulns.Vulns {
		if *v.Id == "GHSA-8489-44mv-ggj8" {
			knownVuln = &v
		}
	}

	assert.NotNil(t, knownVuln)

	assert.True(t, strings.Contains(*knownVuln.Details, "Log4j2"))
	assert.False(t, strings.Contains(*knownVuln.Details, "Log4j31337"))
}
