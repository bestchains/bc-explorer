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
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func UnmarshalProcessedTransaction(bytes []byte) (*peer.ProcessedTransaction, error) {
	processedTx := &peer.ProcessedTransaction{}
	err := proto.Unmarshal(bytes, processedTx)
	if err != nil {
		return nil, err
	}
	return processedTx, nil
}

func UnmarshalConfigEnvelope(bytes []byte) (*common.ConfigEnvelope, error) {
	configEnv := &common.ConfigEnvelope{}
	err := proto.Unmarshal(bytes, configEnv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config envelope")
	}
	return configEnv, nil
}

func UnmarshalConfig(bytes []byte) (*common.Config, error) {
	configEnv, err := UnmarshalConfigEnvelope(bytes)
	if err != nil {
		return nil, err
	}
	return configEnv.GetConfig(), nil
}

func UnmarshalConfigUpdateEnvelope(bytes []byte) (*common.ConfigUpdateEnvelope, error) {
	configUpdateEnv := &common.ConfigUpdateEnvelope{}
	err := proto.Unmarshal(bytes, configUpdateEnv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config update envelope")
	}
	return configUpdateEnv, nil
}

func UnmarshalConfigUpdate(bytes []byte) (*common.ConfigUpdate, error) {
	configUpdateEnv, err := UnmarshalConfigUpdateEnvelope(bytes)
	if err != nil {
		return nil, err
	}
	configUpdate := &common.ConfigUpdate{}
	err = proto.Unmarshal(configUpdateEnv.GetConfigUpdate(), configUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config update envelope")
	}
	return configUpdate, nil
}

// UnmarshalBlock unmarshals bytes to a Block
func UnmarshalBlock(encoded []byte) (*common.Block, error) {
	block := &common.Block{}
	err := proto.Unmarshal(encoded, block)
	return block, errors.Wrap(err, "error unmarshaling Block")
}

// UnmarshalChaincodeDeploymentSpec unmarshals bytes to a ChaincodeDeploymentSpec
func UnmarshalChaincodeDeploymentSpec(code []byte) (*peer.ChaincodeDeploymentSpec, error) {
	cds := &peer.ChaincodeDeploymentSpec{}
	err := proto.Unmarshal(code, cds)
	return cds, errors.Wrap(err, "error unmarshaling ChaincodeDeploymentSpec")
}

// UnmarshalChaincodeInvocationSpec unmarshals bytes to a ChaincodeInvocationSpec
func UnmarshalChaincodeInvocationSpec(encoded []byte) (*peer.ChaincodeInvocationSpec, error) {
	cis := &peer.ChaincodeInvocationSpec{}
	err := proto.Unmarshal(encoded, cis)
	return cis, errors.Wrap(err, "error unmarshaling ChaincodeInvocationSpec")
}

// UnmarshalPayload unmarshals bytes to a Payload
func UnmarshalPayload(encoded []byte) (*common.Payload, error) {
	payload := &common.Payload{}
	err := proto.Unmarshal(encoded, payload)
	return payload, errors.Wrap(err, "error unmarshaling Payload")
}

// UnmarshalEnvelope unmarshals bytes to a Envelope
func UnmarshalEnvelope(encoded []byte) (*common.Envelope, error) {
	envelope := &common.Envelope{}
	err := proto.Unmarshal(encoded, envelope)
	return envelope, errors.Wrap(err, "error unmarshaling Envelope")
}

// UnmarshalChannelHeader unmarshals bytes to a ChannelHeader
func UnmarshalChannelHeader(bytes []byte) (*common.ChannelHeader, error) {
	chdr := &common.ChannelHeader{}
	err := proto.Unmarshal(bytes, chdr)
	return chdr, errors.Wrap(err, "error unmarshaling ChannelHeader")
}

// UnmarshalChaincodeID unmarshals bytes to a ChaincodeID
func UnmarshalChaincodeID(bytes []byte) (*peer.ChaincodeID, error) {
	ccid := &peer.ChaincodeID{}
	err := proto.Unmarshal(bytes, ccid)
	return ccid, errors.Wrap(err, "error unmarshaling ChaincodeID")
}

// UnmarshalSignatureHeader unmarshals bytes to a SignatureHeader
func UnmarshalSignatureHeader(bytes []byte) (*common.SignatureHeader, error) {
	sh := &common.SignatureHeader{}
	err := proto.Unmarshal(bytes, sh)
	return sh, errors.Wrap(err, "error unmarshaling SignatureHeader")
}

func UnmarshalSerializedIdentity(bytes []byte) (*msp.SerializedIdentity, error) {
	sid := &msp.SerializedIdentity{}
	err := proto.Unmarshal(bytes, sid)
	return sid, errors.Wrap(err, "error unmarshaling SerializedIdentity")
}

// UnmarshalHeader unmarshals bytes to a Header
func UnmarshalHeader(bytes []byte) (*common.Header, error) {
	hdr := &common.Header{}
	err := proto.Unmarshal(bytes, hdr)
	return hdr, errors.Wrap(err, "error unmarshaling Header")
}

// UnmarshalChaincodeHeaderExtension unmarshals bytes to a ChaincodeHeaderExtension
func UnmarshalChaincodeHeaderExtension(hdrExtension []byte) (*peer.ChaincodeHeaderExtension, error) {
	chaincodeHdrExt := &peer.ChaincodeHeaderExtension{}
	err := proto.Unmarshal(hdrExtension, chaincodeHdrExt)
	return chaincodeHdrExt, errors.Wrap(err, "error unmarshaling ChaincodeHeaderExtension")
}

