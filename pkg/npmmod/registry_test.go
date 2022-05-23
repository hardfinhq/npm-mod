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
	"testing"

	"github.com/hardfinhq/npm-mod/pkg/npmmod"
	testifyassert "github.com/stretchr/testify/assert"
)

func TestFilenameFromURL(outer *testing.T) {
	outer.Parallel()

	type testCase struct {
		URL      string
		Filename string
		Error    string
	}

	cases := []testCase{
		{URL: "https://registry.npmjs.org/builtins/-/builtins-1.0.3.tgz", Filename: "builtins-1.0.3.tgz"},
		{URL: "https://registry.npmjs.org/@babel/cli/-/cli-7.15.7.tgz", Filename: "babel__cli-7.15.7.tgz"},
		{URL: "https://registry.npmjs.org/missing.tgz", Error: "npm url in unexpected format; url: https://registry.npmjs.org/missing.tgz"},
		{URL: "https://registry.npmjs.org/babel/cli/-/cli-7.15.7.tgz", Error: "package scope is missing @ prefix; babel/cli"},
		{URL: "https://registry.npmjs.org/babel/c/l/i/-/cli-7.15.7.tgz", Error: "package name has more than two components; babel/c/l/i"},
		{URL: " https://web.invalid", Error: `parse " https://web.invalid": first path segment in URL cannot contain colon`},
	}
	for _, tc := range cases {
		tc := tc // Copy to local to avoid closure around pointer
		outer.Run(tc.URL, func(t *testing.T) {
			t.Parallel()
			assert := testifyassert.New(t)

			filename, err := npmmod.FilenameFromURL(tc.URL)
			assert.Equal(tc.Filename, filename, tc.URL)
			if tc.Error == "" {
				assert.Nil(err, tc.URL)
			} else {
				assert.NotNil(err, tc.URL)
				assert.Equal(tc.Error, fmt.Sprintf("%v", err), tc.URL)
			}
		})
	}
}
