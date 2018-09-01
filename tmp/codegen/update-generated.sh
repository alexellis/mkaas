#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/operator-framework/operator-sdk/minikube/pkg/generated \
github.com/operator-framework/operator-sdk/minikube/pkg/apis \
alexellis:v1alpha1 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
