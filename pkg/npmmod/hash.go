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
	"fmt"
	"strings"
)

var (
	acceptedAlgorithms = map[string]bool{
		"sha1":   true,
		"sha512": true,
	}
)

func splitIntegrity(integrity string) (string, string, error) {
	parts := strings.SplitN(integrity, "-", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("unexpected integrity format; %s", integrity)
		return "", "", err
	}

	algorithm := parts[0]
	hash := parts[1]
	_, ok := acceptedAlgorithms[algorithm]
	if !ok {
		err := fmt.Errorf("unknown integrity algorithm; %s", integrity)
		return "", "", err
	}
	return algorithm, hash, nil
}
