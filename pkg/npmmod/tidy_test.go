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
)

func TestGenerateTidyFile(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	tf, err := npmmod.GenerateTidyFile("testdata")
	assert.Nil(err)
	asJSON, err := json.MarshalIndent(tf, "", "  ")
	assert.Nil(err)
	asJSON = append(asJSON, '\n')
	expected, err := os.ReadFile(filepath.Join("testdata", "golden.npm-mod.tidy.json"))
	assert.Nil(err)
	assert.True(bytes.Equal(expected, asJSON), "golden.npm-mod.tidy.json")
}

func TestTidyFile_Persist(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	destination, err := os.MkdirTemp("", "")
	assert.Nil(err)
	t.Cleanup(func() {
		err = os.RemoveAll(destination)
		assert.Nil(err)
	})

	tf := npmmod.TidyFile{
		Root:            destination,
		Version:         "fake",
		PackageJSON:     []byte("a"),
		PackageLockJSON: []byte("b"),
	}
	err = tf.Persist()
	assert.Nil(err)

	actual, err := os.ReadFile(filepath.Join(destination, ".npm-mod.tidy.json"))
	assert.Nil(err)
	expected := []byte(`{
  "version": "fake",
  "package.json": "YQ==",
  "package-lock.json": "Yg==",
  "packages": null
}
`)
	assert.True(bytes.Equal(expected, actual), ".npm-mod.tidy.json")
}
