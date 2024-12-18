# Dependency Graph

Sometimes it is necessary to investigate dependency graph and find out dependency paths to certain package.
This is especially helpful when resolving security CVE issues. This document documents a way to find necessary
information easily.

## Deptree

Install [deptree](https://github.com/vc60er/deptree) in the GOBIN dir on the path.

Since it takes a lot of time, it will save time to record module list once, so deptree can be run quickly
passing it as an argument.

```shell
go list -u -m -json all > upgradefile.txt
```

Now you can run detptree to render the dependency tree.

```shell
go mod graph | deptree -upgrade=upgradefile.txt > tree.txt
```
