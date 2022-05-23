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
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	testifyassert "github.com/stretchr/testify/assert"

	"github.com/hardfinhq/npm-mod/pkg/npmmod"
	"github.com/hardfinhq/npm-mod/pkg/ordered"
)

// NOTE: Ensure that
//       * `replaceWithCaret` satisfies `extract.ReplacePairFunc`.
var (
	_ npmmod.ReplacePairFunc = replaceWithCaret
)

func TestPackageJSONReplaceDependencies(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	b, err := os.ReadFile(filepath.Join("testdata", "package.json"))
	assert.Nil(err)
	packageJSON := ordered.NewOrderedMap()
	err = json.Unmarshal(b, &packageJSON)
	assert.Nil(err)

	err = npmmod.PackageJSONReplaceDependencies(packageJSON, replaceWithCaret)
	assert.Nil(err)

	asJSON, err := json.MarshalIndent(packageJSON, "", "  ")
	assert.Nil(err)
	asJSON = append(asJSON, '\n')
	expected, err := os.ReadFile(filepath.Join("testdata", "golden.package.json"))
	assert.Nil(err)
	assert.True(bytes.Equal(expected, asJSON), "golden.package.json")
}

func replaceWithCaret(_, packageVersion string) string {
	return "^" + packageVersion
}
