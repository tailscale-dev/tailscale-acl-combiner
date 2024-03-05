# tailscale-acl-combiner

[![status: experimental](https://img.shields.io/badge/status-experimental-blue)](https://tailscale.com/kb/1167/release-stages/#experimental)

A CLI tool to facilitate delegation of managing Tailscale ACLs across multiple people and teams within an organization. Provide a parent ACL file and a directory for "child" ACL files and all are merged into a single ACL file.

## Usage

```shell
$ go run main.go -f <parent-file> -d <directory-of-child-files> -allow <acl-sections-to-allow>
...
```

For example, using the `testdata` directory in this repo:

```shell
$ tailscale-acl-combiner -f testdata/input-parent.hujson -d testdata/departments -allow acls,extraDNSRecords,grants,groups,ssh,tests
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
