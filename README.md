# jf

`jf` takes a json object and filters it with a GraphQL query.

This is useful for comparing subsets of json objects, and for canonicalizing and hashing json.

## As a CLI tool

### Install

```bash
go install github.com/ecordell/jf
```

### Usage

```bash
$ jf --help                                                                             127 ↵
jf is a tool to pull out a subset of a json document in a standard, hashable way.

Usage:
  jf [flags] [file] [-]

Flags:
  -f, --file string    file containing query string in graphql format
  -x, --hash           output sha256 instead of filtered content
  -h, --help           help for jf
  -q, --query string   query string in graphql format
```

Query a json object:

```bash
# query from a file
$ jf -f query.graphql obj.json
{"metadata":{"name":"anyuid"},"priority":10,"volumes":["configMap","downwardAPI","emptyDir","persistentVolumeClaim","projected","secret"]}

# read json obj from stdin
$ cat obj.json | jf -f query.graphql -
{"metadata":{"name":"anyuid"},"priority":10,"volumes":["configMap","downwardAPI","emptyDir","persistentVolumeClaim","projected","secret"]}

# inline query
$ jf -q "{metadata{name},apiVersion,priority,volumes,seLinuxContext{type}}" obj.json
{"apiVersion":"v1","metadata":{"name":"anyuid"},"priority":10,"seLinuxContext":{"type":"MustRunAs"},"volumes":["configMap","downwardAPI","emptyDir","persistentVolumeClaim","projected","secret"]}
```

Hash a filtered json object:

```bash
$ jf -x -f query.graphql obj.json
676d78184e27913243d23029939b95cf5744d1b1ccf7ab5eb49210d3283f555f
```

## As a go library

```go
import github.com/ecordell/jf

document := `{
  "allowHostDirVolumePlugin": false,
  "allowHostIPC": false,
  "allowHostNetwork": false,
  "allowHostPID": false,
  "allowHostPorts": false,
  "allowPrivilegeEscalation": true,
  "allowPrivilegedContainer": false,
  "allowedCapabilities": null,
  "apiVersion": "v1",
  "defaultAddCapabilities": null,
  "fsGroup": {
    "type": "RunAsAny"
  },
  "groups": [
    "system:cluster-admins"
  ],
  "kind": "SecurityContextConstraints",
  "metadata": {
    "annotations": {
      "kubernetes.io/description": "anyuid provides all features of the restricted SCC but allows users to run with any UID and any GID."
    },
    "creationTimestamp": "2018-08-27T15:34:59Z",
    "name": "anyuid",
    "resourceVersion": "49",
    "selfLink": "/api/v1/securitycontextconstraints/anyuid",
    "uid": "c1e222ba-aa0e-11e8-aa09-42010a8e0004"
  },
  "priority": 10,
  "readOnlyRootFilesystem": false,
  "requiredDropCapabilities": [
    "MKNOD"
  ],
  "runAsUser": {
    "type": "RunAsAny"
  },
  "seLinuxContext": {
    "type": "MustRunAs"
  },
  "supplementalGroups": {
    "type": "RunAsAny"
  },
  "users": [],
  "volumes": [
    "configMap",
    "downwardAPI",
    "emptyDir",
    "persistentVolumeClaim",
    "projected",
    "secret"
  ]
}
`
filtered := jf.FilterJson(`{metadata{name},apiVersion,priority,volumes,seLinuxContext{type}}`, document)
// value: {"apiVersion":"v1","metadata":{"name":"anyuid"},"priority":10,"seLinuxContext":{"type":"MustRunAs"},"volumes":["configMap","downwardAPI","emptyDir","persistentVolumeClaim","projected","secret"]}
hash := jf.Sha256Json(`{metadata{name},apiVersion,priority,volumes,seLinuxContext{type}}`, document)
// value: 73136c02726e5f6b6339c82bfa6c447c15303533fde65413e89cc7832c337315
```
