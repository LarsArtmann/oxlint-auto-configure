package config

import (
	"fmt"
	"strings"

	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// GenerateProjectConfig creates an OxlintConfig for the given project.
// This is the core business logic separated from CLI concerns.
func GenerateProjectConfig(
	p profile.Profile,
	reg *rule.Registry,
	pluginConfig profile.PluginConfig,
	projectTypes []detect.ProjectType,
) (*OxlintConfig, error) {
	if !p.IsValid() {
		return nil, fmt.Errorf("%w %q: choose from %s",
			ErrInvalidProfile, p, strings.Join(profile.AllProfileNames(), ", "))
	}

	cat := profile.NewCategorizer(p, pluginConfig)
	gen := NewGenerator(cat, reg, projectTypes)

	if p == profile.ProfileMaximalTypesafe {
		return gen.GenerateMaximal(), nil
	}

	return gen.Generate(), nil
}

