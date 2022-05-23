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
	"errors"
	"fmt"

	"github.com/hardfinhq/npm-mod/pkg/ordered"
)

// PackageJSONReplaceDependencies iterates through all entries in the
// `package.json` dependencies maps and then replaces each package version based
// on a "replace" function.
func PackageJSONReplaceDependencies(packageJSON *ordered.OrderedMap, replace ReplacePairFunc) error {
	rp := ReplaceDependency{Replace: replace}

	err := walkPackageJSON(packageJSON, "dependencies", rp.Visit)
	if err != nil {
		return err
	}

	err = walkPackageJSON(packageJSON, "devDependencies", rp.Visit)
	if err != nil {
		return err
	}

	err = walkPackageJSON(packageJSON, "peerDependencies", rp.Visit)
	return err
}

// walkPackageJSON iterates through all entries in a `package.json` dependencies
// map (i.e. `dependencies`, `devDependencies` or `peerDependencies`) and then
// applies a "visitor" function to each key / value pair in the map
func walkPackageJSON(packageJSON *ordered.OrderedMap, key string, visitor VisitorFunc) error {
	depsAny, ok := packageJSON.GetValue(key)
	if !ok {
		// Early exit if the dependencies key is absent
		return nil
	}

	deps, ok := depsAny.(*ordered.OrderedMap)
	if !ok {
		return fmt.Errorf("%q key is present, but not a map", key)
	}

	nextPair := deps.EntriesIter()
	// NOTE: Use a bounded for loop to avoid an accidental infinite loop.
	loopComplete := false
	for i := 0; i < 10000; i++ {
		pair, ok := nextPair()
		if !ok {
			loopComplete = true
			break
		}

		err := visitor(deps, pair.Key, pair.Value)
		if err != nil {
			return err
		}
	}

	if !loopComplete {
		return errors.New("loop over dependencies never terminated")
	}

	return nil
}
