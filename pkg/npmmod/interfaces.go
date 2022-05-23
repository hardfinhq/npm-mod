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
	"github.com/hardfinhq/npm-mod/pkg/ordered"
)

// VisitorFunc is a function for visiting a value in an ordered map. In
// addition to taking the key / value pair as input, it also returns the
// parent map so it can be modified if needed.
type VisitorFunc func(m *ordered.OrderedMap, k string, v any) error

// ReplacePairFunc replaces a value based on the existing key/value pair.
type ReplacePairFunc func(key, value string) string

// ReplaceFunc replaces a value based on the value.
type ReplaceFunc func(value string) string
