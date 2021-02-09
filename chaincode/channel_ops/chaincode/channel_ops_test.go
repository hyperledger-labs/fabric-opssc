/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/stretchr/testify/require"
	"github.com/hyperledger-labs/fabric-opssc/chaincode/channel_ops/chaincode/mocks"
)

var (
	update                 = marshalProtoOrPanic(&common.ConfigUpdate{ChannelId: "mychannel", ReadSet: &common.ConfigGroup{}, WriteSet: &common.ConfigGroup{}})
	updateBase64           = base64.StdEncoding.EncodeToString(update)
	updateOrgsInAppChannel = marshalProtoOrPanic(&common.ConfigUpdate{ChannelId: "mychannel",
		ReadSet: &common.ConfigGroup{},
		WriteSet: &common.ConfigGroup{
			Groups: map[string]*common.ConfigGroup{
				"Application": {
					Groups: map[string]*common.ConfigGroup{
						"Org1MSP": {}, "Org2MSP": {}, "Org3MSP": {}, "Org4MSP": {},
					},
				},
			},
		},
	})
	updateOrgsInAppChannelBase64 = base64.StdEncoding.EncodeToString(updateOrgsInAppChannel)
	updateOrgsInSystemChannel    = marshalProtoOrPanic(&common.ConfigUpdate{ChannelId: "system-channel",
		ReadSet: &common.ConfigGroup{},
		WriteSet: &common.ConfigGroup{
			Groups: map[string]*common.ConfigGroup{
				"Consortiums": {
					Groups: map[string]*common.ConfigGroup{
						"Consortium": {
							Groups: map[string]*common.ConfigGroup{
								"Org1MSP": {}, "Org2MSP": {}, "Org3MSP": {}, "Org4MSP": {},
							},
						},
					},
				},
			},
		},
	})
	updateOrgsInSystemChannelBase64 = base64.StdEncoding.EncodeToString(updateOrgsInSystemChannel)

	signature       = marshalProtoOrPanic(&common.ConfigSignature{Signature: []byte("mysignature"), SignatureHeader: []byte("myheader")})
	signatureBase64 = base64.StdEncoding.EncodeToString(signature)

	org1MSP = marshalProtoOrPanic(&msp.SerializedIdentity{Mspid: "Org1MSP", IdBytes: []byte("myid")})
	org2MSP = marshalProtoOrPanic(&msp.SerializedIdentity{Mspid: "Org2MSP", IdBytes: []byte("myid")})

	opsProfileJSON = `[{"Command": "setorg", "Parameters": {"OrgType": "Application|Orderer", "Org": {"Name": "Org1MSP", "ID": "Org1MSP", "MSP": {}, "Policies": {}}}}]`
	opsProfile     = unmarshalJSONOrPanic(opsProfileJSON)
)

// marshalProtoOrPanic is a helper for proto marshal.
func marshalProtoOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}

	return data
}

func unmarshalJSONOrPanic(data string) interface{} {
	var v interface{}
	err := json.Unmarshal([]byte(data), &v)
	if err != nil {
		panic(err)
	}
	return v
}

//go:generate counterfeiter -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

// Dummy implementation to create compose key
func createComposeKey(objectType string, keys []string) (string, error) {
	return objectType + "_" + strings.Join(keys, "_"), nil
}

