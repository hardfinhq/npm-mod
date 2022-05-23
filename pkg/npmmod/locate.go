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
	"os"
	"path/filepath"
)

// Locate determines the location of the `package.json` file. It searches
// the current directory and then all parents until the file is found. This
// errors if the file cannot be found, if the file cannot be accessed by the
// current user or if the package lock cannot be found.
func Locate(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return locateAbs(dir)
}

func locateAbs(dir string) (string, error) {
	exists, err := packageAndLockExist(dir)
	if err != nil {
		return "", err
	}
	if exists {
		return dir, nil
	}

	parent := filepath.Dir(dir)
	if parent == dir {
		return "", fmt.Errorf("path has no parent; %s", dir)
	}

	return locateAbs(parent)
}

func packageAndLockExist(dir string) (bool, error) {
	p := filepath.Join(dir, "package.json")
	exists, err := fileExists(p)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	pl := filepath.Join(dir, "package-lock.json")
	exists, err = fileExists(pl)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("package.json exists but package-lock.json does not; %s", dir)
	}

	return true, nil
}

func fileExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if fi.Mode().IsDir() {
		return false, fmt.Errorf("path exists but is a directory; %s", path)
	}
	return true, nil
}
