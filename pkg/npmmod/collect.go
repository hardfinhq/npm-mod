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
//       * `CollectPackages{}.Visit` satisfies `VisitorFunc`.
var (
	_ VisitorFunc = (&CollectPackages{}).Visit
)

// CollectPackages produces a vis]]]itor function that collects informationa about
// all packages in a `package-lock.json` (in particular, about the `resolved`
// URLs).
type CollectPackages struct {
	ByNodeModulesPath map[string]RegistryPackage
	ByURL             map[string]RegistryPackage
	ParentKey         string
}

// Visit is a visitor function that **tracks** a package `resolved` URL.
func (cp *CollectPackages) Visit(deps *ordered.OrderedMap, k string, v any) error {
	name := k
	if cp.ParentKey == "packages" && name == "" {
		return nil
	}

	packageMap, ok := v.(*ordered.OrderedMap)
	if !ok {
		return fmt.Errorf("package %q does not point at a map", name)
	}

	resolvedAny := packageMap.Get("resolved")
	resolved, ok := resolvedAny.(string)
	if !ok {
		return fmt.Errorf(`package %q "resolved" is not a string`, name)
	}

	integrityAny := packageMap.Get("integrity")
	integrity, ok := integrityAny.(string)
	if !ok {
		return fmt.Errorf(`package %q "integrity" is not a string`, name)
	}

	algorithm, hash, err := splitIntegrity(integrity)
	if err != nil {
		return err
	}

	rp := RegistryPackage{
		URL:       resolved,
		Algorithm: algorithm,
		Hash:      hash,
	}

	existing, ok := cp.ByURL[rp.URL]
	if ok && !rp.Equal(existing) {
		return fmt.Errorf("conflict with existing package; existing: %#v, to add: %#v", existing, rp)
	}
	cp.ByURL[rp.URL] = rp

	if cp.ParentKey == "packages" {
		cp.ByNodeModulesPath[name] = rp
	}

	return nil
}
