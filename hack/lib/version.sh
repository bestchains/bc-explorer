#!/usr/bin/env bash

#
# Copyright 2023. The Bestchains Authors.
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
#

set -o errexit
set -o nounset
set -o pipefail

version::ldflags() {
	version::git_version
	GIT_COMMIT=$(git rev-parse HEAD)
	if GIT_STATUS=$(git status --porcelain 2>/dev/null) && [[ -z ${GIT_STATUS} ]]; then
		GIT_TREE_STATE="clean"
	else
		GIT_TREE_STATE="dirty"
	fi
	BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

	local -a ldflags
	function add_ld_flag() {
		local key=${1}
		local val=${2}
		ldflags+=(
			"-X 'k8s.io/component-base/version.${key}=${val}'"
		)
	}

	if [[ -n ${BUILD_DATE-} ]]; then
		add_ld_flag "buildDate" "${BUILD_DATE}"
	fi

	if [[ -n ${GIT_COMMIT-} ]]; then
		add_ld_flag "gitCommit" "${GIT_COMMIT}"
		add_ld_flag "gitTreeState" "${GIT_TREE_STATE}"
	fi

	if [[ -n ${GIT_VERSION-} ]]; then
		add_ld_flag "gitVersion" "${GIT_VERSION}"
	fi

	if [[ -n ${GIT_MAJOR-} && -n ${GIT_MINOR-} ]]; then
		add_ld_flag "gitMajor" "${GIT_MAJOR}"
		add_ld_flag "gitMinor" "${GIT_MINOR}"
	fi

	echo "${ldflags[*]-}"
}

version::git_version() {
	GIT_VERSION=$(git describe --tags --dirty --match "v*" --abbrev=14 2>/dev/null || echo "v0.1.0")
}
