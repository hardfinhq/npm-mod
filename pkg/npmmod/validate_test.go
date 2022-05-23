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
	"os"
	"path/filepath"
	"testing"

	testifyassert "github.com/stretchr/testify/assert"

	"github.com/hardfinhq/npm-mod/pkg/npmmod"
)

func TestValidateIntegrity_SHA1(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	data, err := os.ReadFile(filepath.Join("testdata", "builtins-1.0.3.tgz"))
	assert.Nil(err)
	// "https://registry.npmjs.org/builtins/-/builtins-1.0.3.tgz",
	err = npmmod.ValidateIntegrity(data, "sha1", "y5T662HIaWRR2zZTThQi+U8K7og=")
	assert.Nil(err)
}

func TestValidateIntegrity_SHA512(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	data, err := os.ReadFile(filepath.Join("testdata", "shebang-regex-3.0.0.tgz"))
	assert.Nil(err)
	// "https://registry.npmjs.org/shebang-regex/-/shebang-regex-3.0.0.tgz",
	err = npmmod.ValidateIntegrity(data, "sha512", "7++dFhtcx3353uBaq8DDR4NuxBetBzC7ZQOhmTQInHEd6bSrXdiEyzCvG07Z44UYdLShWUyXt5M/yhz8ekcb1A==")
	assert.Nil(err)
}