func TestRequestProposal(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey

	sc := SmartContract{}

	// Case: Request a proposal to update a channel
	input := ProposalInput{
		ID:           "request-1",
		ChannelID:    "mychannel",
		Description:  "test description",
		OpsProfile:   opsProfile,
		ConfigUpdate: updateBase64,
		Signature:    signatureBase64,
	}
	chaincodeStub.GetCreatorReturns(org1MSP, nil)
	actualID, err := sc.RequestProposal(transactionContext, input)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "proposal_request-1", key)
	expectedSignatures := make(map[string]string)
	expectedSignatures["Org1MSP"] = input.Signature
	expectedProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          input.ID,
		Creator:     "Org1MSP",
		ChannelID:   input.ChannelID,
		Description: input.Description,
		Action:      UpdateAction,
		OpsProfile:  opsProfile,
		Status:      Proposed,
		Artifacts: Artifacts{
			ConfigUpdate: input.ConfigUpdate,
			Signatures:   expectedSignatures,
		},
	}
	expectedJSON, err := json.Marshal(expectedProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))
	require.Equal(t, input.ID, actualID)
	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "newProposalEvent.request-1", eventName)
	expectedEventPayload := []byte(input.ID)
	require.Equal(t, string(expectedEventPayload), string(eventPayload))

	// Case: Fail to request when an invalid action is inputted
	input = ProposalInput{
		ID:           "request-1",
		ChannelID:    "mychannel",
		Action:       "invalid action",
		Description:  "test description",
		ConfigUpdate: updateBase64,
		Signature:    signatureBase64,
	}
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "incorrect operation type - expecting update or create")

	// Case: Fail to request when the ConfigUpdate is invalid
	input = ProposalInput{
		ID:           "request-1",
		ChannelID:    "mychannel",
		Description:  "test description",
		ConfigUpdate: "Invalid config update",
		Signature:    signatureBase64,
	}
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "error happened decoding the configUpdate base64 string: illegal base64 data at input byte 7")

	// Case: Fail to request when the Signature is invalid
	input = ProposalInput{
		ID:           "request-1",
		ChannelID:    "mychannel",
		Description:  "test description",
		ConfigUpdate: updateBase64,
		Signature:    "Invalid signature",
	}
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "error happened decoding the signature base64 string: illegal base64 data at input byte 7")

	// Case: Fail to request when the proposal ID is already in use
	input = ProposalInput{
		ID:           "request-1",
		ChannelID:    "mychannel",
		Description:  "test description",
		ConfigUpdate: updateBase64,
		Signature:    signatureBase64,
	}
	dummyState := Proposal{}
	dummyJSON, err := json.Marshal(dummyState)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(dummyJSON, nil)
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "proposalID already in use")

	// Case: Fail to request when getting MSP ID is failed
	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.GetCreatorReturns(nil, fmt.Errorf("failed to get MSP ID"))
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "failed to get MSP ID: error happened reading the transaction creator: failed to get MSP ID")

	// Case: Fail to request when putProposal occurs an error
	chaincodeStub.GetCreatorReturns(org1MSP, nil)
	chaincodeStub.CreateCompositeKeyReturns("", fmt.Errorf("failed to create composite key"))
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "failed to put the proposal: error happend creating composite key for proposal: failed to create composite key")

	// Case: Fail to request when setEvent occurs an error
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.SetEventReturns(fmt.Errorf("failed to set event"))
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "error happened emitting event: failed to set event")
}

