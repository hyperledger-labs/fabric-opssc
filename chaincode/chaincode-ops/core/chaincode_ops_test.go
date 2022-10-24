/*
Copyright 2020-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabric-opssc/chaincode/chaincode-ops/core/mocks"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
)

var (
	org1MSP = marshalProtoOrPanic(&msp.SerializedIdentity{Mspid: "Org1MSP", IdBytes: []byte("myid")})
	org2MSP = marshalProtoOrPanic(&msp.SerializedIdentity{Mspid: "Org2MSP", IdBytes: []byte("myid")})
)

// marshalProtoOrPanic is a helper for proto marshal.
func marshalProtoOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}

	return data
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

// Dummy implementation for chaincode to chaincode
func invokeChaincode(arg1 string, arg2 [][]byte, arg3 string) peer.Response {
	funcName := string(arg2[0])
	switch funcName {
	case "GetOrganizationsInChannel":
		orgList := []string{"Org1MSP", "Org2MSP"}
		orgListJSON, err := json.Marshal(orgList)
		if err != nil {
			panic(err)
		}
		return peer.Response{
			Status:  shim.OK,
			Payload: []byte(orgListJSON),
		}
	case "CountOrganizationsInChannel":
		return peer.Response{
			Status:  shim.OK,
			Payload: []byte("2"),
		}
	case "GetChannelType":
		return peer.Response{
			Status:  shim.OK,
			Payload: []byte("application"),
		}
	}
	panic("Unexpected func name")
}

func TestRequestProposal(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: Request a proposal to update a chaincode
	chaincodeStub.GetCreatorReturns(org1MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)
	expectedProposal, input := baseProposalAndInput(formattedTS)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	actual, err := sc.RequestProposal(transactionContext, input)
	require.NoError(t, err)
	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "proposal_request-1", key)

	expectedJSON, err := json.Marshal(expectedProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))
	require.JSONEq(t, string(expectedJSON), string(actualJSON))

	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "newProposalEvent.request-1", eventName)
	require.Equal(t, string(expectedJSON), string(eventPayload))

	key, state = chaincodeStub.PutStateArgsForCall(1)
	require.Equal(t, "history_request-1_vote_Org1MSP", key)
	expectedHistory := History{
		ObjectType: HistoryObjectType,
		ProposalID: input.ID,
		OrgID:      "Org1MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	expectedJSON, err = json.Marshal(expectedHistory)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))

	// Case: the proposal can be approved without any other votes
	expectedProposal, input = baseProposalAndInput(formattedTS)
	input.ID = "request-2"
	config := VotingConfig{
		ObjectType:       VotingConfigObjectType,
		MaxMaliciousOrgs: 0,
	}
	configJSON, err := json.Marshal(config)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(configJSON, nil)

	getStateCount := chaincodeStub.GetStateCallCount()
	chaincodeStub.GetStateReturnsOnCall(getStateCount, nil, nil)          // GetProposal
	chaincodeStub.GetStateReturnsOnCall(getStateCount+1, nil, nil)        // GetHistory
	chaincodeStub.GetStateReturnsOnCall(getStateCount+2, configJSON, nil) // GetVotingConfig
	_, err = sc.RequestProposal(transactionContext, input)
	require.NoError(t, err)

	key, state = chaincodeStub.PutStateArgsForCall(3)
	require.Equal(t, "history_request-2_vote_Org1MSP", key)
	expectedHistory.ProposalID = input.ID
	expectedJSON, err = json.Marshal(expectedHistory)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))

	expectedProposal.ID = input.ID
	expectedProposal.Status = Approved
	expectedJSON, err = json.Marshal(expectedProposal)
	require.NoError(t, err)
	key, state = chaincodeStub.PutStateArgsForCall(4)
	require.Equal(t, "proposal_request-2", key)
	require.JSONEq(t, string(expectedJSON), string(state))

	// Case: Fail to request when the proposal ID is empty
	_, input = baseProposalAndInput(formattedTS)
	input.ID = ""
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "the required parameter proposal 'ID' is empty")

	// Case: Fail to request when the channelID is empty
	_, input = baseProposalAndInput(formattedTS)
	input.ChannelID = ""
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "the required parameter 'ChannelID' is empty")

	// Case: Fail to request when the ChaincodeName is empty
	_, input = baseProposalAndInput(formattedTS)
	input.ChaincodeName = ""
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "the required parameter 'ChaincodeName' is empty")

	// Case: Fail to request when the ChaincodeDefinition.Sequence is invalid
	_, input = baseProposalAndInput(formattedTS)
	input.ChaincodeDefinition.Sequence = 0
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "the parameter 'ChaincodeDefinition.Sequence' should be >= 1")

	// Case: Fail to request when the ChaincodeDefinition.ValidationParameter is empty
	_, input = baseProposalAndInput(formattedTS)
	input.ChaincodeDefinition.ValidationParameter = ""
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "the required parameter 'ChaincodeDefinition.ValidationParameter' is empty")

	// Case: Fail to request when the ChaincodePackage.Repository is invalid
	_, input = baseProposalAndInput(formattedTS)
	input.ChaincodePackage.Repository = "https://github.com/hyperledger/fabric-samples"
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "the parameter 'ChaincodePackage.Repository' should be repository path (e.g., github.com/project_name/repository_name)")

	// Case: Fail to request when the ChaincodePackage.Type is empty
	_, input = baseProposalAndInput(formattedTS)
	input.ChaincodePackage.Type = ""
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "the required parameter 'ChaincodePackage.Type' is empty")

	// Case: Fail to request when the ChaincodePackage.CommitID is invalid
	_, input = baseProposalAndInput(formattedTS)
	input.ChaincodePackage.CommitID = ""
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "the required parameter 'ChaincodePackage.CommitID' is empty")

	// Case: Fail to request when getting MSP ID is failed
	chaincodeStub.GetCreatorReturns(nil, fmt.Errorf("failed to get creator"))
	_, input = baseProposalAndInput(formattedTS)
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "failed to get MSP ID: error happened reading the transaction creator: failed to get creator")

	// Case: Fail to request when getting tx timestamp is failed
	chaincodeStub.GetCreatorReturns(org1MSP, nil)
	chaincodeStub.GetTxTimestampReturns(nil, fmt.Errorf("failed to get tx timestamp"))
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "failed to get tx timestamp: failed to get tx timestamp")

	// Case: Fail to request when the proposal is requested to an unacceptable channel
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	chaincodeStub.InvokeChaincodeReturns(peer.Response{
		Status:  shim.ERROR,
		Message: "error",
	})
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "proposal is not accepted by the channel. The proposal should be made to the 'application' or 'ops' channel: failed to call get channel type (code: 500, message: error)")

	// Case: Fail to request when the proposal ID is already in use
	chaincodeStub.InvokeChaincodeStub = invokeChaincode
	dummyState := ChaincodeUpdateProposal{}
	dummyJSON, err := json.Marshal(dummyState)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(dummyJSON, nil)
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "proposalID already in use")

	// Case: Fail to request when setEvent occurs an error
	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.SetEventReturns(fmt.Errorf("failed to set event"))
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "error happened emitting event: failed to set event")

	// Case: Fail to request when putHistory occurs an error
	cc := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(cc+2, "", fmt.Errorf("failed to create composite key"))
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "failed to put the history that the org votes for: error happened creating composite key for history: failed to create composite key")

	// Case: Fail to request when putProposal occurs an error
	chaincodeStub.CreateCompositeKeyReturns("", fmt.Errorf("failed to create composite key"))
	_, err = sc.RequestProposal(transactionContext, input)
	require.EqualError(t, err, "failed to put the proposal: error happened creating composite key for proposal: failed to create composite key")
}

func TestVote(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: Vote for the proposal and the votes do not pass the majority
	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	err = sc.Vote(transactionContext, request)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "history_request-1_vote_Org2MSP", key)
	historyOrg2 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org2MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	historyOrg2JSON, err := json.Marshal(historyOrg2)
	require.NoError(t, err)
	require.JSONEq(t, string(historyOrg2JSON), string(state))

	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "newVoteEvent.request-1", eventName)
	require.Equal(t, []byte(nil), eventPayload)

	// Case: Vote for the proposal and the votes pass the majority
	getStateCount := chaincodeStub.GetStateCallCount()
	chaincodeStub.GetStateReturnsOnCall(getStateCount, baseProposalJSON, nil)
	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	historyOrg1JSON, err := json.Marshal(historyOrg1)
	require.NoError(t, err)
	iterator = &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: historyOrg1JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	err = sc.Vote(transactionContext, request)
	require.NoError(t, err)
	key, state = chaincodeStub.PutStateArgsForCall(1)
	require.Equal(t, "history_request-1_vote_Org2MSP", key)
	historyOrg2JSON, err = json.Marshal(historyOrg2)
	require.NoError(t, err)
	require.JSONEq(t, string(historyOrg2JSON), string(state))

	key, state = chaincodeStub.PutStateArgsForCall(2)
	baseProposal.Status = Approved
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(baseProposalJSON), string(state))

	expectedEventDetail := DeploymentEventDetail{
		OperationTargets: []string{"Org1MSP", "Org2MSP"},
		Proposal:         baseProposal,
	}
	expectedEventDetailJSON, err := json.Marshal(expectedEventDetail)
	require.NoError(t, err)
	eventName, eventPayload = chaincodeStub.SetEventArgsForCall(1)
	require.Equal(t, "prepareToDeployEvent.request-1", eventName)
	require.JSONEq(t, string(expectedEventDetailJSON), string(eventPayload))

	// Case: Fail to vote for the proposal and the the status is already approved
	baseProposal, _ = baseProposalAndInput(formattedTS)
	baseProposal.Status = Approved
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	iterator = &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "the voting is already closed")
}

func TestVoteWhenRejected(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: The proposal is rejected
	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
		Status:     Disagreed,
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)

	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)
	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil) // No history for org2

	historyOrg1JSON, err := json.Marshal(historyOrg1)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: historyOrg1JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	err = sc.Vote(transactionContext, request)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "history_request-1_vote_Org2MSP", key)
	historyOrg2 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org2MSP",
		TaskID:     Vote,
		Status:     Disagreed,
		Time:       formattedTS,
	}
	historyOrg2JSON, err := json.Marshal(historyOrg2)
	require.NoError(t, err)
	require.JSONEq(t, string(historyOrg2JSON), string(state))
	key, state = chaincodeStub.PutStateArgsForCall(1)
	baseProposal.Status = Rejected
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(baseProposalJSON), string(state))

	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "rejectedEvent.request-1", eventName)
	require.Equal(t, []byte(nil), eventPayload)
}

func TestVoteWhenTryingUpdate(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: Fail to update the vote from the same org
	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org1MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)
	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	historyOrg1JSON, err := json.Marshal(historyOrg1)
	chaincodeStub.GetStateReturnsOnCall(1, historyOrg1JSON, nil)

	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "failed to put the history: the state is already exists: Org1MSP")
}

func TestVoteWithInvalidInputParameters(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: Fail to vote when the proposalID is empty
	request := TaskStatusUpdateRequest{}
	err := sc.Vote(transactionContext, request)
	require.EqualError(t, err, "the required parameter 'ProposalID' is empty")

	// Case: Fail to vote when the status is invalid
	request = TaskStatusUpdateRequest{
		ProposalID: "request-1",
		Status:     "Invalid",
	}
	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "task status for vote should be agreed or disagreed")
}

func TestVoteWhenPutProposalFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	// Prepare a state for GetProposal()
	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	// Prepare a state for getting history (there is no state for the voting from Org2)
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil)

	// Prepare states for GetHistories()
	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	historyOrg1JSON, err := json.Marshal(historyOrg1)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: historyOrg1JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	// Case: Fail to vote when putProposal occurs an error
	request = TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount+2, "", fmt.Errorf("failed to create composite key"))
	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "failed to update the status: error happened creating composite key for proposal: failed to create composite key")
}

func TestVoteWhenSetEventErrorOccurs(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	// Prepare a state for GetProposal()
	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	// Prepare a state for getting history (there is no state for the voting from Org2)
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil)

	// Prepare states for GetHistories()
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	// Case: Fail to vote when setEvent occurs an error
	chaincodeStub.SetEventReturns(fmt.Errorf("failed to set event"))
	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "error happened emitting event: failed to set event")
}

func TestVoteWhenPutHistoryFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	// Case: Fail to vote when putHistory occurs an error
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount+1, "", fmt.Errorf("failed to create composite key"))
	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "failed to put the history: error happened creating composite key for history: failed to create composite key")

	// Case: Internal state read error to check whether overwrite or not
	getStateCount := chaincodeStub.GetStateCallCount()
	chaincodeStub.GetStateReturnsOnCall(getStateCount, baseProposalJSON, nil)                           // proposal
	chaincodeStub.GetStateReturnsOnCall(getStateCount+1, nil, fmt.Errorf("unable to retrieve history")) // history
	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "failed to put the history: failed to read from world state: unable to retrieve history")
}

func TestVoteWhenGetProposalFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	// Case: Fail to vote when getProposal occurs an error
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount, "", fmt.Errorf("failed to create composite key"))
	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "failed to get the proposal: error happened creating composite key for proposal: failed to create composite key")
}

func TestVoteWhenChaincodeToChaincodeCallFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeReturns(peer.Response{
		Status:  shim.ERROR,
		Message: "error",
	})

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	// Prepare a state for GetProposal()
	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	// Prepare a state for getting history (there is no state for the voting from Org2)
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil)

	// Prepare states for GetHistories()
	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	historyOrg1JSON, err := json.Marshal(historyOrg1)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: historyOrg1JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	// Case: Fail to vote when chaincode to chaincode fails
	err = sc.Vote(transactionContext, request)
	require.EqualError(t, err, "failed to do meetCriteria: failed to call count organization in channel (code: 500, message: error)")
}

func TestAcknowledge(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: acknowledge for the proposal and the proposal is not acknowledged by ALL organizations
	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Approved
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	err = sc.Acknowledge(transactionContext, request)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "history_request-1_acknowledge_Org2MSP", key)
	historyOrg2 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org2MSP",
		TaskID:     Acknowledge,
		Status:     Success,
		Time:       formattedTS,
	}
	historyOrg2JSON, err := json.Marshal(historyOrg2)
	require.NoError(t, err)
	require.JSONEq(t, string(historyOrg2JSON), string(state))

	setEventCallCount := chaincodeStub.SetEventCallCount()
	require.Equal(t, 0, setEventCallCount)

	// Case: acknowledge for the proposal and the proposal is acknowledged by ALL organizations
	chaincodeStub.GetStateReturnsOnCall(1, baseProposalJSON, nil)
	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Acknowledge,
		Status:     Success,
		Time:       formattedTS,
	}
	historyOrg1JSON, err := json.Marshal(historyOrg1)
	iterator = &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: historyOrg1JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	err = sc.Acknowledge(transactionContext, request)
	require.NoError(t, err)
	key, state = chaincodeStub.PutStateArgsForCall(1)
	require.Equal(t, "history_request-1_acknowledge_Org2MSP", key)
	historyOrg2JSON, err = json.Marshal(historyOrg2)
	require.NoError(t, err)
	require.JSONEq(t, string(historyOrg2JSON), string(state))

	key, state = chaincodeStub.PutStateArgsForCall(2)
	baseProposal.Status = Acknowledged
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(baseProposalJSON), string(state))

	expectedEventDetail := DeploymentEventDetail{
		OperationTargets: []string{"Org2MSP"},
		Proposal:         baseProposal,
	}
	expectedEventDetailJSON, err := json.Marshal(expectedEventDetail)
	require.NoError(t, err)
	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "deployEvent.request-1", eventName)
	require.JSONEq(t, string(expectedEventDetailJSON), string(eventPayload))

	// Case: Acknowledge for the proposal and the the status is already acknowledged
	baseProposal, _ = baseProposalAndInput(formattedTS)
	baseProposal.Status = Acknowledged
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	iterator = &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	err = sc.Acknowledge(transactionContext, request)
	require.NoError(t, err)
}

func TestAcknowledgeWithInvalidInputParameters(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: Fail to acknowledge when the proposalID is empty
	request := TaskStatusUpdateRequest{}
	err := sc.Acknowledge(transactionContext, request)
	require.EqualError(t, err, "the required parameter 'ProposalID' is empty")

	// Case: Fail to acknowledge when the status is invalid
	request = TaskStatusUpdateRequest{
		ProposalID: "request-1",
		Status:     "Invalid",
	}
	err = sc.Acknowledge(transactionContext, request)
	require.EqualError(t, err, "task status for acknowledge should be success or failure")
}

func TestAcknowledgeWhenPutProposalFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Approved
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Acknowledge,
		Status:     Success,
		Time:       formattedTS,
	}
	historyOrg1JSON, err := json.Marshal(historyOrg1)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: historyOrg1JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	// Case: Fail to acknowledge when putProposal occurs an error
	request = TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount+2, "", fmt.Errorf("failed to create composite key"))
	err = sc.Acknowledge(transactionContext, request)
	require.EqualError(t, err, "failed to update the status: error happened creating composite key for proposal: failed to create composite key")
}

func TestAcknowledgeWhenPutHistoryFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Approved
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	// Case: Fail to acknowledge when putHistory occurs an error
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount+1, "", fmt.Errorf("failed to create composite key"))
	err = sc.Acknowledge(transactionContext, request)
	require.EqualError(t, err, "failed to put the history: error happened creating composite key for history: failed to create composite key")
}

func TestAcknowledgeWhenGetProposalFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Approved
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	// Case: Fail to acknowledge when getProposal occurs an error
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount, "", fmt.Errorf("failed to create composite key"))
	err = sc.Acknowledge(transactionContext, request)
	require.EqualError(t, err, "failed to get the proposal: error happened creating composite key for proposal: failed to create composite key")
}

func TestAcknowledgeWhenChaincodeToChaincodeCallFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeReturns(peer.Response{
		Status:  shim.ERROR,
		Message: "error",
	})

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Approved
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Acknowledge,
		Status:     Success,
		Time:       formattedTS,
	}
	historyOrg1JSON, err := json.Marshal(historyOrg1)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: historyOrg1JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	// Case: Fail to acknowledge when chaincode to chaincode fails
	err = sc.Acknowledge(transactionContext, request)
	require.EqualError(t, err, "failed to do meetCriteria: failed to call count organization in channel (code: 500, message: error)")
}

func TestNotifyCommitResult(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: notify commit for the proposal
	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Acknowledged
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	err = sc.NotifyCommitResult(transactionContext, request)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "history_request-1_commit_Org2MSP", key)
	historyOrg2 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org2MSP",
		TaskID:     Commit,
		Status:     Success,
		Time:       formattedTS,
	}
	historyOrg2JSON, err := json.Marshal(historyOrg2)
	require.NoError(t, err)
	require.JSONEq(t, string(historyOrg2JSON), string(state))

	key, state = chaincodeStub.PutStateArgsForCall(1)
	baseProposal.Status = Committed
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(baseProposalJSON), string(state))

	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "committedEvent.request-1", eventName)
	require.Equal(t, []byte(nil), eventPayload)

	// Case: Notify commit for the proposal and the the status is already committed
	baseProposal, _ = baseProposalAndInput(formattedTS)
	baseProposal.Status = Committed
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	iterator = &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	err = sc.NotifyCommitResult(transactionContext, request)
	require.NoError(t, err)
}

func TestNotifyCommitResultWhenTaskIsFailure(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: notify commit for the proposal
	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
		Status:     Failure,
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Acknowledged
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	err = sc.NotifyCommitResult(transactionContext, request)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "history_request-1_commit_Org2MSP", key)
	historyOrg2 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org2MSP",
		TaskID:     Commit,
		Status:     Failure,
		Time:       formattedTS,
	}
	historyOrg2JSON, err := json.Marshal(historyOrg2)
	require.NoError(t, err)
	require.JSONEq(t, string(historyOrg2JSON), string(state))

	setEventCount := chaincodeStub.SetEventCallCount()
	require.Equal(t, 0, setEventCount)
}

func TestNotifyCommitWithInvalidInputParameters(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: Fail to notify commit when the proposalID is empty
	request := TaskStatusUpdateRequest{}
	err := sc.NotifyCommitResult(transactionContext, request)
	require.EqualError(t, err, "the required parameter 'ProposalID' is empty")

	// Case: Fail to notify commit when the status is invalid
	request = TaskStatusUpdateRequest{
		ProposalID: "request-1",
		Status:     "Invalid",
	}
	err = sc.NotifyCommitResult(transactionContext, request)
	require.EqualError(t, err, "task status for commit should be success or failure")
}

func TestNotifyCommitResultWhenPutProposalFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Acknowledged
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	historyOrg1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Commit,
		Status:     Success,
		Time:       formattedTS,
	}
	historyOrg1JSON, err := json.Marshal(historyOrg1)
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: historyOrg1JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)

	// Case: Fail to notify commit when putProposal occurs an error
	request = TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount+2, "", fmt.Errorf("failed to create composite key"))
	err = sc.NotifyCommitResult(transactionContext, request)
	require.EqualError(t, err, "failed to update the status: error happened creating composite key for proposal: failed to create composite key")
}

func TestNotifyCommitResultWhenPutHistoryFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Acknowledged
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	// Case: Fail to notify commit when putHistory occurs an error
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount+1, "", fmt.Errorf("failed to create composite key"))
	err = sc.NotifyCommitResult(transactionContext, request)
	require.EqualError(t, err, "failed to put the history: error happened creating composite key for history: failed to create composite key")
}

func TestNotifyCommitResultWhenGetProposalFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	request := TaskStatusUpdateRequest{
		ProposalID: "request-1",
	}
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposal.Status = Acknowledged
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	// Case: Fail to notify commit when getProposal occurs an error
	createComposeKeyCount := chaincodeStub.CreateCompositeKeyCallCount()
	chaincodeStub.CreateCompositeKeyReturnsOnCall(createComposeKeyCount, "", fmt.Errorf("failed to create composite key"))
	err = sc.NotifyCommitResult(transactionContext, request)
	require.EqualError(t, err, "failed to get the proposal: error happened creating composite key for proposal: failed to create composite key")
}

func TestWithdrawProposal(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: withdraw the proposal
	chaincodeStub.GetCreatorReturns(org1MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, baseProposalJSON, nil)

	err = sc.WithdrawProposal(transactionContext, "request-1")
	require.NoError(t, err)

	_, state := chaincodeStub.PutStateArgsForCall(0)
	baseProposal.Status = Withdrawn
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	require.JSONEq(t, string(baseProposalJSON), string(state))

	eventName, eventPayload := chaincodeStub.SetEventArgsForCall(0)
	require.Equal(t, "withdrawnEvent.request-1", eventName)
	require.Equal(t, []byte(nil), eventPayload)
}

func TestWithdrawProposalFails(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	chaincodeStub.InvokeChaincodeStub = invokeChaincode

	sc := SmartContract{}

	// Case: Failure due to invalid parameters
	err := sc.WithdrawProposal(transactionContext, "")
	require.EqualError(t, err, "the required parameter 'proposalID' is empty")

	// Case: Failure that the proposal is not found
	err = sc.WithdrawProposal(transactionContext, "request-1")
	require.EqualError(t, err, "proposal not found")

	// Case: Failure due to the request of anyone other than the proposer
	chaincodeStub.GetCreatorReturns(org2MSP, nil)
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	baseProposal, _ := baseProposalAndInput(formattedTS)
	baseProposalJSON, err := json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)

	err = sc.WithdrawProposal(transactionContext, "request-1")
	require.EqualError(t, err, "only the proposer (Org1MSP) can withdraw the proposal")

	// Case: Failure due to the voting is closed
	chaincodeStub.GetCreatorReturns(org1MSP, nil)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	baseProposal, _ = baseProposalAndInput(formattedTS)
	baseProposal.Status = Acknowledged
	baseProposalJSON, err = json.Marshal(baseProposal)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(baseProposalJSON, nil)
	err = sc.WithdrawProposal(transactionContext, "request-1")
	require.EqualError(t, err, "the voting is already closed")
}

func TestGetAllProposals(t *testing.T) {

	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	sc := &SmartContract{}
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	// Case: Get 2 proposals
	proposal1, _ := baseProposalAndInput(formattedTS)
	proposal1JSON, err := json.Marshal(proposal1)
	require.NoError(t, err)

	proposal2, _ := baseProposalAndInput(formattedTS)
	proposal2.ID = "request-2"
	proposal2JSON, err := json.Marshal(proposal2)
	require.NoError(t, err)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, true)
	iterator.HasNextReturnsOnCall(2, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Key: "proposal_request-1", Value: proposal1JSON}, nil)
	iterator.NextReturnsOnCall(1, &queryresult.KV{Key: "proposal_request-2", Value: proposal2JSON}, nil)

	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	proposals, err := sc.GetAllProposals(transactionContext)
	require.NoError(t, err)
	require.Equal(t, map[string]*ChaincodeUpdateProposal{"proposal_request-1": &proposal1, "proposal_request-2": &proposal2}, proposals)

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
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	// Case: Get the proposal
	expected, _ := baseProposalAndInput(formattedTS)
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
	require.EqualError(t, err, "error happened creating composite key for proposal: failed to create composite key")
	require.Nil(t, actual)
}

func TestGetHistories(t *testing.T) {

	chaincodeStub := &mocks.ChaincodeStub{}
	chaincodeStub.CreateCompositeKeyStub = createComposeKey
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	sc := &SmartContract{}
	timestamp := ptypes.TimestampNow()
	chaincodeStub.GetTxTimestampReturns(timestamp, nil)
	ts := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	formattedTS := ts.Format(time.RFC3339)

	// Case: Get histories

	history1 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-1",
		OrgID:      "Org1MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	history1JSON, err := json.Marshal(history1)

	history2 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-2",
		OrgID:      "Org1MSP",
		TaskID:     Vote,
		Status:     Agreed,
		Time:       formattedTS,
	}
	history2JSON, err := json.Marshal(history2)

	history3 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-2",
		OrgID:      "Org1MSP",
		TaskID:     Acknowledge,
		Status:     Success,
		Time:       formattedTS,
	}
	history3JSON, err := json.Marshal(history3)

	history4 := History{
		ObjectType: HistoryObjectType,
		ProposalID: "request-2",
		OrgID:      "Org2MSP",
		TaskID:     Acknowledge,
		Status:     Success,
		Time:       formattedTS,
	}
	history4JSON, err := json.Marshal(history4)

	// Case: Get all histories
	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, true)
	iterator.HasNextReturnsOnCall(2, true)
	iterator.HasNextReturnsOnCall(3, true)
	iterator.HasNextReturnsOnCall(4, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Key: "history_request-1_vote_Org1MSP", Value: history1JSON}, nil)
	iterator.NextReturnsOnCall(1, &queryresult.KV{Key: "history_request-2_vote_Org1MSP", Value: history2JSON}, nil)
	iterator.NextReturnsOnCall(2, &queryresult.KV{Key: "history_request-2_acknowledge_Org1MSP", Value: history3JSON}, nil)
	iterator.NextReturnsOnCall(3, &queryresult.KV{Key: "history_request-2_acknowledge_Org2MSP", Value: history4JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	params := HistoryQueryParams{}
	histories, err := sc.GetHistories(transactionContext, params)
	require.NoError(t, err)
	require.Equal(t, map[string]*History{"history_request-1_vote_Org1MSP": &history1, "history_request-2_vote_Org1MSP": &history2,
		"history_request-2_acknowledge_Org1MSP": &history3, "history_request-2_acknowledge_Org2MSP": &history4}, histories)

	// Case: Get histories with the given proposalID
	iterator = &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, true)
	iterator.HasNextReturnsOnCall(2, true)
	iterator.HasNextReturnsOnCall(3, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Key: "history_request-2_vote_Org1MSP", Value: history2JSON}, nil)
	iterator.NextReturnsOnCall(1, &queryresult.KV{Key: "history_request-2_acknowledge_Org1MSP", Value: history3JSON}, nil)
	iterator.NextReturnsOnCall(2, &queryresult.KV{Key: "history_request-2_acknowledge_Org2MSP", Value: history4JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	params = HistoryQueryParams{
		ProposalID: "request-2",
	}
	histories, err = sc.GetHistories(transactionContext, params)
	require.NoError(t, err)
	require.Equal(t, map[string]*History{"history_request-2_vote_Org1MSP": &history2,
		"history_request-2_acknowledge_Org1MSP": &history3, "history_request-2_acknowledge_Org2MSP": &history4}, histories)

	// Case: Get histories with the given proposalID and taskID
	iterator = &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, true)
	iterator.HasNextReturnsOnCall(2, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Key: "history_request-2_acknowledge_Org1MSP", Value: history3JSON}, nil)
	iterator.NextReturnsOnCall(1, &queryresult.KV{Key: "history_request-2_acknowledge_Org2MSP", Value: history4JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	params = HistoryQueryParams{
		ProposalID: "request-2",
		TaskID:     Acknowledge,
	}
	histories, err = sc.GetHistories(transactionContext, params)
	require.NoError(t, err)
	require.Equal(t, map[string]*History{"history_request-2_acknowledge_Org1MSP": &history3, "history_request-2_acknowledge_Org2MSP": &history4}, histories)

	// Case: Get histories with the given proposalID and taskID and orgID
	iterator = &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Key: "history_request-2_acknowledge_Org2MSP", Value: history4JSON}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	params = HistoryQueryParams{
		ProposalID: "request-2",
		TaskID:     Acknowledge,
		OrgID:      "Org2MSP",
	}
	histories, err = sc.GetHistories(transactionContext, params)
	require.NoError(t, err)
	require.Equal(t, map[string]*History{"history_request-2_acknowledge_Org2MSP": &history4}, histories)

	// Case: Fail to get histories when failed retrieving next item
	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	histories, err = sc.GetHistories(transactionContext, params)
	require.EqualError(t, err, "error happened iterating over available histories: failed retrieving next item")
	require.Nil(t, histories)

	// Case: Fail to get histories when failed retrieving histories
	chaincodeStub.GetStateByPartialCompositeKeyReturns(nil, fmt.Errorf("failed retrieving all histories"))
	histories, err = sc.GetHistories(transactionContext, params)
	require.EqualError(t, err, "error happened reading keys from ledger: failed retrieving all histories")
	require.Nil(t, histories)
}

func baseProposalAndInput(formattedTimeStamp string) (ChaincodeUpdateProposal, ChaincodeUpdateProposalInput) {
	proposal := ChaincodeUpdateProposal{
		ObjectType:    ProposalObjectType,
		ID:            "request-1",
		Creator:       "Org1MSP",
		ChannelID:     "mychannel",
		ChaincodeName: "basic",
		ChaincodePackage: ChaincodePackage{
			Repository:        "github.com/hyperledger/fabric-samples",
			CommitID:          "main",
			PathToSourceFiles: "asset-transfer-basic/chaincode-go",
			Type:              "golang",
		},
		ChaincodeDefinition: ChaincodeDefinition{
			Sequence:            1,
			InitRequired:        false,
			ValidationParameter: "L0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ=",
		},
		Status: Proposed,
		Time:   formattedTimeStamp,
	}

	proposalInput := ChaincodeUpdateProposalInput{
		ID:            "request-1",
		ChannelID:     "mychannel",
		ChaincodeName: "basic",
		ChaincodePackage: ChaincodePackage{
			Repository:        "github.com/hyperledger/fabric-samples",
			CommitID:          "main",
			PathToSourceFiles: "asset-transfer-basic/chaincode-go",
			Type:              "golang",
		},
		ChaincodeDefinition: ChaincodeDefinition{
			Sequence:            1,
			InitRequired:        false,
			ValidationParameter: "L0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ=",
		},
	}

	return proposal, proposalInput
}

func TestSetMaxMaliciousOrgsInVotes(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	sc := SmartContract{}

	// Case: Set voting config
	expectedConfig := VotingConfig{
		ObjectType:       VotingConfigObjectType,
		MaxMaliciousOrgs: 2,
	}
	expectedJSON, err := json.Marshal(expectedConfig)
	require.NoError(t, err)

	err = sc.SetMaxMaliciousOrgsInVotes(transactionContext, 2)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "votingConfig", key)
	require.JSONEq(t, string(expectedJSON), string(state))

	// Case: Fail to set voting config when the number is invalid
	err = sc.SetMaxMaliciousOrgsInVotes(transactionContext, -1)
	require.EqualError(t, err, "number of max malicious orgs in votes should be greater than 0")
}

func TestUnsetMaxMaliciousOrgsInVotes(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	sc := SmartContract{}

	// Case: UnSet voting config
	err := sc.UnsetMaxMaliciousOrgsInVotes(transactionContext)
	require.NoError(t, err)

	// Case: Fail when an internal error occurs
	chaincodeStub.DelStateReturns(fmt.Errorf("fail to delete voting config"))
	err = sc.UnsetMaxMaliciousOrgsInVotes(transactionContext)
	require.EqualError(t, err, "error happened delete the voting config from the ledger: fail to delete voting config")
}

func TestGetVotingConfig(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	sc := &SmartContract{}

	// Case: Get null
	chaincodeStub.GetStateReturns(nil, nil)
	actual, err := sc.GetVotingConfig(transactionContext)
	require.NoError(t, err)
	require.Nil(t, actual)

	// Case: Get the voting config
	expected := VotingConfig{
		ObjectType:       VotingConfigObjectType,
		MaxMaliciousOrgs: 0,
	}
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(expectedJSON, nil)
	actual, err = sc.GetVotingConfig(transactionContext)
	require.NoError(t, err)
	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(actualJSON))

	// Case: Internal state read error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve voting config"))
	_, err = sc.GetVotingConfig(transactionContext)
	require.EqualError(t, err, "error happened reading voting config: unable to retrieve voting config")
}
