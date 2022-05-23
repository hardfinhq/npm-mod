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

// PackageLockReplaceDependencies iterates through all entries in the
// `package-lock.json` packages and dependencies maps and then replaces each
// package version based on a "replace" function.
func PackageLockReplaceDependencies(packageLock *ordered.OrderedMap, replace ReplaceFunc) error {
	rr := ReplaceResolved{Replace: replace, ParentKey: "packages"}
	err := walkPackageLockPackages(packageLock, rr.Visit)
	if err != nil {
		return err
	}

	rr = ReplaceResolved{Replace: replace, ParentKey: "dependencies"}
	return walkPackageLockDependencies(packageLock, rr.Visit)
}

// PackageLockExtractDependencies iterates through all entries in the
// `package-lock.json` packages and dependencies maps and extracts the
// "resolved" URL.
func PackageLockExtractDependencies(packageLock *ordered.OrderedMap) (map[string]RegistryPackage, map[string]RegistryPackage, error) {
	byNodeModulesPath := map[string]RegistryPackage{}
	byURL := map[string]RegistryPackage{}
	cp := CollectPackages{ByNodeModulesPath: byNodeModulesPath, ByURL: byURL, ParentKey: "packages"}
	err := walkPackageLockPackages(packageLock, cp.Visit)
	if err != nil {
		return nil, nil, err
	}

	cp = CollectPackages{ByURL: byURL, ParentKey: "dependencies"}
	err = walkPackageLockDependencies(packageLock, cp.Visit)
	if err != nil {
		return nil, nil, err
	}

	return byNodeModulesPath, byURL, nil
}

// walkPackageLockPackages iterates through all entries in the
// `package-lock.json` packages map and then replaces each package version based
// on a "replace" function.
func walkPackageLockPackages(packageLock *ordered.OrderedMap, visitor VisitorFunc) error {
	packagesAny, ok := packageLock.GetValue("packages")
	if !ok {
		// Early exit if the dependencies key is absent
		return nil
	}

	packages, ok := packagesAny.(*ordered.OrderedMap)
	if !ok {
		return errors.New(`"packages" key is present, but not a map`)
	}

	nextPair := packages.EntriesIter()
	// NOTE: Use a bounded for loop to avoid an accidental infinite loop.
	loopComplete := false
	for i := 0; i < 10000; i++ {
		pair, ok := nextPair()
		if !ok {
			loopComplete = true
			break
		}

		err := visitor(packages, pair.Key, pair.Value)
		if err != nil {
			return err
		}
	}

	if !loopComplete {
		return errors.New("loop over packages never terminated")
	}

	return nil
}

// walkPackageLockDependencies iterates through all entries in the
// `package-lock.json` dependencies map (and so on recursively) and then
// replaces each package version based on a "replace" function.
//
// The `hasDependencies` map can either be the root `package-lock.json` or a
// child of it.
func walkPackageLockDependencies(hasDependencies *ordered.OrderedMap, visitor VisitorFunc) error {
	depsAny, ok := hasDependencies.GetValue("dependencies")
	if !ok {
		// Early exit if the dependencies key is absent
		return nil
	}

	deps, ok := depsAny.(*ordered.OrderedMap)
	if !ok {
		return errors.New(`"dependencies" key is present, but not a map`)
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

		dependencyMap, ok := pair.Value.(*ordered.OrderedMap)
		if !ok {
			return fmt.Errorf("dependency %q does not point at a map", pair.Key)
		}

		// Recursively apply this as well (if the dependency has a
		// `dependencies` key).
		err = walkPackageLockDependencies(dependencyMap, visitor)
		if err != nil {
			return err
		}
	}

	if !loopComplete {
		return errors.New("loop over dependencies never terminated")
	}

	return nil
}
