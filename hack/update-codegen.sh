#!/bin/bash

# Update generated client code
~/go/src/k8s.io/code-generator/generate-groups.sh all github.com/damemi/opentelemetry-operator/pkg/generated github.com/damemi/opentelemetry-operator/pkg/apis otel:v1alpha1

# Update CRDs
controller-gen crd:trivialVersions=true paths=./pkg/apis/... output:dir=./manifests

# Update bindata
go-bindata -nocompress -nometadata -pkg assets -o pkg/operator/assets/assets.go bindata/
