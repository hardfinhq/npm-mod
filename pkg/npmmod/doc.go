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

// Package npmmod contains the core implementation of code used to parse,
// interpret and modify `package.json` and `package-lock.json` files.
//
// In particular it seeks to replace versions in `package.json` from
//
// > {
// >   // ...
// >   "dependencies": {
// >     "react": "^18.0.0",
// >     // ...
// >   },
// >   "devDependencies": {
// >     "@types/react": "^18.0.0",
// >     // ...
// >   },
// >   "peerDependencies": {
// >     "@babel/core" "^7.0.0"
// >     // ...
// >   },
// >   // ...
// > }
//
// to
//
// > {
// >   // ...
// >   "dependencies": {
// >     "react": "file:vendor/react-18.0.0.tgz",
// >     // ...
// >   },
// >   "devDependencies": {
// >     "@types/react": "file:vendor/types__react-18.0.6.tgz",
// >     // ...
// >   },
// >   "peerDependencies": {
// >     "@babel/core" "file:vendor/babel__core-7.17.9.tgz"
// >     // ...
// >   },
// >   // ...
// > }
//
// And to replace all resolved packages in `package-lock.json` from
//
// > {
// >   // ...
// >   "packages": {
// >     // ...
// >     "node_modules/@babel/core": {
// >       "version": "7.17.9",
// >       "resolved": "https://registry.npmjs.org/@babel/core/-/core-7.17.9.tgz",
// >       // ...
// >     },
// >     // ...
// >     "node_modules/@types/react": {
// >       "version": "18.0.6",
// >       "resolved": "https://registry.npmjs.org/@types/react/-/react-18.0.6.tgz",
// >       // ...
// >     },
// >     // ...
// >     "node_modules/react": {
// >       "version": "18.0.0",
// >       "resolved": "https://registry.npmjs.org/react/-/react-18.0.0.tgz",
// >       // ...
// >     },
// >     // ...
// >   },
// >   "dependencies": {
// >     // ...
// >     "@babel/core": {
// >       "version": "7.17.9",
// >       "resolved": "https://registry.npmjs.org/@babel/core/-/core-7.17.9.tgz",
// >       // ...
// >       "dependencies": {
// >         "semver": {
// >           "version": "6.3.0",
// >           "resolved": "https://registry.npmjs.org/semver/-/semver-6.3.0.tgz",
// >           // ...
// >         }
// >       }
// >     },
// >     // ...
// >     "@types/react": {
// >       "version": "18.0.6",
// >       "resolved": "https://registry.npmjs.org/@types/react/-/react-18.0.6.tgz",
// >       // ...
// >     },
// >     // ...
// >     "react": {
// >       "version": "18.0.0",
// >       "resolved": "https://registry.npmjs.org/react/-/react-18.0.0.tgz",
// >       // ...
// >     },
// >     // ...
// >   }
// > }
//
// to
//
// > {
// >   // ...
// >   "packages": {
// >     // ...
// >     "node_modules/@babel/core": {
// >       "version": "file:vendor/babel__core-7.17.9.tgz",
// >       "resolved": "file:vendor/babel__core-7.17.9.tgz",
// >       // ...
// >     },
// >     // ...
// >     "node_modules/@types/react": {
// >       "version": "file:vendor/types__react-18.0.6.tgz",
// >       "resolved": "file:vendor/types__react-18.0.6.tgz",
// >       // ...
// >     },
// >     // ...
// >     "node_modules/react": {
// >       "version": "file:vendor/react-18.0.0.tgz",
// >       "resolved": "file:vendor/react-18.0.0.tgz",
// >       // ...
// >     },
// >     // ...
// >   },
// >   "dependencies": {
// >     // ...
// >     "@babel/core": {
// >       "version": "file:vendor/babel__core-7.17.9.tgz",
// >       "resolved": "file:vendor/babel__core-7.17.9.tgz",
// >       // ...
// >       "dependencies": {
// >         "semver": {
// >           "version": "file:vendor/semver-6.3.0.tgz",
// >           "resolved": "file:vendor/semver-6.3.0.tgz",
// >           // ...
// >         }
// >       }
// >     },
// >     // ...
// >     "@types/react": {
// >       "version": "file:vendor/types__react-18.0.6.tgz",
// >       "resolved": "file:vendor/types__react-18.0.6.tgz",
// >       // ...
// >     },
// >     // ...
// >     "react": {
// >       "version": "file:vendor/react-18.0.0.tgz",
// >       "resolved": "file:vendor/react-18.0.0.tgz",
// >       // ...
// >     },
// >     // ...
// >   }
// > }
package npmmod