func TestVote(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	sc := SmartContract{}

	proposalID := "request-1"

	// Case: Request a vote to a proposal (to create a channel) and the votes do not meet the majority
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	baseProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          proposalID,
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      CreationAction,
		Status:      Proposed,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   nil,
		},
	}
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	baseChannel := Channel{
		ObjectType:  ChannelObjectType,
		ID:          "system-channel",
		ChannelType: SystemChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": "",
			"Org3MSP": ""},
	}
	baseChannelJSON, err := json.Marshal(baseChannel)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(1, baseChannelJSON, nil)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: baseChannelJSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	err = sc.Vote(transactionContext, proposalID, signatureBase64)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "proposal_request-1", key)
	expectedProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          proposalID,
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      CreationAction,
		Status:      Proposed,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org2MSP": signatureBase64},
		},
	}
	expectedJSON, err := json.Marshal(expectedProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))
	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "newVoteEvent.request-1", eventName)
	expectedEventPayload := []byte(proposalID)
	require.Equal(t, string(expectedEventPayload), string(eventPayload))

	// Case: Request a vote to a proposal (to update a channel) and the votes meet the majority
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	baseProposal = Proposal{
		ObjectType:  ProposalObjectType,
		ID:          proposalID,
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Proposed,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64},
		},
	}
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(2, baseProposalJSON, nil)

	baseChannel = Channel{
		ObjectType:  ChannelObjectType,
		ID:          "mychannel",
		ChannelType: ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": "",
			"Org3MSP": ""},
	}
	baseChannelJSON, err = json.Marshal(baseChannel)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(3, baseChannelJSON, nil)

	err = sc.Vote(transactionContext, proposalID, signatureBase64)
	require.NoError(t, err)
	key, state = chaincodeStub.PutStateArgsForCall(1)
	require.Equal(t, "proposal_request-1", key)
	expectedProposal = Proposal{
		ObjectType:  ProposalObjectType,
		ID:          proposalID,
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	expectedJSON, err = json.Marshal(expectedProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))
	eventName, eventPayload = chaincodeStub.SetEventArgsForCall(1)
	require.Equal(t, "readyToUpdateConfigEvent.request-1", eventName)

	expectedReadyToUpdateConfigEventPayload := EventDetail{
		ProposalID:       proposalID,
		OperationTargets: []string{"Org2MSP"},
	}
	expectedEventPayloadJSON, err := json.Marshal(expectedReadyToUpdateConfigEventPayload)
	require.NoError(t, err)
	require.Equal(t, string(expectedEventPayloadJSON), string(eventPayload))

	// Case: Fail to vote when setEvent occurs an error
	chaincodeStub.GetStateReturnsOnCall(4, baseProposalJSON, nil)
	chaincodeStub.GetStateReturnsOnCall(5, baseChannelJSON, nil)
	chaincodeStub.SetEventReturns(fmt.Errorf("failed to set event"))
	err = sc.Vote(transactionContext, proposalID, signatureBase64)
	require.EqualError(t, err, "error happened emitting event: failed to set event")

	// Case: Fail to request when putProposal occurs an error
	chaincodeStub.GetStateReturnsOnCall(6, baseProposalJSON, nil)
	chaincodeStub.GetStateReturnsOnCall(7, baseChannelJSON, nil)
	chaincodeStub.SetEventReturns(nil)
	cc := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(cc+2, "", fmt.Errorf("failed to create composite key"))
	err = sc.Vote(transactionContext, proposalID, signatureBase64)
	require.EqualError(t, err, "failed to put the proposal: error happend creating composite key for proposal: failed to create composite key")

	// Case: Fail to vote when checking number of votes fails
	chaincodeStub.GetStateReturnsOnCall(8, baseProposalJSON, nil)
	chaincodeStub.GetStateReturnsOnCall(9, nil, fmt.Errorf("failed to get state"))
	err = sc.Vote(transactionContext, proposalID, signatureBase64)
	require.EqualError(t, err, "fail to check whether the votes passed: error happened checking to meet criteria: fail to get the num of organizations: failed to read channel: failed to read from world state: failed to get state")

	// Case: Fail to vote when the Signature is invalid
	err = sc.Vote(transactionContext, proposalID, "Invalid signature")
	require.EqualError(t, err, "error happened decoding the signature base64 string: illegal base64 data at input byte 7")

	// Case: Fail to vote when getting MSP ID is failed
	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.GetCreatorReturns(nil, fmt.Errorf("failed to get MSP ID"))
	err = sc.Vote(transactionContext, proposalID, signatureBase64)
	require.EqualError(t, err, "failed to get MSP ID: error happened reading the transaction creator: failed to get MSP ID")

	// Case: Fail to vote when getting the proposal is failed
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	err = sc.Vote(transactionContext, proposalID, signatureBase64)
	require.EqualError(t, err, "failed to get proposal: proposal not found")
}

func TestNotifyCommitResult(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	sc := SmartContract{}

	// Case: Notify commit result to update an application channel (without updating the organizations)
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	baseProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	err = sc.NotifyCommitResult(transactionContext, "request-1")
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "proposal_request-1", key)
	expectedProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Committed,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	expectedProposalJSON, err := json.Marshal(expectedProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedProposalJSON), string(state))
	require.Equal(t, 0, chaincodeStub.SetEventCallCount())
}

func TestNotifyCommitResultWhenSetEventFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	sc := SmartContract{}

	// Case: Fail to notify commit result when set event fails
	chaincodeStub.SetEventReturns(fmt.Errorf("failed to set event"))
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	baseProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      CreationAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateOrgsInAppChannelBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	err = sc.NotifyCommitResult(transactionContext, "request-1")
	require.EqualError(t, err, "error happened emitting event: failed to set event")
}

