/*
Copyright 2023 The Bestchains Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

type Block struct {
	BlockHash         string `pg:"blockHash,pk"`
	Network           string `pg:"network"`
	BlockNumber       uint64 `pg:"blockNumber"`
	PrevioudBlockHash string `pg:"preBlockHash"`
	DataHash          string `pg:"dataHash"`
	CreatedAt         int64  `pg:"createdAt"`
	BlockSize         int    `pg:"blockSize"`
	TxCount           int    `pg:"txCount"`
}
