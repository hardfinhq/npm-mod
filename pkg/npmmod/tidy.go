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
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/hardfinhq/npm-mod/pkg/ordered"
)

const (
	tidyFileVersion = "22.05"
)

// TidyFile represents a `.npm-mod.tidy.json`
type TidyFile struct {
	Version         string            `json:"version"`
	PackageJSON     []byte            `json:"package.json"`
	PackageLockJSON []byte            `json:"package-lock.json"`
	Packages        []RegistryPackage `json:"packages"`

	Root              string              `json:"-"`
	PackageParsed     *ordered.OrderedMap `json:"-"`
	PackageLockParsed *ordered.OrderedMap `json:"-"`
}

// Persist writes a `.npm-mod.tidy.json` to disk.
func (tf *TidyFile) Persist() error {
	asJSON, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return err
	}
	asJSON = append(asJSON, '\n')

	target := filepath.Join(tf.Root, ".npm-mod.tidy.json")
	return os.WriteFile(target, asJSON, 0644)
}

// Restore writes back a `package.json` and `package-lock.json` based on the
// contents of a `.npm-mod.tidy.json` file.
func (tf *TidyFile) Restore() error {
	// NOTE: This should use the **existing** permissions of the `package.json`
	//       instead of just hardcoding `0644`.
	err := os.WriteFile(filepath.Join(tf.Root, "package.json"), tf.PackageJSON, 0644)
	if err != nil {
		return err
	}

	// NOTE: This should use the **existing** permissions of the `package-lock.json`
	//       instead of just hardcoding `0644`.
	return os.WriteFile(filepath.Join(tf.Root, "package-lock.json"), tf.PackageLockJSON, 0644)
}

// TidyPackageJSON updates (and writes) a `package.json` file with the
// vendored dependencies.
//
// This is a bit hacky. The algorithm is as follows:
// - Iterate over every package in `dependencies`, `devDependencies` and
//   `peerDependencies`
// - Find the package in `packages` in the `package-lock.json`, for example the
//   `node_modules/@testing-library/jest-dom` key corresponds to the
//   `@testing-library/jest-dom` dependency
// - Use the `resolved` URL for the `node_modules/...` match to determine the
//   local filename to use
func (tf *TidyFile) TidyPackageJSON() error {
	// Re-parse package JSON so we can modify it without mutating the value
	// stored on `tf`.
	pj := ordered.NewOrderedMap()
	err := json.Unmarshal(tf.PackageJSON, &pj)
	if err != nil {
		return err
	}

	byNodeModulesPath, _, err := PackageLockExtractDependencies(tf.PackageLockParsed)
	if err != nil {
		return err
	}

	pjr := PackageJSONReplace{ByNodeModulesPath: byNodeModulesPath}
	err = PackageJSONReplaceDependencies(pj, pjr.Replace)
	if err != nil {
		return err
	}

	asJSON, err := marshalWithoutHTMLEscape(pj)
	if err != nil {
		return err
	}

	filename := filepath.Join(tf.Root, "package.json")
	// NOTE: This should re-use the existing file permissions.
	return os.WriteFile(filename, asJSON, 0644)
}

// TidyPackageJSON updates (and writes) a `package-lock.json` file with the
// vendored dependencies.
func (tf *TidyFile) TidyPackageLockJSON() error {
	// Re-parse package lock so we can modify it without mutating the value
	// stored on `tf`.
	pl := ordered.NewOrderedMap()
	err := json.Unmarshal(tf.PackageLockJSON, &pl)
	if err != nil {
		return err
	}

	// Just re-compute `resolved` mapping by URL (it should also be stored in
	// `tf.Packages` but not as a map-by-URL).
	_, byURL, err := PackageLockExtractDependencies(pl)
	if err != nil {
		return err
	}

	plr := PackageLockReplace{ByURL: byURL}
	err = PackageLockReplaceDependencies(pl, plr.Replace)
	if err != nil {
		return err
	}

	asJSON, err := marshalWithoutHTMLEscape(pl)
	if err != nil {
		return err
	}

	filename := filepath.Join(tf.Root, "package-lock.json")
	// NOTE: This should re-use the existing file permissions.
	return os.WriteFile(filename, asJSON, 0644)
}

// GenerateTidyFile generates a `.npm-mod.tidy.json` by reading files from
// a `package.json` and `package-lock.json`.
func GenerateTidyFile(root string) (*TidyFile, error) {
	packageJSON, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return nil, err
	}
	packageLock, err := os.ReadFile(filepath.Join(root, "package-lock.json"))
	if err != nil {
		return nil, err
	}

	pl := ordered.NewOrderedMap()
	err = json.Unmarshal(packageLock, &pl)
	if err != nil {
		return nil, err
	}

	pj := ordered.NewOrderedMap()
	err = json.Unmarshal(packageJSON, &pj)
	if err != nil {
		return nil, err
	}

	_, byURL, err := PackageLockExtractDependencies(pl)
	if err != nil {
		return nil, err
	}

	tf := TidyFile{
		Version:         tidyFileVersion,
		PackageJSON:     packageJSON,
		PackageLockJSON: packageLock,
		Packages:        sortedPackages(byURL),

		Root:              root,
		PackageParsed:     pj,
		PackageLockParsed: pl,
	}
	return &tf, nil
}

// ReadTidyFile reads a `.npm-mod.tidy.json` file.
func ReadTidyFile(root string) (*TidyFile, error) {
	target := filepath.Join(root, ".npm-mod.tidy.json")
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, err
	}

	tf := TidyFile{
		Root:              root,
		PackageParsed:     ordered.NewOrderedMap(),
		PackageLockParsed: ordered.NewOrderedMap(),
	}
	err = json.Unmarshal(data, &tf)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tf.PackageJSON, &tf.PackageParsed)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tf.PackageLockJSON, &tf.PackageLockParsed)
	if err != nil {
		return nil, err
	}

	return &tf, nil
}

func resolvedKeys(byURL map[string]RegistryPackage) []string {
	keys := []string{}
	for k := range byURL {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedPackages(byURL map[string]RegistryPackage) []RegistryPackage {
	keys := resolvedKeys(byURL)
	resolved := make([]RegistryPackage, len(keys))
	for i, k := range keys {
		resolved[i] = byURL[k]
	}
	return resolved
}
