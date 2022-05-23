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
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Fetch downloads a package from `npm`, validates the checksum and then
// writes it to disk.
func Fetch(ctx context.Context, url, algorithm, hash, filename string) error {
	data, err := download(ctx, url)
	if err != nil {
		return err
	}

	err = ValidateIntegrity(data, algorithm, hash)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	return err
}

func download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request failed; response code: %d", resp.StatusCode)
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