// UnmarshalProposalResponse unmarshals bytes to a ProposalResponse
func UnmarshalProposalResponse(prBytes []byte) (*peer.ProposalResponse, error) {
	proposalResponse := &peer.ProposalResponse{}
	err := proto.Unmarshal(prBytes, proposalResponse)
	return proposalResponse, errors.Wrap(err, "error unmarshaling ProposalResponse")
}

// UnmarshalChaincodeAction unmarshals bytes to a ChaincodeAction
func UnmarshalChaincodeAction(caBytes []byte) (*peer.ChaincodeAction, error) {
	chaincodeAction := &peer.ChaincodeAction{}
	err := proto.Unmarshal(caBytes, chaincodeAction)
	return chaincodeAction, errors.Wrap(err, "error unmarshaling ChaincodeAction")
}

// UnmarshalResponse unmarshals bytes to a Response
func UnmarshalResponse(resBytes []byte) (*peer.Response, error) {
	response := &peer.Response{}
	err := proto.Unmarshal(resBytes, response)
	return response, errors.Wrap(err, "error unmarshaling Response")
}

// UnmarshalChaincodeEvents unmarshals bytes to a ChaincodeEvent
func UnmarshalChaincodeEvents(eBytes []byte) (*peer.ChaincodeEvent, error) {
	chaincodeEvent := &peer.ChaincodeEvent{}
	err := proto.Unmarshal(eBytes, chaincodeEvent)
	return chaincodeEvent, errors.Wrap(err, "error unmarshaling ChaicnodeEvent")
}

// UnmarshalProposalResponsePayload unmarshals bytes to a ProposalResponsePayload
func UnmarshalProposalResponsePayload(prpBytes []byte) (*peer.ProposalResponsePayload, error) {
	prp := &peer.ProposalResponsePayload{}
	err := proto.Unmarshal(prpBytes, prp)
	return prp, errors.Wrap(err, "error unmarshaling ProposalResponsePayload")
}

// UnmarshalProposal unmarshals bytes to a Proposal
func UnmarshalProposal(propBytes []byte) (*peer.Proposal, error) {
	prop := &peer.Proposal{}
	err := proto.Unmarshal(propBytes, prop)
	return prop, errors.Wrap(err, "error unmarshaling Proposal")
}

// UnmarshalTransaction unmarshals bytes to a Transaction
func UnmarshalTransaction(txBytes []byte) (*peer.Transaction, error) {
	tx := &peer.Transaction{}
	err := proto.Unmarshal(txBytes, tx)
	return tx, errors.Wrap(err, "error unmarshaling Transaction")
}

// UnmarshalChaincodeActionPayload unmarshals bytes to a ChaincodeActionPayload
func UnmarshalChaincodeActionPayload(capBytes []byte) (*peer.ChaincodeActionPayload, error) {
	cap := &peer.ChaincodeActionPayload{}
	err := proto.Unmarshal(capBytes, cap)
	return cap, errors.Wrap(err, "error unmarshaling ChaincodeActionPayload")
}

// UnmarshalChaincodeProposalPayload unmarshals bytes to a ChaincodeProposalPayload
func UnmarshalChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, errors.Wrap(err, "error unmarshaling ChaincodeProposalPayload")
}

// UnmarshalPayloadOrPanic unmarshals bytes to a Payload structure or panics
// on error
func UnmarshalPayloadOrPanic(encoded []byte) *common.Payload {
	payload, err := UnmarshalPayload(encoded)
	if err != nil {
		panic(err)
	}
	return payload
}

// UnmarshalEnvelopeOrPanic unmarshals bytes to an Envelope structure or panics
// on error
func UnmarshalEnvelopeOrPanic(encoded []byte) *common.Envelope {
	envelope, err := UnmarshalEnvelope(encoded)
	if err != nil {
		panic(err)
	}
	return envelope
}

// UnmarshalBlockOrPanic unmarshals bytes to an Block or panics
// on error
func UnmarshalBlockOrPanic(encoded []byte) *common.Block {
	block, err := UnmarshalBlock(encoded)
	if err != nil {
		panic(err)
	}
	return block
}

// UnmarshalChannelHeaderOrPanic unmarshals bytes to a ChannelHeader or panics
// on error
func UnmarshalChannelHeaderOrPanic(bytes []byte) *common.ChannelHeader {
	chdr, err := UnmarshalChannelHeader(bytes)
	if err != nil {
		panic(err)
	}
	return chdr
}

// UnmarshalSignatureHeaderOrPanic unmarshals bytes to a SignatureHeader or panics
// on error
func UnmarshalSignatureHeaderOrPanic(bytes []byte) *common.SignatureHeader {
	sighdr, err := UnmarshalSignatureHeader(bytes)
	if err != nil {
		panic(err)
	}
	return sighdr
}

func UnmarshalRWSet(bytes []byte) (*rwset.TxReadWriteSet, error) {
	rwSet := &rwset.TxReadWriteSet{}
	err := proto.Unmarshal(bytes, rwSet)
	if err != nil {
		return nil, err
	}
	return rwSet, nil
}
