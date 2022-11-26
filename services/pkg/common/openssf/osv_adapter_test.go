package openssf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Potential flaky test
func TestOsvExternalServiceApi(t *testing.T) {
	cases := []struct {
		name string

		// Inputs
		ecosystem, pkgName, pkgVersion string

		// Assertions
		isEmptyVulns bool
		knownVulnId  string
	}{
		{
			"Log4j in Maven Ecosystem",
			"Maven",
			"org.apache.logging.log4j:log4j-core",
			"2.16.0",

			false,
			"GHSA-8489-44mv-ggj8",
		},
		{
			"No such package in Ecosystem",
			"Maven",
			"no.such.pkg:no.pkg",
			"0.0.0",

			true,
			"",
		},
	}

	svc := NewOsvServiceAdapter(DefaultServiceAdapterConfig())

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			vulns, err := svc.QueryPackage(test.ecosystem,
				test.pkgName, test.pkgVersion)

			if test.isEmptyVulns {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.True(t, len(*vulns.Vulns) > 0)

				var knownVuln *OsvVulnerability = nil
				for _, v := range *vulns.Vulns {
					if *v.Id == test.knownVulnId {
						knownVuln = &v
					}
				}

				assert.NotNil(t, knownVuln)
			}
		})
	}
}
