# Third-party notices

The following code dependencies are bundled into the Cress executable.
Their tagged license texts were reviewed against the approved code-license
policy in [ADR 0011](adr/0011-bsd-3-clause-license.md).

| Component | Version | License | Source |
|-----------|---------|---------|--------|
| urfave/cli | v3.14.0 | MIT | <https://github.com/urfave/cli/tree/v3.14.0> |
| go-toml v2 | v2.4.3 | MIT | <https://github.com/pelletier/go-toml/tree/v2.4.3> |
| goldmark | v1.8.6 | MIT | <https://github.com/yuin/goldmark/tree/v1.8.6> |
| Chroma | v2.27.0 | MIT | <https://github.com/alecthomas/chroma/tree/v2.27.0> |
| regexp2 | v2.2.1 | MIT | <https://github.com/dlclark/regexp2/tree/v2.2.1> |
| Go text | v0.43.0 | BSD-3-Clause | <https://github.com/golang/text/tree/v0.43.0> |
| Go modules (semver) | v0.42.0 | BSD-3-Clause | <https://github.com/golang/mod/tree/v0.42.0> |

These licenses permit redistribution with their copyright and permission
notices. No third-party theme assets are bundled. Release-specific SBOMs
and automated component-review enforcement remain phase 9 work.

Chroma's HTML formatter, lexers, and styles are linked. Its SVG formatter
and bundled Liberation Mono font are not linked or redistributed; the font
license in Chroma's `COPYING` does not apply to the bundled components.

The schema tests also use
[jsonschema v6.0.3](https://github.com/santhosh-tekuri/jsonschema/tree/v6.0.3),
reviewed under its
[Apache-2.0 license](https://github.com/santhosh-tekuri/jsonschema/blob/v6.0.3/LICENSE).
That validator is a test-only dependency, not part of the Cress executable.
Its Go text dependency is covered above.

## MIT license

The following copyright notices apply to their respective components:

- urfave/cli: Copyright (c) 2023 urfave/cli maintainers
- go-toml v2: Copyright (c) 2021 - 2023 Thomas Pelletier
- goldmark: Copyright (c) 2019 Yusuke Inuzuka
- Chroma: Copyright (C) 2017 Alec Thomas
- regexp2: Copyright (c) Doug Clark

The following license text applies to each component listed above:

```text
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Go modules and text BSD-3-Clause license

```text
Copyright 2009 The Go Authors.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google LLC nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

## Go modules and text additional IP rights grant

```text
Additional IP Rights Grant (Patents)

"This implementation" means the copyrightable works distributed by
Google as part of the Go project.

Google hereby grants to You a perpetual, worldwide, non-exclusive,
no-charge, royalty-free, irrevocable (except as stated in this section)
patent license to make, have made, use, offer to sell, sell, import,
transfer and otherwise run, modify and propagate the contents of this
implementation of Go, where such license applies only to those patent
claims, both currently owned or controlled by Google and acquired in
the future, licensable by Google that are necessarily infringed by this
implementation of Go.  This grant does not include claims that would be
infringed only as a consequence of further modification of this
implementation.  If you or your agent or exclusive licensee institute or
order or agree to the institution of patent litigation against any
entity (including a cross-claim or counterclaim in a lawsuit) alleging
that this implementation of Go or any code incorporated within this
implementation of Go constitutes direct or contributory patent
infringement, or inducement of patent infringement, then any patent
rights granted to you under this License for this implementation of Go
shall terminate as of the date such litigation is filed.
```
