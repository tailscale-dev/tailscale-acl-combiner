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

> **Note**: the arguments for parent file, directory of child files, and acl sections to allow are all required. This is to prevent accidental omission resulting in an unexpected final file.

### Example

Using the `testdata` directory in this repo:

```shell
$ tailscale-acl-combiner \
  -f testdata/input-parent.hujson \
  -d testdata/departments \
  -allow acls,extraDNSRecords,grants,groups,ssh,tests

{
  "acls": [
    // acls from `testdata/input-parent.hujson` and any files found under `testdata/departments`
  ],
  "groups": [
    // groups from `testdata/input-parent.hujson` and any files found under `testdata/departments`
  ],
  // ...
}
```

### Usage in a GitOps workflow

We recommend the following steps when using `tailscale-acl-combiner` in a GitOps workflow:

1. Create a new branch in the repo containing your policy files.
1. Make your change locally.
1. Use `tailscale-acl-combiner` to generate an updated file and commit the combined file to your branch.
1. Open a pull or merge request with your updates and ask a peer to review your changes.
1. In your GitOps workflow, run `tailscale-acl-combiner` to generate a new, temporary file and compare to the committed file - e.g. `diff -c policy.hujson $RUNNER_TEMP/generated-by-pull-request.hujson`.
    1. If differences **are** found, cancel the workflow and require updates.
    1. If differences are **not** found, allow the workflow to proceed.
1. Once the pull request is merged, have the GitOps workflow repeat the generate and compare steps then test and apply the ACL to your Tailnet.

By committing the file you have a versioned artifact to review in the future and revert to if necessary.

See [.github/workflows/combine-and-push-acls.yaml.example](.github/workflows/combine-and-push-acls.yaml.example) for an example.

## Limitations

- Top-level objects and arrays are appended, not merged.
  - For example, if one child file has `"groups": { "group1": ["user1"] })` and another child has `"groups": { "group1": ["user2"] })`, the resulting file will have two `group1` groups with different members.
  - *See the next limitation about Duplicate names.*
- Duplicate names (e.g. `"groups": { "group1": [], "group1": [] })`) will not result in an error from `tailscale-acl-combiner`.
  - Go's "encoding/json" does not enforce this, see [https://golang.org/issue/48298](https://golang.org/issue/48298).
- `autoApprovers`, `derpMap`, `disableIPv4`, `OneCGNATRoute`, `randomizeClientPort`, and other [network-wide policy settings](https://tailscale.com/kb/1337/acl-syntax#network-policy-options) are only allowed in the provided parent file.
