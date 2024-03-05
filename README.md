# tailscale-acl-combiner

[![status: experimental](https://img.shields.io/badge/status-experimental-blue)](https://tailscale.com/kb/1167/release-stages/#experimental)

A CLI tool to facilitate delegation of managing Tailscale ACLs across multiple people and teams within an organization. Provide a parent ACL file and a directory for "child" ACL files and all are merged into a single ACL file.

## Installation

```shell
go install github.com/clstokes/tailscale-acl-combiner@latest
```

## Usage

```shell
tailscale-acl-combiner -f <parent-file> -d <directory-of-child-files> -allow <acl-sections-to-allow>
...
```

> **Note**: the arguments for parent file, directory of child files, and acl sections to allow are all required. This is to prevent accidental omission resulting in a syntactically correct but unexpected final file.

### Example

Using the `testdata` directory in this repo:

```shell
tailscale-acl-combiner -f testdata/input-parent.hujson -d testdata/departments -allow acls,extraDNSRecords,grants,groups,ssh,tests
{
  "acls": [
    // acls from `testdata/input-parent.hujson` and any files found under `testdata/departments`
  ],
  "groups": [
    // groups from `testdata/input-parent.hujson` and any files found under `testdata/departments`
  ],
  // etc
}
```

## Limitations

- Top-level objects and arrays are appended, not merged.
  - For example, if one child file has `"groups": { "group1": ["user1"] })` and another child has `"groups": { "group1": ["user2"] })`, the resulting file will have two `group1` groups with different members.
  - *See the next limitation about Duplicate names.*
- Duplicate names (e.g. `"groups": { "group1": [], "group1": [] })`) will not result in an error.
  - Go's "encoding/json" does not enforce this, see [https://golang.org/issue/48298](https://golang.org/issue/48298).
- `autoApprovers`, `derpMap`, `disableIPv4`, `OneCGNATRoute`, `randomizeClientPort`, and other [network-wide policy settings](https://tailscale.com/kb/1337/acl-syntax#network-policy-options) are only allowed in the provided parent file.
