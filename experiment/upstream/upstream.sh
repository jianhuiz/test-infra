#!/usr/bin/env bash
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

pushd `dirname $0`

#go get k8s.io/test-infra/prow/github
mkdir -p $GOPATH/src/k8s.io/
ln -s `realpath ../../../test-infra` $GOPATH/src/k8s.io/

go build
./upstream --token /etc/token/bot-github-token --org $REPO_OWNER --repo $REPO_NAME --pull $PULL_NUMBER

popd
