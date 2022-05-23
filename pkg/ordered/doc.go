// Package ordered is a vendored in version of
// `gitlab.com/c0b/go-ordered-json@febf46534d5a4f9f64ea4c1874f0e55791373a2b`
//
// The package is vendored in because usages of `json.Marshal()` for
// **values** (as opposed to keys) in the ordered map are not configurable.
// Here in particular we have semver ranges e.g. `>= 2.3.1` that by default
// get HTML escaped as `\u003e= 2.3.1`.
package ordered
