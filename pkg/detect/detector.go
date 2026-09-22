// Package detect detects project type and framework usage to determine
// which oxlint plugins should be enabled.
package detect

import (
	"encoding/json/v2"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// ProjectType represents the detected project framework/type.
type ProjectType string

// ProjectTypeUnknown and other project type constants represent detected framework categories.
const (
	ProjectTypeUnknown ProjectType = "unknown"    // no recognized framework
	ProjectTypeReact   ProjectType = "react"      // React SPA
	ProjectTypeNextJS  ProjectType = "nextjs"     // Next.js SSR
	ProjectTypeVue     ProjectType = "vue"        // Vue SPA
	ProjectTypeNode    ProjectType = "node"       // Node.js backend
	ProjectTypeTest    ProjectType = "test"       // test framework (vitest/jest)
	ProjectTypePlainTS ProjectType = "typescript" // TypeScript without framework
	ProjectTypeLibrary ProjectType = "library"    // shared library
)

// Detector detects project type from the filesystem.
type Detector struct {
	rootDir string

	depsOnce sync.Once
	deps     map[string]bool
}

// NewDetector creates a project detector for the given root directory.
func NewDetector(rootDir string) *Detector {
	return &Detector{
		rootDir:  rootDir,
		depsOnce: sync.Once{},
		deps:     make(map[string]bool),
	}
}

// Detect analyzes the project and returns the plugin configuration.
func (d *Detector) Detect() (profile.PluginConfig, []ProjectType, error) {
	deps := d.dependencies()

	types := d.detectProjectTypes(deps)
	pc := d.toPluginConfig(types, deps)

	return pc, types, nil
}

// DetectExternalPlugins returns the known oxlint JS plugins (loaded at
// runtime via the "jsPlugins" config key, e.g. @shadcn/lint) installed in
// this project's dependencies or devDependencies. Presence of the package is
// the signal: an installed plugin should be registered in the generated
// config, while its rules stay off (design-system policy is the project's
// choice, not ours).
func (d *Detector) DetectExternalPlugins() []rule.ExternalPlugin {
	deps := d.dependencies()

	var found []rule.ExternalPlugin

	for _, external := range rule.KnownExternalPlugins() {
		if deps[external.Package] {
			found = append(found, external)
		}
	}

	slices.SortFunc(found, func(a, b rule.ExternalPlugin) int {
		return strings.Compare(a.Package, b.Package)
	})

	return found
}

// dependencies returns the merged dependency set, reading package.json at
// most once per Detector so Detect and DetectExternalPlugins share the work.
func (d *Detector) dependencies() map[string]bool {
	d.depsOnce.Do(func() {
		d.deps = d.collectDependencies(d.readPackageJSON())
	})

	return d.deps
}

// depTypeRules maps dependency names to project types.
// If any dep in a rule's list is present, the type is detected.
var depTypeRules = []struct { //nolint:gochecknoglobals // immutable lookup table
	deps []string
	typ  ProjectType
}{
	{[]string{"next"}, ProjectTypeNextJS},
	{[]string{"react", "react-dom"}, ProjectTypeReact},
	{[]string{"vue"}, ProjectTypeVue},
	{[]string{"jest"}, ProjectTypeTest},
	{[]string{"vitest"}, ProjectTypeTest},
}

// detectProjectTypes determines project types from dependencies.
func (d *Detector) detectProjectTypes(deps map[string]bool) []ProjectType {
	var types []ProjectType

	for _, rule := range depTypeRules {
		if anyDep(deps, rule.deps) {
			types = append(types, rule.typ)
		}
	}

	if len(types) == 0 {
		types = d.inferFallbackTypes()
	}

	return types
}

func anyDep(deps map[string]bool, names []string) bool {
	for _, n := range names {
		if deps[n] {
			return true
		}
	}

	return false
}

func (d *Detector) inferFallbackTypes() []ProjectType {
	var types []ProjectType
	if d.hasTSConfig() {
		types = append(types, ProjectTypePlainTS)
	}

	if d.hasNodeModules() || d.hasPackageJSON() {
		types = append(types, ProjectTypeNode)
	}

	if len(types) == 0 {
		types = append(types, ProjectTypeUnknown)
	}

	return types
}

// toPluginConfig converts detected types to plugin configuration.
func (d *Detector) toPluginConfig(types []ProjectType, deps map[string]bool) profile.PluginConfig {
	pc := make(profile.PluginConfig)
	d.applyTypePlugins(types, pc)
	d.applyDepPlugins(deps, pc)

	return pc
}

func (d *Detector) applyTypePlugins(types []ProjectType, pluginConfig profile.PluginConfig) {
	for _, t := range types {
		switch t {
		case ProjectTypeReact, ProjectTypeNextJS:
			pluginConfig[rule.PluginReact] = true
			pluginConfig[rule.PluginJSXA11y] = true
			pluginConfig[rule.PluginReactPerf] = true
		case ProjectTypeVue:
			pluginConfig[rule.PluginVue] = true
		case ProjectTypeTest:
			pluginConfig[rule.PluginNode] = true
		case ProjectTypeNode:
			pluginConfig[rule.PluginNode] = true
		case ProjectTypeUnknown, ProjectTypePlainTS, ProjectTypeLibrary:
			// no additional plugins
		}
	}
}

// depPluginRules maps dependency names to plugins that should be enabled.
// If any dep in a rule's list is present, the plugin is enabled.
var depPluginRules = []struct { //nolint:gochecknoglobals // immutable lookup table
	deps   []string
	plugin rule.Plugin
}{
	{[]string{"next"}, rule.PluginNextJS},
	{[]string{"jest"}, rule.PluginJest},
	{[]string{"vitest"}, rule.PluginVitest},
	{[]string{"jsdoc", "documentation"}, rule.PluginJSDoc},
	{
		[]string{"bluebird", "es6-promise", "promise", "q", "rsvp", "promise-polyfill", "core-js"},
		rule.PluginPromise,
	},
}

func (d *Detector) applyDepPlugins(deps map[string]bool, pluginConfig profile.PluginConfig) {
	for _, r := range depPluginRules {
		if anyDep(deps, r.deps) {
			pluginConfig[r.plugin] = true
		}
	}

	if d.hasPromiseSubstringDep(deps) {
		pluginConfig[rule.PluginPromise] = true
	}

	if d.hasImportUsage() {
		pluginConfig[rule.PluginImport] = true
	}
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
		if !os.IsNotExist(err) {
			slog.Warn("cannot read package.json", "path", path, "error", err)
		}

		return nil
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		slog.Warn(
			"malformed package.json, skipping dependency detection",
			"path",
			path,
			"error",
			err,
		)

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

func (d *Detector) hasFile(filename string) bool {
	_, err := os.Stat(filepath.Join(d.rootDir, filename))

	return err == nil
}

func (d *Detector) hasTSConfig() bool {
	return d.hasFile("tsconfig.json")
}

func (d *Detector) hasNodeModules() bool {
	return d.hasFile("node_modules")
}

func (d *Detector) hasPackageJSON() bool {
	return d.hasFile("package.json")
}

func (d *Detector) hasImportUsage() bool {
	return d.hasGlob("*.mjs") || d.hasGlob("*.mts")
}

func (d *Detector) hasGlob(pattern string) bool {
	matches, _ := filepath.Glob(filepath.Join(d.rootDir, pattern))

	return len(matches) > 0
}

func (d *Detector) hasPromiseSubstringDep(deps map[string]bool) bool {
	for dep := range deps {
		if strings.Contains(dep, "promise") {
			return true
		}
	}

	return false
}

// FormatTypes returns a human-readable string of detected types.
func FormatTypes(types []ProjectType) string {
	// Deliberate clone of the ~string-to-string conversion loop (stdlib has
	// no slices.Map); a shared generic helper would over-couple.
	names := make([]string, 0, len(types))
	for _, t := range types {
		names = append(names, string(t))
	}

	if len(names) == 0 {
		return "unknown"
	}

	return strings.Join(names, ", ")
}
