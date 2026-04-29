// Package detect detects project type and framework usage to determine
// which oxlint plugins should be enabled.
package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// ProjectType represents the detected project framework/type.
type ProjectType string

const (
	// ProjectType values for detected project frameworks.
	ProjectTypeUnknown ProjectType = "unknown"
	ProjectTypeReact   ProjectType = "react"
	ProjectTypeNextJS  ProjectType = "nextjs"
	ProjectTypeVue     ProjectType = "vue"
	ProjectTypeNode    ProjectType = "node"
	ProjectTypePlainTS ProjectType = "typescript"
	ProjectTypeLibrary ProjectType = "library"
)

// Detector detects project type from the filesystem.
type Detector struct {
	rootDir string
}

// NewDetector creates a project detector for the given root directory.
func NewDetector(rootDir string) *Detector {
	return &Detector{rootDir: rootDir}
}

// Detect analyzes the project and returns the plugin configuration.
func (d *Detector) Detect() (profile.PluginConfig, []ProjectType, error) {
	pkg := d.readPackageJSON()
	deps := d.collectDependencies(pkg)

	types := d.detectProjectTypes(deps)
	pc := d.toPluginConfig(types, deps)

	return pc, types, nil
}

// detectProjectTypes determines project types from dependencies.
func (d *Detector) detectProjectTypes(deps map[string]bool) []ProjectType {
	var types []ProjectType

	if deps["next"] {
		types = append(types, ProjectTypeNextJS)
	}
	if deps["react"] || deps["react-dom"] {
		types = append(types, ProjectTypeReact)
	}
	if deps["vue"] {
		types = append(types, ProjectTypeVue)
	}
	if deps["jest"] {
		types = append(types, ProjectTypeNode) // Jest implies Node
	}
	if deps["vitest"] {
		// Vitest implies Node for test runner environment
		types = append(types, ProjectTypeNode)
	}

	if len(types) == 0 {
		if d.hasTSConfig() {
			types = append(types, ProjectTypePlainTS)
		}
		if d.hasNodeModules() || d.hasPackageJSON() {
			types = append(types, ProjectTypeNode)
		}
		if len(types) == 0 {
			types = append(types, ProjectTypeUnknown)
		}
	}

	return types
}

// toPluginConfig converts detected types to plugin configuration.
func (d *Detector) toPluginConfig(types []ProjectType, deps map[string]bool) profile.PluginConfig {
	pc := make(profile.PluginConfig)

	for _, t := range types {
		switch t {
		case ProjectTypeReact, ProjectTypeNextJS:
			pc[rule.PluginReact] = true
			pc[rule.PluginJSXA11y] = true
			pc[rule.PluginReactPerf] = true
		case ProjectTypeVue:
			pc[rule.PluginVue] = true
		case ProjectTypeNode:
			pc[rule.PluginNode] = true
		case ProjectTypeUnknown, ProjectTypePlainTS, ProjectTypeLibrary:
			// no additional plugins
		}
	}

	if deps["next"] {
		pc[rule.PluginNextJS] = true
	}
	if deps["jest"] {
		pc[rule.PluginJest] = true
	}
	if deps["vitest"] {
		pc[rule.PluginVitest] = true
	}
	if d.hasJSDocUsage(deps) {
		pc[rule.PluginJSDoc] = true
	}
	if d.hasImportUsage() {
		pc[rule.PluginImport] = true
	}
	if d.hasPromiseUsage(deps) {
		pc[rule.PluginPromise] = true
	}

	return pc
}

// packageJSON represents relevant fields from package.json.
type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// readPackageJSON reads and parses the project's package.json.
func (d *Detector) readPackageJSON() *packageJSON {
	path := filepath.Join(d.rootDir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	return &pkg
}

// collectDependencies merges dependencies and devDependencies into a set.
func (d *Detector) collectDependencies(pkg *packageJSON) map[string]bool {
	deps := make(map[string]bool)

	if pkg != nil {
		for name := range pkg.Dependencies {
			deps[name] = true
		}
		for name := range pkg.DevDependencies {
			deps[name] = true
		}
	}

	return deps
}

func (d *Detector) hasTSConfig() bool {
	_, err := os.Stat(filepath.Join(d.rootDir, "tsconfig.json"))
	return err == nil
}

func (d *Detector) hasNodeModules() bool {
	_, err := os.Stat(filepath.Join(d.rootDir, "node_modules"))
	return err == nil
}

func (d *Detector) hasPackageJSON() bool {
	_, err := os.Stat(filepath.Join(d.rootDir, "package.json"))
	return err == nil
}

func (d *Detector) hasJSDocUsage(deps map[string]bool) bool {
	return deps["jsdoc"] || deps["documentation"]
}

func (d *Detector) hasImportUsage() bool {
	matches, _ := filepath.Glob(filepath.Join(d.rootDir, "*.mjs"))
	if len(matches) > 0 {
		return true
	}

	matches, _ = filepath.Glob(filepath.Join(d.rootDir, "*.mts"))
	return len(matches) > 0
}

func (d *Detector) hasPromiseUsage(deps map[string]bool) bool {
	promiseDeps := []string{
		"bluebird",
		"es6-promise",
		"promise",
		"q",
		"rsvp",
		"promise-polyfill",
		"core-js",
	}
	for _, dep := range promiseDeps {
		if deps[dep] {
			return true
		}
	}
	for dep := range deps {
		if strings.Contains(dep, "promise") {
			return true
		}
	}
	return false
}

// FormatTypes returns a human-readable string of detected types.
func FormatTypes(types []ProjectType) string {
	names := make([]string, 0, len(types))
	for _, t := range types {
		names = append(names, string(t))
	}
	if len(names) == 0 {
		return "unknown"
	}
	return strings.Join(names, ", ")
}
