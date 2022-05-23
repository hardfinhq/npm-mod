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
	"strings"
	"testing"

	testifyassert "github.com/stretchr/testify/assert"

	"github.com/hardfinhq/npm-mod/pkg/npmmod"
	"github.com/hardfinhq/npm-mod/pkg/ordered"
)

// NOTE: Ensure that
//       * `replaceWithFile` satisfies `extract.ReplaceFunc`.
var (
	_ npmmod.ReplaceFunc = replaceWithFile
)

func TestPackageLockReplaceDependencies(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	b, err := os.ReadFile(filepath.Join("testdata", "package-lock.json"))
	assert.Nil(err)
	packageLock := ordered.NewOrderedMap()
	err = json.Unmarshal(b, &packageLock)
	assert.Nil(err)

	err = npmmod.PackageLockReplaceDependencies(packageLock, replaceWithFile)
	assert.Nil(err)

	asJSON, err := marshalWithoutHTMLEscape(packageLock)
	assert.Nil(err)
	expected, err := os.ReadFile(filepath.Join("testdata", "golden.package-lock.json"))
	assert.Nil(err)
	assert.True(bytes.Equal(expected, asJSON), "golden.package-lock.json")
}

func TestPackageLockExtractDependencies(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	b, err := os.ReadFile(filepath.Join("testdata", "package-lock.json"))
	assert.Nil(err)
	packageLock := ordered.NewOrderedMap()
	err = json.Unmarshal(b, &packageLock)
	assert.Nil(err)

	byNodeModulesPath, byURL, err := npmmod.PackageLockExtractDependencies(packageLock)
	assert.Nil(err)
	toSerialize := map[string]any{
		"node_modules": byNodeModulesPath,
		"url":          byURL,
	}
	asJSON, err := json.MarshalIndent(toSerialize, "", "  ")
	assert.Nil(err)
	asJSON = append(asJSON, '\n')
	expected, err := os.ReadFile(filepath.Join("testdata", "golden.extracted.json"))
	assert.Nil(err)
	assert.True(bytes.Equal(expected, asJSON), "golden.extracted.json")
}

func replaceWithFile(resolved string) string {
	parts := strings.Split(resolved, "/")
	return "file:" + parts[len(parts)-1]
}

func marshalWithoutHTMLEscape(m *ordered.OrderedMap) ([]byte, error) {
	var b bytes.Buffer
	je := json.NewEncoder(&b)
	je.SetEscapeHTML(false)
	je.SetIndent("", "  ")
	err := je.Encode(m)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
