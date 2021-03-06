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

package tidycmd

import (
	"context"
	"os"

	"github.com/hardfinhq/npm-mod/pkg/npmmod"
)

// Run executes the `npm-mod tidy` command.
func Run(_ context.Context) error {
	here, err := os.Getwd()
	if err != nil {
		return err
	}

	root, err := npmmod.Locate(here)
	if err != nil {
		return err
	}

	tf, err := npmmod.GenerateTidyFile(root)
	if err != nil {
		return err
	}

	err = tf.Persist()
	if err != nil {
		return err
	}

	err = tf.TidyPackageJSON()
	if err != nil {
		return err
	}

	return tf.TidyPackageLockJSON()
}
