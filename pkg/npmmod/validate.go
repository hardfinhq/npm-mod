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
	"crypto/sha1"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// ValidateIntegrity checks the hash of a downloaded package.
func ValidateIntegrity(data []byte, algorithm, hash string) error {
	if algorithm == "sha1" {
		return validateIntegritySHA1(data, hash)
	}
	if algorithm == "sha512" {
		return validateIntegritySHA512(data, hash)
	}
	return fmt.Errorf("unexpected integrity format; %s", algorithm)
}

func validateIntegritySHA1(data []byte, hashBase64 string) error {
	expected, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return err
	}

	actual := sha1.Sum(data)
	if subtle.ConstantTimeCompare(expected, actual[:]) == 1 {
		return nil
	}

	actualBase64 := base64.StdEncoding.EncodeToString(actual[:])
	return fmt.Errorf("sha1 hashes do not match; expected: %s; actual: %s", hashBase64, actualBase64)
}

func validateIntegritySHA512(data []byte, hashBase64 string) error {
	expected, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return err
	}

	actual := sha512.Sum512(data)
	if subtle.ConstantTimeCompare(expected, actual[:]) == 1 {
		return nil
	}

	actualBase64 := base64.StdEncoding.EncodeToString(actual[:])
	return fmt.Errorf("sha512 hashes do not match; expected: %s; actual: %s", hashBase64, actualBase64)
}
