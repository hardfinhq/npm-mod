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

package vendorcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hardfinhq/npm-mod/pkg/concurrency"
	"github.com/hardfinhq/npm-mod/pkg/npmmod"
)

var (
	poolSize = runtime.NumCPU()
)

// fetchPackageArchives runs `fetchPackageArchive.Do()` for every (deduplicated)
// registry package.
func fetchPackageArchives(ctx context.Context, tf *npmmod.TidyFile) error {
	// Ensure vendor directory exists.
	targetDir := filepath.Join(tf.Root, "vendor")
	err := os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Fan out check file / download tasks to a worker pool.
	tasks := make([]*concurrency.Task, len(tf.Packages))
	for i, rp := range tf.Packages {
		fpa := fetchPackageArchive{
			Context:         ctx,
			RegistryPackage: rp,
			Target:          targetDir,
		}
		tasks[i] = concurrency.NewTask(fpa.Do)
	}

	pool := concurrency.NewPool(tasks, poolSize)
	return pool.Run()
}

type fetchPackageArchive struct {
	Context         context.Context
	RegistryPackage npmmod.RegistryPackage
	Target          string
}

// Do either
// - validates the checksum if the package archive file already exists
// - downloads (and validates) the package archive file
func (fpa *fetchPackageArchive) Do(_ int) error {
	filename, err := fpa.RegistryPackage.Filename()
	if err != nil {
		return err
	}

	archiveFilename := filepath.Join(fpa.Target, filename)
	data, err := os.ReadFile(archiveFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return fpa.loggedFetch()
		}

		return err
	}

	err = npmmod.ValidateIntegrity(data, fpa.RegistryPackage.Algorithm, fpa.RegistryPackage.Hash)
	if err != nil {
		return err
	}

	fmt.Printf("Validated %s\n", filename)
	return nil
}

func (fpa *fetchPackageArchive) loggedFetch() error {
	rp := fpa.RegistryPackage
	filename, err := rp.Filename()
	if err != nil {
		return err
	}

	downloadFilename := filepath.Join(fpa.Target, filename)
	err = npmmod.Fetch(fpa.Context, rp.URL, rp.Algorithm, rp.Hash, downloadFilename)
	if err != nil {
		return err
	}

	fmt.Printf("Saved %s\n", filename)
	return nil
}
