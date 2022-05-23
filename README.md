# `npm-mod`

> An experiment in bringing the `go mod vendor` experience to the `npm`
> ecosystem. See the [blog post][2].

<p align="center">
  <img alt="npm and Go" src="./_images/npm-and-go.png?raw=true" />
</p>

## Installation

```
$ go install github.com/hardfinhq/npm-mod/cmd/npm-mod@v1.20220523.1
```

## `npm-mod tidy` Subcommand

For the purposes of demonstration, we'll be running subcommands in an
application (called `sample`) created with the `create-react-app` [tool][1].
Running the `tidy` subcommand we see a new file and two changed files:

```bash
$ npm-mod tidy
$ git status
On branch main
Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
        modified:   package-lock.json
        modified:   package.json

Untracked files:
  (use "git add <file>..." to include in what will be committed)
        .npm-mod.tidy.json

no changes added to commit (use "git add" and/or "git commit -a")
```

Digging in a bit to see what's changed. The `.npm-mod.tidy.json` captures
the contents of the modified files so we can revert easily if need be and
also tracks the URLs and file integrity for all packages referenced:

```bash
$ cat .npm-mod.tidy.json
{
  "version": "22.05",
  "package.json": "ewogICJuYW1lIjogInNhbXBsZSIsCiAg...",
  "package-lock.json": "ewogICJuYW1lIjogInNhbXBsZSIsCiAg...",
  "packages": [
    {
      "url": "https://registry.npmjs.org/@ampproject/remapping/-/remapping-2.1.2.tgz",
      "algorithm": "sha512",
      "hash": "hoyByceqwKirw7w3Z7gnIIZC3Wx3J484Y3L/cMpXFbr7d9ZQj2mODrirNzcJa+SM3UlpWXYvKV4RlRpFXlWgXg=="
    },
    {
      "url": "https://registry.npmjs.org/@apideck/better-ajv-errors/-/better-ajv-errors-0.3.3.tgz",
      "algorithm": "sha512",
      "hash": "9o+HO2MbJhJHjDYZaDxJmSDckvDpiuItEsrIShV0DXeCshXWRHhqYyU/PKHMkuClOmFnZhRd6wzv4vpDu/dRKg=="
    },
...
```

The `package.json` has had every semver version range swapped for an explicit
**local** file reference:

```diff
diff --git a/package.json b/package.json
index 82115d2..753457c 100644
--- a/package.json
+++ b/package.json
@@ -3,13 +3,13 @@
   "version": "0.0.1",
   "private": true,
   "dependencies": {
-    "@testing-library/jest-dom": "^5.16.4",
-    "@testing-library/react": "^13.1.1",
-    "@testing-library/user-event": "^13.5.0",
-    "react": "^18.0.0",
-    "react-dom": "^18.0.0",
-    "react-scripts": "5.0.1",
-    "web-vitals": "^2.1.4"
+    "@testing-library/jest-dom": "file:vendor/testing-library__jest-dom-5.16.4.tgz",
+    "@testing-library/react": "file:vendor/testing-library__react-13.1.1.tgz",
+    "@testing-library/user-event": "file:vendor/testing-library__user-event-13.5.0.tgz",
+    "react": "file:vendor/react-18.0.0.tgz",
+    "react-dom": "file:vendor/react-dom-18.0.0.tgz",
+    "react-scripts": "file:vendor/react-scripts-5.0.1.tgz",
+    "web-vitals": "file:vendor/web-vitals-2.1.4.tgz"
   },
   "scripts": {
     "start": "react-scripts start",
```

and the `package-lock.json` has had a similar transformation, this time
swapping URLs for local file references:

```diff
diff --git a/package-lock.json b/package-lock.json
index 5303979..02c30e9 100644
--- a/package-lock.json
+++ b/package-lock.json
@@ -18,8 +18,8 @@
       }
     },
     "node_modules/@ampproject/remapping": {
-      "version": "2.1.2",
-      "resolved": "https://registry.npmjs.org/@ampproject/remapping/-/remapping-2.1.2.tgz",
+      "version": "file:vendor/ampproject__remapping-2.1.2.tgz",
+      "resolved": "file:vendor/ampproject__remapping-2.1.2.tgz",
       "integrity": "sha512-hoyByceqwKirw7w3Z7gnIIZC3Wx3J484Y3L/cMpXFbr7d9ZQj2mODrirNzcJa+SM3UlpWXYvKV4RlRpFXlWgXg==",
       "dependencies": {
         "@jridgewell/trace-mapping": "^0.3.0"
@@ -29,8 +29,8 @@
...
```

## `npm-mod vendor` Subcommand

Just checking in the changes from `npm-mod tidy` is insufficient; the
application will be broken because all of the `file:vendor/...` references
don't point anywhere. To fix this, the `npm-mod vendor` subcommand will
download all packages referenced in `.npm-mod.tidy.json`:

```bash
$ npm-mod vendor
Saved babel__helper-builder-binary-assignment-operator-visitor-7.16.7.tgz
Saved babel__generator-7.17.9.tgz
...
Saved yargs-16.2.0.tgz
Saved yocto-queue-0.1.0.tgz
$
$
$ git status
On branch main
Untracked files:
  (use "git add <file>..." to include in what will be committed)
        vendor/

nothing added to commit but untracked files present (use "git add" to track)
$
$
$ ls -1 vendor/
abab-2.0.6.tgz
accepts-1.3.8.tgz
...
yargs-parser-20.2.9.tgz
yocto-queue-0.1.0.tgz
```

## `npm-mod unvendor` Subcommand

Using the `package.json` and `package-lock.json` fields in `.npm-mod.tidy.json`
the `unvendor` can restore the original semver ranges. This would enable a
workflow using the `npm-mod` that can switch from "native `npm`" mode back to
"vendor-style packages checked into source tree". With this workflow, updating
a dependency could be done as follows:

- Run `npm-mod unvendor` subcommand to "go back" to standard form for
  `package.json`
- Modify `package.json` semver ranges to prepare for any new packages or
  package version updates needed
- Delete any or all of `vendor/`, `node_modules/`, `package-lock.json` and
  `.npm-mod.tidy.json` to ensure all desired package updates happen
- Run `npm-mod tidy` again to switch back to `file:vendor/...` references
- Run `npm-mod unvendor` again to download newly added or changed packages

## Caveats

The primary goal of this project is to enable a thought experiment in `npm`
packaging. As a result, the CLI may have incomplete support for all of the
various shapes a `package-lock.json` file can exhibit.

Some things we explicitly don't support, but may choose to expand support
for over time:

- Providing a `yarn-vendor` equivalent (or `pnpm-vendor` for that matter)
- Support `lockfileVersion=1` for the `npm` package lock format (only
  `lockfileVersion=2` is supported)

[1]: https://reactjs.org/docs/create-a-new-react-app.html
[2]: https://engineering.hardfin.com/2022/05/npm-mod/
