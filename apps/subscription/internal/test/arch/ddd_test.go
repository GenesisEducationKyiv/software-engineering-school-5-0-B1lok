//go:build arch
// +build arch

package arch

import (
	"github.com/matthewmcnew/archtest"
	"path"
	"testing"
)

const modulePath = "subscription-service/internal"

func TestDomain_ShouldNotDependOnOtherLayers(t *testing.T) {
	archtest.Package(t, path.Join(modulePath, "domain", "...")).
		ShouldNotDependOn(
			path.Join(modulePath, "infrastructure", "..."),
			path.Join(modulePath, "interface", "..."),
			path.Join(modulePath, "application", "..."),
		)
}

func TestApplication_ShouldNotDependOnInfrastructureOrInterface(t *testing.T) {
	archtest.Package(t, path.Join(modulePath, "application", "...")).
		ShouldNotDependOn(
			path.Join(modulePath, "infrastructure", "..."),
			path.Join(modulePath, "interface", "..."),
		)
}

func TestInfrastructure_ShouldNotDependOnInterface(t *testing.T) {
	archtest.Package(t, path.Join(modulePath, "infrastructure", "...")).
		ShouldNotDependOn(path.Join(modulePath, "interface", "..."))
}
