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

package npmmod_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	testifyassert "github.com/stretchr/testify/assert"

	"github.com/hardfinhq/npm-mod/pkg/npmmod"
)

func TestLocate(outer *testing.T) {
	outer.Parallel()
	assertOuter := testifyassert.New(outer)

	here, err := os.Getwd()
	assertOuter.Nil(err)

	type testCase struct {
		Path    string
		Located string
		Error   string
	}

	cases := []testCase{
		{Path: "testdata", Located: filepath.Join(here, "testdata")},
		{
			Path:  filepath.Join("testdata", "a"),
			Error: fmt.Sprintf("package.json exists but package-lock.json does not; %s", filepath.Join(here, "testdata", "a")),
		},
		{Path: filepath.Join("testdata", "a", "b"), Located: filepath.Join(here, "testdata", "a", "b")},
		{Path: filepath.Join("testdata", "a", "b", "c"), Located: filepath.Join(here, "testdata", "a", "b")},
	}
	for _, tc := range cases {
		tc := tc // Copy to local to avoid closure around pointer
		outer.Run(tc.Path, func(t *testing.T) {
			t.Parallel()
			assert := testifyassert.New(t)

			located, err := npmmod.Locate(tc.Path)
			assert.Equal(tc.Located, located)
			if tc.Error == "" {
				assert.Nil(err)
			} else {
				assert.NotNil(err)
				assert.Equal(tc.Error, fmt.Sprintf("%v", err))
			}
		})
	}
}