func TestNotifyCommitResultToCreateAppChannel(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	sc := SmartContract{}

	// Case: Notify commit result to create an application channel
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	baseProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      CreationAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateOrgsInAppChannelBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	err = sc.NotifyCommitResult(transactionContext, "request-1")
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "proposal_request-1", key)
	expectedProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      CreationAction,
		Status:      Committed,
		Artifacts: Artifacts{
			ConfigUpdate: updateOrgsInAppChannelBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	expectedProposalJSON, err := json.Marshal(expectedProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedProposalJSON), string(state))
	key, state = chaincodeStub.PutStateArgsForCall(1)
	require.Equal(t, "channel_mychannel", key)
	expectedChannel := Channel{
		ObjectType:    ChannelObjectType,
		ID:            "mychannel",
		ChannelType:   ApplicationChannelType,
		Organizations: map[string]string{"Org1MSP": "", "Org2MSP": "", "Org3MSP": "", "Org4MSP": ""},
	}
	expectedChannelJSON, err := json.Marshal(expectedChannel)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedChannelJSON), string(state))

}

func TestNotifyCommitResultToUpdateOrgsInSystemChannel(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	sc := SmartContract{}

	// Case: Notify commit result to update an system channel (the update includes changing organizations)
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	baseProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "system-channel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateOrgsInSystemChannelBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	baseChannel := Channel{
		ObjectType:  ChannelObjectType,
		ID:          "system-channel",
		ChannelType: SystemChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": "",
			"Org3MSP": ""},
	}
	baseChannelJSON, err := json.Marshal(baseChannel)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(1, baseChannelJSON, nil)
	chaincodeStub.GetStateReturnsOnCall(2, baseChannelJSON, nil)

	err = sc.NotifyCommitResult(transactionContext, "request-1")
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "proposal_request-1", key)
	expectedProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "system-channel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Committed,
		Artifacts: Artifacts{
			ConfigUpdate: updateOrgsInSystemChannelBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	expectedProposalJSON, err := json.Marshal(expectedProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedProposalJSON), string(state))
	key, state = chaincodeStub.PutStateArgsForCall(1)
	require.Equal(t, "channel_system-channel", key)
	expectedChannel := Channel{
		ObjectType:    ChannelObjectType,
		ID:            "system-channel",
		ChannelType:   SystemChannelType,
		Organizations: map[string]string{"Org1MSP": "", "Org2MSP": "", "Org3MSP": "", "Org4MSP": ""},
	}
	expectedChannelJSON, err := json.Marshal(expectedChannel)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedChannelJSON), string(state))
	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "updateConfigEvent.request-1", eventName)
	require.Equal(t, string("request-1"), string(eventPayload))
}

func TestNotifyCommitResultToUpdateOrgsWhenTheConsortiumGroupIsMissing(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	sc := SmartContract{}

	updateMissingGroup := marshalProtoOrPanic(&common.ConfigUpdate{ChannelId: "system-channel",
		ReadSet: &common.ConfigGroup{},
		WriteSet: &common.ConfigGroup{
			Groups: map[string]*common.ConfigGroup{
				"Consortiums": {
					Groups: map[string]*common.ConfigGroup{},
				},
			},
		},
	})
	updateMissingGroupBase64 := base64.StdEncoding.EncodeToString(updateMissingGroup)

	// Case: Notify commit result to update an system channel (the consortium group in the update is missing)
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	baseProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "system-channel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateMissingGroupBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	err = sc.NotifyCommitResult(transactionContext, "request-1")
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "proposal_request-1", key)
	expectedProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "system-channel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Committed,
		Artifacts: Artifacts{
			ConfigUpdate: updateMissingGroupBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	expectedProposalJSON, err := json.Marshal(expectedProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedProposalJSON), string(state))
	require.Equal(t, 0, chaincodeStub.SetEventCallCount())
}

