// Copyright 2022 Hardfin, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package npmmod

import (
	"fmt"

	"github.com/hardfinhq/npm-mod/pkg/ordered"
)

// NOTE: Ensure that
//       * `ReplaceDependency{}.Visit` satisfies `VisitorFunc`.
//       * `ReplaceResolved{}.Visit` satisfies `VisitorFunc`.
//       * `PackageJSONReplace{}.Replace` satisfies `ReplacePairFunc`.
//       * `PackageLockReplace{}.Replace` satisfies `ReplaceFunc`.
var (
	_ VisitorFunc     = (&ReplaceDependency{}).Visit
	_ VisitorFunc     = (&ReplaceResolved{}).Visit
	_ ReplacePairFunc = (&PackageJSONReplace{}).Replace
	_ ReplaceFunc     = (&PackageLockReplace{}).Replace
)

// ReplaceDependency produces a visitor function that **replaces** a
// package version based on a `replace` function.
type ReplaceDependency struct {
	Replace ReplacePairFunc
}

// Visit is a visitor function that **replaces** a package version based on a
// `replace` function.
func (rd *ReplaceDependency) Visit(deps *ordered.OrderedMap, k string, v any) error {
	packageName := k
	packageVersion, ok := v.(string)
	if !ok {
		return fmt.Errorf("dependency %q is not a string", packageName)
	}

	newPackageVersion := rd.Replace(packageName, packageVersion)
	deps.Set(packageName, newPackageVersion)
	return nil
}

// ReplaceResolved produces a visitor function that **replaces** a package
// `resolved` (and `version`) key based on a `replace` function.
type ReplaceResolved struct {
	Replace   ReplaceFunc
	ParentKey string
}

// Visit is a visitor function that **replaces** a package `resolved` (and
// `version`) key based on a `replace` function.
func (rr *ReplaceResolved) Visit(deps *ordered.OrderedMap, k string, v any) error {
	name := k
	if rr.ParentKey == "packages" && name == "" {
		return nil
	}

	m, ok := v.(*ordered.OrderedMap)
	if !ok {
		return fmt.Errorf("package %q does not point at a map", name)
	}

	resolvedAny := m.Get("resolved")
	resolved, ok := resolvedAny.(string)
	if !ok {
		return fmt.Errorf(`package %q "resolved" is not a string`, name)
	}

	newResolved := rr.Replace(resolved)
	m.Set("resolved", newResolved)
	m.Set("version", newResolved)
	return nil
}

// PackageJSONReplace provides a `replace` helper that replaces a `package.json`
// package version with a local `file:` reference.
type PackageJSONReplace struct {
	ByNodeModulesPath map[string]RegistryPackage
}

// Replace replaces a `package.json` package version with a local `file:`
// reference. In the case that the package name or version can't be matched
// or the filename can't be determined, this just returns the `version`.
func (pjr *PackageJSONReplace) Replace(name, version string) string {
	key := fmt.Sprintf("node_modules/%s", name)
	rp, ok := pjr.ByNodeModulesPath[key]
	if !ok {
		// NOTE: This isn't great, the `ReplacePairFunc` should probably allow
		//       an error too.
		return version
	}

	filename, err := rp.Filename()
	if err != nil {
		// NOTE: This isn't great, the `ReplacePairFunc` should probably allow
		//       an error too.
		return version
	}

	// NOTE: We only validate the key in `ByNodeModulesPath` but don't check
	//       anything about the specified version / version range.
	return fmt.Sprintf("file:vendor/%s", filename)
}

// PackageLockReplace provides a `replace` helper that replaces a `resolved` URL
// with a local `file:` reference.
type PackageLockReplace struct {
	ByURL map[string]RegistryPackage
}

// Replace replaces a `resolved` URL with a local `file:` reference. In the case
// that the URL can't be matched or the filename can't be determined, this just
// returns the `resolved`.
func (plr *PackageLockReplace) Replace(resolved string) string {
	rp, ok := plr.ByURL[resolved]
	if !ok {
		// NOTE: This isn't great, the `ReplaceFunc` should probably allow an
		//       error too.
		return resolved
	}

	filename, err := rp.Filename()
	if err != nil {
		// NOTE: This isn't great, the `ReplaceFunc` should probably allow an
		//       error too.
		return resolved
	}

	return fmt.Sprintf("file:vendor/%s", filename)
}
