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
	neturl "net/url"
	"strings"
)

// RegistryPackage represents a package in an `npm` package registry.
type RegistryPackage struct {
	URL       string `json:"url"`
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
}

// Equal compares two registry packages for equality.
func (rp RegistryPackage) Equal(other RegistryPackage) bool {
	return rp.URL == other.URL && rp.Algorithm == other.Algorithm && rp.Hash == other.Hash
}

// Filename creates a normalized filename from the `npm` registry URL.
func (rp RegistryPackage) Filename() (string, error) {
	return FilenameFromURL(rp.URL)
}

// FilenameFromURL creates a normalized filename from an `npm` registry URL.
func FilenameFromURL(url string) (string, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return "", err
	}
	// Normalize path
	path := strings.TrimPrefix(u.Path, "/")

	// Paths are of the form `/{PACKAGE_NAME}/-/{FILENAME}.tgz`.
	parts := strings.Split(path, "/-/")
	if len(parts) != 2 {
		err = fmt.Errorf("npm url in unexpected format; url: %s", url)
		return "", err
	}

	scope, err := getPackageScope(parts[0])
	if err != nil {
		return "", err
	}

	if scope == "" {
		return parts[1], nil
	}

	// NOTE: We could go a step further and validate that the filename matches
	//       the package name.
	return scope + "__" + parts[1], nil
}

func getPackageScope(packageName string) (string, error) {
	parts := strings.Split(packageName, "/")
	if len(parts) > 2 {
		return "", fmt.Errorf("package name has more than two components; %s", packageName)
	}

	if len(parts) == 1 {
		return "", nil
	}

	scope := parts[0]
	without := strings.TrimPrefix(scope, "@")
	if without == scope {
		return "", fmt.Errorf("package scope is missing @ prefix; %s", packageName)
	}

	return without, nil
}
