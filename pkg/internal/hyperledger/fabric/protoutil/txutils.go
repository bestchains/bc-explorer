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

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protoutil

import (
	"encoding/json"
	"strings"

	"github.com/bestchains/bc-explorer/pkg/internal/hyperledger/fabric/rwsetutil"
	"github.com/bestchains/bc-explorer/pkg/models"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
)

// GetTransactionFromEnvelope
func GetTransactionFromEnvelope(txEnvelopBytes []byte) (*models.Transaction, error) {
	var err error

	processedTx, err := UnmarshalProcessedTransaction(txEnvelopBytes)
	if err != nil {
		return nil, err
	}

	txEnvelope, err := UnmarshalEnvelope(txEnvelopBytes)
	if err != nil {
		return nil, err
	}

	txPayload, err := UnmarshalPayload(txEnvelope.Payload)
	if err != nil {
		return nil, err
	}

	sighdr, err := UnmarshalSignatureHeader(txPayload.Header.SignatureHeader)
	if err != nil {
		return nil, err
	}
	creator, err := UnmarshalSerializedIdentity(sighdr.Creator)
	if err != nil {
		return nil, err
	}

	chdr, err := UnmarshalChannelHeader(txPayload.Header.ChannelHeader)
	if err != nil {
		return nil, err
	}

	tx := &models.Transaction{
		ID:             chdr.TxId,
		CreatedAt:      chdr.Timestamp.AsTime().Unix(),
		Creator:        creator.GetMspid(),
		ValidationCode: processedTx.GetValidationCode(),
	}

	switch chdr.Type {
	case int32(common.HeaderType_CONFIG):
		tx.Type = models.Config
		config, err := UnmarshalConfig(txPayload.Data)
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		tx.Payload = raw
	case int32(common.HeaderType_CONFIG_UPDATE):
		tx.Type = models.ConfigUpdate
		configUpdate, err := UnmarshalConfigUpdate(txPayload.Data)
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(configUpdate)
		if err != nil {
			return nil, err
		}
		tx.Payload = raw
	case int32(common.HeaderType_ENDORSER_TRANSACTION):
		tx.Type = models.EndorserTransaction

		invocationSpec, action, err := GetTxDetailsFromPayload(txPayload)
		if err != nil {
			return nil, err
		}

		tx.ChaincodeID = action.ChaincodeId.Name + "_" + action.ChaincodeId.Version
		if len(invocationSpec.ChaincodeSpec.Input.Args) > 0 {
			tx.Method = string(invocationSpec.ChaincodeSpec.Input.Args[0])
			for _, arg := range invocationSpec.ChaincodeSpec.Input.Args[1:] {
				tx.Args = append(tx.Args, string(arg))
			}
		}

		rwset, err := UnmarshalRWSet(action.GetResults())
		if err != nil {
			return nil, err
		}
		txRWSet, err := rwsetutil.TxRwSetFromProtoMsg(rwset)
		if err != nil {
			return nil, err
		}

		fabRWSets := make([]models.FabRWSet, len(txRWSet.NsRwSets))
		for index, rwset := range txRWSet.NsRwSets {
			fabRWSet := models.FabRWSet{
				Namespace: rwset.NameSpace,
			}
			reads := make([]models.Read, 0)
			writes := make([]models.Write, 0)
			for _, read := range rwset.KvRwSet.Reads {
				reads = append(reads, models.Read{
					Key:     strings.Replace(read.GetKey(), "\u0000", "", -1),
					Version: read.GetVersion().String(),
				})
			}
			for _, write := range rwset.KvRwSet.Writes {
				writes = append(writes, models.Write{
					Key:      strings.Replace(write.GetKey(), "\u0000", "", -1),
					Value:    string(write.GetValue()),
					IsDelete: write.IsDelete,
				})
			}
			fabRWSet.Reads = reads
			fabRWSet.Writes = writes

			fabRWSets[index] = fabRWSet
		}

		raw, err := json.Marshal(fabRWSets)
		if err != nil {
			return nil, err
		}
		tx.Payload = raw
	default:
	}

	return tx, nil
}

// GetPayloads gets the underlying payload objects in a TransactionAction
func GetTxDetailsFromPayload(payload *common.Payload) (*peer.ChaincodeInvocationSpec, *peer.ChaincodeAction, error) {
	payloadTx, err := UnmarshalTransaction(payload.Data)
	if err != nil {
		return nil, nil, err
	}
	if len(payloadTx.Actions) == 0 {
		return nil, nil, errors.New("at least one TransactionAction required")
	}

	ccPayload, err := UnmarshalChaincodeActionPayload(payloadTx.Actions[0].Payload)
	if err != nil {
		return nil, nil, err
	}

	if ccPayload.Action == nil || ccPayload.Action.ProposalResponsePayload == nil {
		return nil, nil, errors.New("no payload in ChaincodeActionPayload")
	}

	ccProposalPayload, err := UnmarshalChaincodeProposalPayload(ccPayload.ChaincodeProposalPayload)
	if err != nil {
		return nil, nil, err
	}
	invocationSpec, err := UnmarshalChaincodeInvocationSpec(ccProposalPayload.Input)
	if err != nil {
		return nil, nil, err
	}

	pRespPayload, err := UnmarshalProposalResponsePayload(ccPayload.Action.ProposalResponsePayload)
	if err != nil {
		return nil, nil, err
	}

	if pRespPayload.Extension == nil {
		return nil, nil, errors.New("response payload is missing extension")
	}

	respPayload, err := UnmarshalChaincodeAction(pRespPayload.Extension)
	if err != nil {
		return nil, nil, err
	}
	return invocationSpec, respPayload, nil
}