func TestNotifyCommitResultFailure(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	sc := SmartContract{}

	// Case: Fail to notify commit result when the proposal is not found
	err := sc.NotifyCommitResult(transactionContext, "request-1")
	require.EqualError(t, err, "failed to get proposal: proposal not found")

	// Case: Fail to notify commit result when the proposal status is not approved
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	baseProposal := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Proposed,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	err = sc.NotifyCommitResult(transactionContext, "request-1")
	require.EqualError(t, err, "proposal is not yet approved or already committed")

	// Case: Fail to notify commit result when putProposal occurs an error
	baseProposal = Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	baseProposalJSON, err = json.Marshal(baseProposal)
	gsCallCount := chaincodeStub.GetStateCallCount()
	chaincodeStub.GetStateReturnsOnCall(gsCallCount, baseProposalJSON, nil)
	ccCallCount := chaincodeStub.GetStateCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(ccCallCount+1, "", fmt.Errorf("failed to create composite key"))
	err = sc.NotifyCommitResult(transactionContext, "request-1")
	require.EqualError(t, err, "failed to put proposal: error happend creating composite key for proposal: failed to create composite key")

	// Case: Fail to notify commit when the ConfigUpdate is invalid
	baseProposal = Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: "Invalid config update",
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	err = sc.NotifyCommitResult(transactionContext, "request-1")
	require.EqualError(t, err, "error happened decoding the configUpdate base64 string: illegal base64 data at input byte 7")
}

func TestGetAllProposals(t *testing.T) {

	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	sc := &SmartContract{}

	// Case: Get 2 proposals
	proposal1 := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "system-channel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	proposal1JSON, err := json.Marshal(proposal1)
	require.NoError(t, err)

	proposal2 := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-2",
		Creator:     "Org2MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	proposal2JSON, err := json.Marshal(proposal2)
	require.NoError(t, err)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, true)
	iterator.HasNextReturnsOnCall(2, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: proposal1JSON}, nil)
	iterator.NextReturnsOnCall(1, &queryresult.KV{Value: proposal2JSON}, nil)

	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	proposals, err := sc.GetAllProposals(transactionContext)
	require.NoError(t, err)
	require.Equal(t, map[string]*Proposal{"request-1": &proposal1, "request-2": &proposal2}, proposals)

	// Case: Fail to get proposals when failed retrieving next item
	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	proposals, err = sc.GetAllProposals(transactionContext)
	require.EqualError(t, err, "error happened iterating over available proposals: failed retrieving next item")
	require.Nil(t, proposals)

	// Case: Fail to get proposals when failed retrieving all proposals
	chaincodeStub.GetStateByPartialCompositeKeyReturns(nil, fmt.Errorf("failed retrieving all proposals"))
	proposals, err = sc.GetAllProposals(transactionContext)
	require.EqualError(t, err, "error happened reading keys from ledger: failed retrieving all proposals")
	require.Nil(t, proposals)
}

func TestGetProposal(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	sc := &SmartContract{}

	// Case: Get the proposal
	expected := Proposal{
		ObjectType:  ProposalObjectType,
		ID:          "request-1",
		Creator:     "Org1MSP",
		ChannelID:   "mychannel",
		Description: "test description",
		Action:      UpdateAction,
		Status:      Approved,
		Artifacts: Artifacts{
			ConfigUpdate: updateBase64,
			Signatures:   map[string]string{"Org1MSP": signatureBase64, "Org2MSP": signatureBase64},
		},
	}
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(expectedJSON, nil)
	actual, err := sc.GetProposal(transactionContext, "request-1")
	require.NoError(t, err)
	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(actualJSON))

	// Case: Internal state read error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve proposal"))
	_, err = sc.GetProposal(transactionContext, "request-1")
	require.EqualError(t, err, "error happened reading proposal with id (request-1): unable to retrieve proposal")

	// Case: Fail when the proposal does not exist
	chaincodeStub.GetStateReturns(nil, nil)
	actual, err = sc.GetProposal(transactionContext, "request-1")
	require.EqualError(t, err, "proposal not found")
	require.Nil(t, actual)

	// Case: Fail to create composite key
	chaincodeStub.CreateCompositeKeyReturns("", fmt.Errorf("failed to create composite key"))
	actual, err = sc.GetProposal(transactionContext, "request-1")
	require.EqualError(t, err, "error happend creating composite key for proposal: failed to create composite key")
	require.Nil(t, actual)
}
