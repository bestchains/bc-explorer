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

function docker::build() {
	local -r targets=($@)
	version::git_version
	local -r version="${GIT_VERSION}"
	local output_type="docker"
	if [[ $OUTPUT_TYPE == "registry" ]]; then
		output_type="registry"
	fi
	for target in "${targets[@]}"; do
		local image_name="${IMAGE_NAME:-"${REGISTRY}/${target}:${version}"}"

		echo "Building image for ${BUILD_PLATFORMS}: ${image_name}"
		set -x
		docker buildx build --progress=plain --output=type=$output_type \
			--platform "${BUILD_PLATFORMS}" \
			--build-arg ="${target}" \
			--tag "${image_name}" \
			--file "${ROOT_PATH}/build/dockerfile.${target}" \
			"${ROOT_PATH}"
		set +x
	done
}

function docker::verify_version() {
	if [[ -z "$(command -v docker)" ]]; then
		echo "docker is not installed"
		exit 1
	fi
	if docker buildx --help >/dev/null 2>&1; then
		return 0
	else
		echo "ERROR: docker buildx not available. Docker 19.03 or higher is required with experimental features enabled"
		exit 1
	fi
}
