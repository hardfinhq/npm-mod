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
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	testifyassert "github.com/stretchr/testify/assert"

	"github.com/hardfinhq/npm-mod/pkg/npmmod"
)

func TestFetch(t *testing.T) {
	t.Parallel()
	assert := testifyassert.New(t)

	ctx := context.TODO()

	data, err := os.ReadFile(filepath.Join("testdata", "builtins-1.0.3.tgz"))
	assert.Nil(err)
	destination, err := os.MkdirTemp("", "")
	assert.Nil(err)
	t.Cleanup(func() {
		err = os.RemoveAll(destination)
		assert.Nil(err)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/builtins/-/builtins-1.0.3.tgz" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	t.Cleanup(func() {
		server.Close()
	})

	filename := filepath.Join(destination, "builtins-1.0.3.tgz")
	url := server.URL + "/builtins/-/builtins-1.0.3.tgz"
	algorithm := "sha1"
	hash := "y5T662HIaWRR2zZTThQi+U8K7og="
	err = npmmod.Fetch(ctx, url, algorithm, hash, filename)
	assert.Nil(err)

	expected, err := os.ReadFile(filename)
	assert.Nil(err)
	assert.True(bytes.Equal(expected, data))
}
