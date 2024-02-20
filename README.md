# tailscale-acl-combiner

[![status: experimental](https://img.shields.io/badge/status-experimental-blue)](https://tailscale.com/kb/1167/release-stages/#experimental)

A CLI tool to facilitate delegation of managing Tailscale ACLs across multiple people and teams within an organization. Provide a parent ACL file and a directory for "child" ACL files and all are merged into a single ACL file.

## Usage

```shell
$ go run main.go -f <parent-file> -d <directory-of-child-files>
...
```

For example, using the `testdata` directory in this repo:

```shell
$ go run main.go -f testdata/parent.hujson -d testdata/departments
{
  "acls": [
    // ...
  ],
  // ...
}
```
