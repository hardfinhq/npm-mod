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

package concurrency

import (
	multierror "github.com/hashicorp/go-multierror"
)

// maybeMultiError combines a slice of errors (each of which may be `nil`).
//
// This **only** uses a `multierror` if two or more errors are not `nil`.
// Unfortunately neither `multierror.Error.ErrorOrNil()` nor
// `multierror.Error.Unwrap()` support this behavior, but both come close.
func maybeMultiError(errors ...error) error {
	err := multierror.Append(nil, errors...)
	if len(err.Errors) == 0 {
		return nil
	}
	if len(err.Errors) == 1 {
		return err.Errors[0]
	}
	return err
}
