package main

import (
	"testing"
	"github.com/json-iterator/go/require"
	"github.com/sirupsen/logrus"
	"fmt"
)

var scc = `{
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
func TestFilterJson(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	query := `{metadata{name},apiVersion,priority,volumes,seLinuxContext{type}}`
	expected := `{"apiVersion":"v1","metadata":{"name":"anyuid"},"priority":10,"seLinuxContext":{"type":"MustRunAs"},"volumes":["configMap","downwardAPI","emptyDir","persistentVolumeClaim","projected","secret"]}`
	require.Equal(t, expected, string(FilterJson(query, []byte(scc))))
	require.Equal(t, "73136c02726e5f6b6339c82bfa6c447c15303533fde65413e89cc7832c337315", fmt.Sprintf("%x", Sha256Json(query, []byte(scc))))
}


