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

#https://textkool.com/en/ascii-art-generator?hl=full&vl=default&font=DOS%20Rebel&text=ARBITER
GREETING='
-------------------------------------------------------------------------------------------------
 █████                                                        ████                                       
░░███                                                        ░░███                                       
 ░███████   ██████              ██████  █████ █████ ████████  ░███   ██████  ████████   ██████  ████████ 
 ░███░░███ ███░░███ ██████████ ███░░███░░███ ░░███ ░░███░░███ ░███  ███░░███░░███░░███ ███░░███░░███░░███
 ░███ ░███░███ ░░░ ░░░░░░░░░░ ░███████  ░░░█████░   ░███ ░███ ░███ ░███ ░███ ░███ ░░░ ░███████  ░███ ░░░ 
 ░███ ░███░███  ███           ░███░░░    ███░░░███  ░███ ░███ ░███ ░███ ░███ ░███     ░███░░░   ░███     
 ████████ ░░██████            ░░██████  █████ █████ ░███████  █████░░██████  █████    ░░██████  █████    
░░░░░░░░   ░░░░░░              ░░░░░░  ░░░░░ ░░░░░  ░███░░░  ░░░░░  ░░░░░░  ░░░░░      ░░░░░░  ░░░░░     
                                                    ░███                                                 
                                                    █████                                                
                                                   ░░░░░                                                 
------------------------------------------------------------------------------------------------------
'
ROOT_PATH=$(git rev-parse --show-toplevel)

readonly PACKAGE_NAME="github.com/bestchains/bc-explorer"
readonly OUTPUT_DIR="${ROOT_PATH}/_output"
readonly BUILD_GOPATH="${OUTPUT_DIR}/go"
