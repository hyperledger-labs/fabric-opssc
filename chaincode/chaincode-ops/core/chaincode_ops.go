/*
Copyright 2017-2022 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/common/util"
)

// SmartContract provides functions for operating chaincode
type SmartContract struct {
	contractapi.Contract
}

// ChaincodePackage represents information on chaincode package of a proposed update.
type ChaincodePackage struct {
	Repository        string `json:"repository"`
	CommitID          string `json:"commitID"`
	PathToSourceFiles string `json:"pathToSourceFiles,omitempty" metadata:",optional"`
	Type              string `json:"type"`
}

// ChaincodeDefinition represents information on chaincode definition of a proposed update.
type ChaincodeDefinition struct {
	Sequence            int64  `json:"sequence"`
	InitRequired        bool   `json:"initRequired"`
	ValidationParameter string `json:"validationParameter"`
	Collections         string `json:"collections,omitempty" metadata:",optional"`
}

// ChaincodeUpdateProposal describes a new chaincode update proposal that is stored as a state in the ledger.
type ChaincodeUpdateProposal struct {
	ObjectType          string              `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID                  string              `json:"ID"`
	Creator             string              `json:"creator"`
	ChannelID           string              `json:"channelID"`
	ChaincodeName       string              `json:"chaincodeName"`
	ChaincodePackage    ChaincodePackage    `json:"chaincodePackage"`
	ChaincodeDefinition ChaincodeDefinition `json:"chaincodeDefinition"`
	Status              string              `json:"status"`
	Time                string              `json:"time"`
}

// ChaincodeUpdateProposalInput represents a request input of a new chaincode update proposal.
type ChaincodeUpdateProposalInput struct {
	ID                  string              `json:"ID"`
	ChannelID           string              `json:"channelID"`
	ChaincodeName       string              `json:"chaincodeName"`
	ChaincodePackage    ChaincodePackage    `json:"chaincodePackage"`
	ChaincodeDefinition ChaincodeDefinition `json:"chaincodeDefinition"`
}

// History describes a history of each task (e.g., vote, chaincode commit), and which is stored as a state in the ledger.
type History struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ProposalID string `json:"proposalID"`
	TaskID     string `json:"taskID"`
	OrgID      string `json:"orgID"`
	Status     string `json:"status"`
	Data       string `json:"data"`
	Time       string `json:"time"`
}

// TaskStatusUpdateRequest represents a request input for updating a task status of a proposal.
type TaskStatusUpdateRequest struct {
	ProposalID string `json:"proposalID"`
	Status     string `json:"status,omitempty" metadata:",optional"`
	Data       string `json:"data,omitempty" metadata:",optional"`
}

// HistoryQueryParams represents query parameters for getting histories from the ledger.
type HistoryQueryParams struct {
	ProposalID string `json:"proposalID,omitempty"`
	TaskID     string `json:"taskID,omitempty" metadata:",optional"`
	OrgID      string `json:"orgID,omitempty" metadata:",optional"`
}

// DeploymentEventDetail represents details of DeploymentEvent.
type DeploymentEventDetail struct {
	Proposal         ChaincodeUpdateProposal `json:"proposal"`
	OperationTargets []string                `json:"operationTargets"`
}

// VotingConfig represents voting config.
type VotingConfig struct {
	ObjectType       string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	MaxMaliciousOrgs int    `json:"maxMaliciousOrgs"`
}

// Object types
const (
	ProposalObjectType     = "proposal"
	HistoryObjectType      = "history"
	VotingConfigObjectType = "votingConfig"
)

// Chaincode event names
const (
	NewProposalEvent     = "newProposalEvent"
	NewVoteEvent         = "newVoteEvent"
	PrepareToDeployEvent = "prepareToDeployEvent"
	DeployEvent          = "deployEvent"
	CommittedEvent       = "committedEvent"
	RejectedEvent        = "rejectedEvent"
	WithdrawnEvent       = "withdrawnEvent"
)

// Task IDs
const (
	Vote        = "vote"
	Acknowledge = "acknowledge"
	Commit      = "commit"
)

// Status for Proposal
const (
	Proposed     = "proposed"
	Approved     = "approved"
	Rejected     = "rejected"
	Acknowledged = "acknowledged"
	Committed    = "committed"
	Failed       = "failed" // Unsupported
	Withdrawn    = "withdrawn"
)

// Status for Vote Tasks
const (
	Agreed    = "agreed"
	Disagreed = "disagreed"
)

// Status for Acknowledge and Commit Tasks
const (
	Success = "success"
	Failure = "failure"
)

// Criteria for moving the next task
const (
	ALL      = "all"
	MAJORITY = "majority"
)

// Const for channel-ops
const (
	ChannelOpsChaincodeNameEnv     = "CH_OPS_CC_NAME"
	DefaultChannelOpsChaincodeName = "channel-ops"
)

var (
	// ErrProposalNotFound is returned when the requested object is not found.
	ErrProposalNotFound = fmt.Errorf("proposal not found")
	// ErrProposalIDAreadyInUse is returned when the requested proposal ID is already in use.
	ErrProposalIDAreadyInUse = fmt.Errorf("proposalID already in use")
)

// RequestProposal requests a new chaincode update proposal.
//
// Arguments:
//   0: input - the request input for the chaincode update proposal
//
// Returns:
//   0: the created proposal
//   1: error
//
// Events:
//   (if the request can be approved without any other votes)
//   name: PrepareToCommitEvent(<proposalID>)
//   payload: DeploymentEventDetail
//   (else)
//   name: newProposalEvent(<proposalID>)
//   payload: the created proposal
//
func (s *SmartContract) RequestProposal(ctx contractapi.TransactionContextInterface, input ChaincodeUpdateProposalInput) (*ChaincodeUpdateProposal, error) {

	// Validate input
	// TODO: stricter input check may be desirable here although a wrong proposal is rejected when _lifecycle chaincode is executed.
	if input.ID == "" {
		return nil, fmt.Errorf("the required parameter proposal 'ID' is empty")
	}

	if input.ChannelID == "" {
		return nil, fmt.Errorf("the required parameter 'ChannelID' is empty")
	}

	if input.ChaincodeName == "" {
		return nil, fmt.Errorf("the required parameter 'ChaincodeName' is empty")
	}

	if input.ChaincodeDefinition.Sequence < 1 {
		return nil, fmt.Errorf("the parameter 'ChaincodeDefinition.Sequence' should be >= 1")
	}

	if input.ChaincodeDefinition.ValidationParameter == "" {
		return nil, fmt.Errorf("the required parameter 'ChaincodeDefinition.ValidationParameter' is empty")
	}

	url, err := url.Parse(input.ChaincodePackage.Repository)
	if err != nil || url.Scheme != "" {
		return nil, fmt.Errorf("the parameter 'ChaincodePackage.Repository' should be repository path (e.g., github.com/project_name/repository_name)")
	}

	if input.ChaincodePackage.Type == "" {
		return nil, fmt.Errorf("the required parameter 'ChaincodePackage.Type' is empty")
	}

	if input.ChaincodePackage.CommitID == "" {
		return nil, fmt.Errorf("the required parameter 'ChaincodePackage.CommitID' is empty")
	}

	// Build the proposal
	mspID, err := s.getMSPID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MSP ID: %v", err)
	}

	txTimestamp, err := getTxTimestampRFC3339(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx timestamp: %v", err)
	}

	proposal := ChaincodeUpdateProposal{
		ObjectType:          ProposalObjectType,
		ID:                  input.ID,
		Creator:             mspID,
		Time:                txTimestamp,
		Status:              Proposed,
		ChannelID:           input.ChannelID,
		ChaincodeName:       input.ChaincodeName,
		ChaincodePackage:    input.ChaincodePackage,
		ChaincodeDefinition: input.ChaincodeDefinition,
	}

	// Check whether the proposal is acceptable to the target channel
	if b, err := s.canPropose(ctx, input.ChannelID); !b {
		return nil, fmt.Errorf("proposal is not accepted by the channel. The proposal should be made to the 'application' or 'ops' channel: %v", err)
	}

	// Fail if the proposal with the ID already exists
	if p, _ := s.GetProposal(ctx, input.ID); p != nil {
		return nil, ErrProposalIDAreadyInUse
	}

	// Put the proposal to stateDB
	if err = s.putProposal(ctx, proposal); err != nil {
		return nil, fmt.Errorf("failed to put the proposal: %v", err)
	}

	// Vote for myself
	history, err := s.putHistory(ctx, proposal.ID, Vote, Agreed, "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to put the history that the org votes for: %v", err)
	}

	// If the vote from this organization alone meets the MAJORITY condition,
	// Update proposal status to "Approved" and issue PrepareToCommitEvent (the event is set in the internal function)
	votePassed, err := s.meetCriteria(ctx, *history, MAJORITY, false, proposal.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to do meetCriteria: %v", err)
	}
	if votePassed {
		if err = s.updateStatusToApproved(ctx, proposal); err != nil {
			return nil, fmt.Errorf("failed to update the status: %v", err)
		}
		return &proposal, nil
	}

	// Else issue NewProposalEvent
	proposalJSON, err := json.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("error happened unmarshalling a proposal JSON representation to struct: %v", err)
	}
	if err = ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", NewProposalEvent, proposal.ID), []byte(proposalJSON)); err != nil {
		return nil, fmt.Errorf("error happened emitting event: %v", err)
	}
	return &proposal, nil
}

// Vote votes for / against the chaincode update proposal.
// This function records the vote as a state into the ledger.
// Also, if the proposal is voted by MAJORITY, this changes the status of the proposal from proposed to approved.
//
// Arguments:
//   0: taskStatusUpdateRequest - the request input for voting for/against the chaincode update proposal
//
// Returns:
//   0: error
//
// Events:
//   (if the status is changed to approved)
//   name: PrepareToCommitEvent(<proposalID>)
//   payload: DeploymentEventDetail
//   (else)
//   name: NewVoteEvent(<proposalID>)
//   payload: nil
//
func (s *SmartContract) Vote(ctx contractapi.TransactionContextInterface, taskStatusUpdateRequest TaskStatusUpdateRequest) error {

	// Set default values
	if taskStatusUpdateRequest.Status == "" {
		taskStatusUpdateRequest.Status = Agreed
	}

	// Validate input
	if taskStatusUpdateRequest.ProposalID == "" {
		return fmt.Errorf("the required parameter 'ProposalID' is empty")
	}

	if taskStatusUpdateRequest.Status != Agreed && taskStatusUpdateRequest.Status != Disagreed {
		return fmt.Errorf("task status for vote should be %s or %s", Agreed, Disagreed)
	}

	// Get proposal from StateDB
	proposal, err := s.GetProposal(ctx, taskStatusUpdateRequest.ProposalID)
	if err != nil {
		return fmt.Errorf("failed to get the proposal: %v", err)
	}
	// If the proposal status already got changed from "Proposed", return error
	if proposal.Status != Proposed {
		return fmt.Errorf("the voting is already closed")
	}

	// Put the task status as a history to stateDB
	history, err := s.putHistory(ctx, taskStatusUpdateRequest.ProposalID, Vote, taskStatusUpdateRequest.Status, taskStatusUpdateRequest.Data, false)
	if err != nil {
		return fmt.Errorf("failed to put the history: %v", err)
	}

	// [State Transition]
	// Prerequisite Proposal Status: "Proposed"
	//
	// Conditions:
	//   - Case A: (1) voting status is "Agreed" AND (2) voted by MAJORITY
	//         -> Update proposal status to "Approved" and issue PrepareToCommitEvent (the event is set in the internal function)
	//
	//   - Case B: (1) voting status is "Disagreed" AND (2) the number of "Agreed" can not satisfy MAJORITY
	//         -> Update proposal status to "Rejected" and issue RejectedEvent (the event is set in the internal function)
	//
	//   - Case C: Others
	//         -> Not update proposal status and issue NewVoteEvent
	switch taskStatusUpdateRequest.Status {
	// Case A:
	case Agreed:
		votePassed, err := s.meetCriteria(ctx, *history, MAJORITY, false, proposal.ChannelID)
		if err != nil {
			return fmt.Errorf("failed to do meetCriteria: %v", err)
		}
		if votePassed {
			if err = s.updateStatusToApproved(ctx, *proposal); err != nil {
				return fmt.Errorf("failed to update the status: %v", err)
			}
			return nil
		}
	// Case B:
	case Disagreed:
		voteRejected, err := s.meetCriteria(ctx, *history, MAJORITY, true, proposal.ChannelID)
		if err != nil {
			return fmt.Errorf("failed to do meetCriteria: %v", err)
		}
		if voteRejected {
			if err = s.updateStatusToRejected(ctx, *proposal); err != nil {
				return fmt.Errorf("failed to update the status: %v", err)
			}
			return nil
		}
	}
	// Case C:
	if err := ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", NewVoteEvent, proposal.ID), []byte(nil)); err != nil {
		return fmt.Errorf("error happened emitting event: %v", err)
	}
	return nil
}

// WithdrawProposal withdraws the chaincode update proposal.
// This only accepts the request from the proposing organization.
// This function is only available before the decision of the proposal.
//
// Arguments:
//   0: proposalID - the ID for the chaincode update proposal
//
// Returns:
//   0: error
//
// Events:
//   name: withdrawnEvent(<proposalID>)
//   payload: nil
//
func (s *SmartContract) WithdrawProposal(ctx contractapi.TransactionContextInterface, proposalID string) error {

	// Validate input
	if proposalID == "" {
		return fmt.Errorf("the required parameter 'proposalID' is empty")
	}

	// Get proposal from StateDB
	proposal, err := s.GetProposal(ctx, proposalID)
	if err != nil {
		return ErrProposalNotFound
	}

	// If the proposal status already got changed from "Proposed", return error
	if proposal.Status != Proposed {
		return fmt.Errorf("the voting is already closed")
	}

	// If the proposal is not created by the requester, return error
	mspID, err := s.getMSPID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}
	if proposal.Creator != mspID {
		return fmt.Errorf("only the proposer (%v) can withdraw the proposal", proposal.Creator)
	}

	// Update the proposal status to withdrawn (and issue the withdrawnEvent)
	if err = s.updateStatusToWithdrawn(ctx, *proposal); err != nil {
		return fmt.Errorf("failed to update the status: %v", err)
	}
	return nil
}

// Acknowledge records the task status executed by agents for preparing the deployment based on the chaincode update proposal.
// This function records the result of the task as a state into the ledger.
// Also, if the proposal is acknowledged by ALL organizations, this changes the status of the proposal from approved to acknowledged.
//
// Arguments:
//   0: taskStatusUpdateRequest - the task status executed by agents for preparing the deployment based on the chaincode update proposal
//
// Returns:
//   0: error
//
// Events:
//   (if the status is changed to acknowledged)
//   name: DeployEvent(<proposalID>)
//   payload: DeploymentEventDetail
//
func (s *SmartContract) Acknowledge(ctx contractapi.TransactionContextInterface, taskStatusUpdateRequest TaskStatusUpdateRequest) error {

	// Set default status
	if taskStatusUpdateRequest.Status == "" {
		taskStatusUpdateRequest.Status = Success
	}
	// Validate input
	if taskStatusUpdateRequest.ProposalID == "" {
		return fmt.Errorf("the required parameter 'ProposalID' is empty")
	}
	if taskStatusUpdateRequest.Status != Success && taskStatusUpdateRequest.Status != Failure {
		return fmt.Errorf("task status for acknowledge should be %s or %s", Success, Failure)
	}

	// Get proposal from StateDB
	proposal, err := s.GetProposal(ctx, taskStatusUpdateRequest.ProposalID)
	if err != nil {
		return fmt.Errorf("failed to get the proposal: %v", err)
	}
	// If the proposal status already got changed from "Approved", return success immediately
	if proposal.Status != Approved {
		return nil
	}

	// Put the task status as a history to stateDB
	history, err := s.putHistory(ctx, taskStatusUpdateRequest.ProposalID, Acknowledge, taskStatusUpdateRequest.Status, taskStatusUpdateRequest.Data, true)
	if err != nil {
		return fmt.Errorf("failed to put the history: %v", err)
	}

	// If (1) the proposal status remains "Approved" and (2) the proposal is acknowledged by ALL orgs,
	// then update proposal status to "Acknowledged" and issue commitEvent (the event is internally set)
	isAcknowledgedByAllOrgs, err := s.meetCriteria(ctx, *history, ALL, false, proposal.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to do meetCriteria: %v", err)
	}
	if taskStatusUpdateRequest.Status == Success && isAcknowledgedByAllOrgs {
		if err = s.updateStatusToAcknowledged(ctx, *proposal); err != nil {
			return fmt.Errorf("failed to update the status: %v", err)
		}
	}
	return nil
}

// NotifyCommitResult records the task status executed by agents for commiting the deployment based on the chaincode update proposal.
// This function records the result of the task as a state into the ledger.
// Also, if the proposal is acknowledged by ALL organizations, this changes the status of the proposal from acknowledged to committed.
//
// Arguments:
//   0: taskStatusUpdateRequest - the task status executed by agents for commiting the deployment based on the chaincode update proposal
//
// Returns:
//   0: error
//
// Events:
//   (if the status is changed to acknowledged)
//   name: CommittedEvent(<proposalID>)
//   payload: nil
//
func (s *SmartContract) NotifyCommitResult(ctx contractapi.TransactionContextInterface, taskStatusUpdateRequest TaskStatusUpdateRequest) error {

	// Set default status
	if taskStatusUpdateRequest.Status == "" {
		taskStatusUpdateRequest.Status = Success
	}
	// Validate input
	if taskStatusUpdateRequest.ProposalID == "" {
		return fmt.Errorf("the required parameter 'ProposalID' is empty")
	}
	if taskStatusUpdateRequest.Status != Success && taskStatusUpdateRequest.Status != Failure {
		return fmt.Errorf("task status for commit should be %s or %s", Success, Failure)
	}

	// Get proposal from StateDB
	proposal, err := s.GetProposal(ctx, taskStatusUpdateRequest.ProposalID)
	if err != nil {
		return fmt.Errorf("failed to get the proposal: %v", err)
	}

	// TODO: Before the history registration, this function may need strict condition checking (transaction creator and/or whether to meet criteria)
	_, err = s.putHistory(ctx, taskStatusUpdateRequest.ProposalID, Commit, taskStatusUpdateRequest.Status, taskStatusUpdateRequest.Data, true)
	if err != nil {
		return fmt.Errorf("failed to put the history: %v", err)
	}

	// If the commit task status is not success, return immediately
	if taskStatusUpdateRequest.Status != Success {
		return nil
	}

	// If the proposal status already got changed from "Acknowledged", return success immediately
	if proposal.Status != Acknowledged {
		return nil
	}

	if err = s.updateStatusToCommitted(ctx, *proposal); err != nil {
		return fmt.Errorf("failed to update the status: %v", err)
	}

	return nil
}

// GetAllProposals returns the all chaincode update proposals.
//
// Arguments: none
//
// Returns:
//   0: the map of the all chaincode update proposals
//   1: error
//
func (s *SmartContract) GetAllProposals(ctx contractapi.TransactionContextInterface) (map[string]*ChaincodeUpdateProposal, error) {

	proposals := make(map[string]*ChaincodeUpdateProposal)
	proposalIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(ProposalObjectType, []string{})
	if err != nil {
		return nil, fmt.Errorf("error happened reading keys from ledger: %v", err)
	}
	defer proposalIterator.Close()

	for proposalIterator.HasNext() {
		proposalJSON, err := proposalIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error happened iterating over available proposals: %v", err)
		}
		proposal := &ChaincodeUpdateProposal{}
		if err = json.Unmarshal(proposalJSON.Value, proposal); err != nil {
			return nil, fmt.Errorf("error happened unmarshalling a proposal JSON representation to struct: %v", err)
		}
		proposals[proposalJSON.Key] = proposal
	}
	return proposals, nil
}

// SetMaxMaliciousOrgsInVotes sets number of max malicious orgs in votes.
//
// Arguments:
//   0: number - number of max malicious orgs in votes
//
// Returns:
//   0: error
//
func (s *SmartContract) SetMaxMaliciousOrgsInVotes(ctx contractapi.TransactionContextInterface, number int) error {

	// Validate arguments
	if number < 0 {
		return fmt.Errorf("number of max malicious orgs in votes should be greater than 0")
	}

	// struct to JSON
	votingConfigJSON, err := json.Marshal(VotingConfig{
		ObjectType:       VotingConfigObjectType,
		MaxMaliciousOrgs: number,
	})
	if err != nil {
		return fmt.Errorf("error happened marshalling the voting config: %v", err)
	}

	// Put votingConfig to StateDB
	err = ctx.GetStub().PutState(VotingConfigObjectType, votingConfigJSON)
	if err != nil {
		return fmt.Errorf("error happened persisting the voting config on the ledger: %v", err)
	}

	return nil
}

// UnsetMaxMaliciousOrgsInVotes unsets number of max malicious orgs in votes.
//
// Arguments: None
//
// Returns:
//   0: error
//
func (s *SmartContract) UnsetMaxMaliciousOrgsInVotes(ctx contractapi.TransactionContextInterface) error {

	// Delete votingConfig to StateDB
	err := ctx.GetStub().DelState(VotingConfigObjectType)
	if err != nil {
		return fmt.Errorf("error happened delete the voting config from the ledger: %v", err)
	}

	return nil
}

// GetVotingConfig returns the voting config.
//
// Arguments: None
//
// Returns:
//   0: the voting config (if voting config is not set, the func returns null)
//   1: error
//
func (s *SmartContract) GetVotingConfig(ctx contractapi.TransactionContextInterface) (*VotingConfig, error) {

	votingConfigJSON, err := ctx.GetStub().GetState(VotingConfigObjectType)
	if err != nil {
		return nil, fmt.Errorf("error happened reading voting config: %v", err)
	}

	if votingConfigJSON == nil {
		return nil, nil
	}

	var votingConfig VotingConfig
	err = json.Unmarshal(votingConfigJSON, &votingConfig)
	if err != nil {
		return nil, fmt.Errorf("error happened unmarshalling a voting config JSON representation to struct: %v", err)
	}
	return &votingConfig, nil
}

func buildAttributesForGetHistories(params HistoryQueryParams) []string {
	args := []string{}
	if params.ProposalID == "" {
		return args
	}
	args = append(args, params.ProposalID)

	if params.TaskID == "" {
		return args
	}
	args = append(args, params.TaskID)

	if params.OrgID == "" {
		return args
	}
	args = append(args, params.OrgID)
	return args
}

// GetHistories returns the histories with the given query parameters.
//
// Arguments:
//   0: params - the history query parameters
//
// Returns:
//   0: the map of the histories with the given query parameters
//   1: error
//
func (s *SmartContract) GetHistories(ctx contractapi.TransactionContextInterface, params HistoryQueryParams) (map[string]*History, error) {

	histories := make(map[string]*History)
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(HistoryObjectType, buildAttributesForGetHistories(params))
	if err != nil {
		return nil, fmt.Errorf("error happened reading keys from ledger: %v", err)
	}
	defer iterator.Close()

	for iterator.HasNext() {
		historyJSON, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error happened iterating over available histories: %v", err)
		}
		history := &History{}
		if err = json.Unmarshal(historyJSON.Value, history); err != nil {
			return nil, fmt.Errorf("error happened unmarshalling a history JSON representation to struct: %v", err)
		}
		histories[historyJSON.Key] = history
	}
	return histories, nil
}

// GetProposal returns the proposal with the given ID.
//
// Arguments:
//   0: params - the ID of the proposal
//
// Returns:
//   0: the proposal with the given ID
//   1: error
//
func (s *SmartContract) GetProposal(ctx contractapi.TransactionContextInterface, proposalID string) (*ChaincodeUpdateProposal, error) {

	compositeKey, err := ctx.GetStub().CreateCompositeKey(ProposalObjectType, []string{proposalID})
	if err != nil {
		return nil, fmt.Errorf("error happened creating composite key for proposal: %v", err)
	}

	proposalJSON, err := ctx.GetStub().GetState(compositeKey)
	if err != nil {
		return nil, fmt.Errorf("error happened reading proposal with id (%v): %v", proposalID, err)
	}

	if proposalJSON == nil {
		return nil, ErrProposalNotFound
	}

	var proposal ChaincodeUpdateProposal
	err = json.Unmarshal(proposalJSON, &proposal)
	if err != nil {
		return nil, fmt.Errorf("error happened unmarshalling a proposal JSON representation to struct: %v", err)
	}
	return &proposal, nil
}

// -- Internal logics

// Function to check whether to meet criteria for the proposal state transitions
func (s *SmartContract) meetCriteria(ctx contractapi.TransactionContextInterface, currentHistory History, criteria string, checkUnachivable bool, targetChannel string) (bool, error) {
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(HistoryObjectType, []string{currentHistory.ProposalID, currentHistory.TaskID})
	if err != nil {
		return false, fmt.Errorf("error happened reading keys from ledger: %v", err)
	}
	defer iterator.Close()

	channelOpsArgs := util.ToChaincodeArgs("CountOrganizationsInChannel", targetChannel)
	response := ctx.GetStub().InvokeChaincode(channelOpsCCName(), channelOpsArgs, "")
	if response.Status != shim.OK {
		return false, fmt.Errorf("failed to call count organization in channel (code: %d, message: %v)",
			response.Status, response.Message)
	}
	totalOrgNum, err := strconv.Atoi(string(response.Payload))
	if err != nil {
		return false, fmt.Errorf("failed to call count organization in channel: %v", err)
	}

	criteriaNum := totalOrgNum
	switch criteria {
	case MAJORITY:
		criteriaNum = criteriaNum/2 + 1

		votingConfig, err := s.GetVotingConfig(ctx)
		if err != nil {
			return false, err
		}
		if votingConfig != nil {
			criteriaNum = votingConfig.MaxMaliciousOrgs*2 + 1
		}
	case ALL:
		// use total org num
	default:
		return false, fmt.Errorf("invalid criteria type: %v", criteria)
	}

	if checkUnachivable {
		criteriaNum = totalOrgNum - (criteriaNum - 1)
	}

	orgs := map[string]string{}
	orgs[currentHistory.OrgID] = currentHistory.Status
	for iterator.HasNext() {
		result, err := iterator.Next()
		if err != nil {
			return false, nil
		}
		var resultHistory History
		err = json.Unmarshal(result.Value, &resultHistory)
		if err != nil {
			return false, nil
		}
		if resultHistory.Status == currentHistory.Status {
			orgs[resultHistory.OrgID] = resultHistory.Status
		}
	}

	if len(orgs) >= criteriaNum {
		return true, nil
	}
	return false, nil
}

// Functions to manage proposal status
func (s *SmartContract) updateStatusToAcknowledged(ctx contractapi.TransactionContextInterface, proposal ChaincodeUpdateProposal) error {
	proposal.Status = Acknowledged

	// Put proposal to stateDB
	if err := s.putProposal(ctx, proposal); err != nil {
		return err
	}

	// Issue CommitEvent
	// Set this org as a chaincode committer
	mspID, err := s.getMSPID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	// Create deployment event detail
	eventDetail := DeploymentEventDetail{
		OperationTargets: []string{mspID},
		Proposal:         proposal,
	}

	// struct to JSON
	eventDetailJSON, err := json.Marshal(eventDetail)
	if err != nil {
		return err
	}

	// Set Event
	if err = ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", DeployEvent, proposal.ID), eventDetailJSON); err != nil {
		return fmt.Errorf("error happened emitting event: %v", err)
	}
	return nil
}

func (s *SmartContract) updateStatusToApproved(ctx contractapi.TransactionContextInterface, proposal ChaincodeUpdateProposal) error {
	proposal.Status = Approved

	// Put proposal to stateDB
	if err := s.putProposal(ctx, proposal); err != nil {
		return err
	}

	// Issue PrepareToCommitEvent

	// -- Get organization list from channel-ops
	channelOpsArgs := util.ToChaincodeArgs("GetOrganizationsInChannel", proposal.ChannelID)
	response := ctx.GetStub().InvokeChaincode(channelOpsCCName(), channelOpsArgs, "")
	if response.Status != shim.OK {
		return fmt.Errorf("error happened querying " + channelOpsCCName() + ":" + response.Message)
	}
	oList := []string{}
	if err := json.Unmarshal(response.Payload, &oList); err != nil {
		return err
	}

	// -- Create deployment event detail
	eventDetail := DeploymentEventDetail{
		OperationTargets: oList,
		Proposal:         proposal,
	}

	// -- struct to JSON
	eventDetailJSON, err := json.Marshal(eventDetail)
	if err != nil {
		return err
	}

	// -- Set Event
	if err = ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", PrepareToDeployEvent, proposal.ID), eventDetailJSON); err != nil {
		return fmt.Errorf("error happened emitting event: %v", err)
	}
	return nil
}

func (s *SmartContract) updateStatusToRejected(ctx contractapi.TransactionContextInterface, proposal ChaincodeUpdateProposal) error {
	proposal.Status = Rejected

	// Put proposal to stateDB
	if err := s.putProposal(ctx, proposal); err != nil {
		return err
	}

	// -- Set Event
	if err := ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", RejectedEvent, proposal.ID), nil); err != nil {
		return fmt.Errorf("error happened emitting event: %v", err)
	}
	return nil
}

func (s *SmartContract) updateStatusToWithdrawn(ctx contractapi.TransactionContextInterface, proposal ChaincodeUpdateProposal) error {
	proposal.Status = Withdrawn

	// Put proposal to stateDB
	if err := s.putProposal(ctx, proposal); err != nil {
		return err
	}

	// -- Set Event
	if err := ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", WithdrawnEvent, proposal.ID), nil); err != nil {
		return fmt.Errorf("error happened emitting event: %v", err)
	}
	return nil
}

func (s *SmartContract) updateStatusToCommitted(ctx contractapi.TransactionContextInterface, proposal ChaincodeUpdateProposal) error {
	proposal.Status = Committed

	// Put proposal to stateDB
	if err := s.putProposal(ctx, proposal); err != nil {
		return err
	}

	// -- Set Event
	if err := ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", CommittedEvent, proposal.ID), []byte(nil)); err != nil {
		return fmt.Errorf("error happened emitting event: %v", err)
	}

	return nil
}

// Accessors to StateDB
func (s *SmartContract) putProposal(ctx contractapi.TransactionContextInterface, proposal ChaincodeUpdateProposal) error {
	// Create composite key
	compositeKey, err := ctx.GetStub().CreateCompositeKey(ProposalObjectType, []string{proposal.ID})
	if err != nil {
		return fmt.Errorf("error happened creating composite key for proposal: %v", err)
	}

	// struct to JSON
	proposalJSON, err := json.Marshal(proposal)
	if err != nil {
		return fmt.Errorf("error happened marshalling the new proposal: %v", err)
	}

	// Put proposal to StateDB
	err = ctx.GetStub().PutState(compositeKey, proposalJSON)
	if err != nil {
		return fmt.Errorf("error happened persisting the new proposal on the ledger: %v", err)
	}

	return nil
}

func (s *SmartContract) putHistory(ctx contractapi.TransactionContextInterface, proposalID string, taskID string, status string, data string, overwritable bool) (*History, error) {

	// Validate input
	if proposalID == "" {
		return nil, fmt.Errorf("the required parameter 'Proposal ID' is empty")
	}

	mspID, err := s.getMSPID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MSP ID: %v", err)
	}

	txTimestamp, err := getTxTimestampRFC3339(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx timestamp: %v", err)
	}

	// JSON to struct
	history := &History{
		ObjectType: HistoryObjectType,
		TaskID:     taskID,
		ProposalID: proposalID,
		OrgID:      mspID,
		Status:     status,
		Data:       data,
		Time:       txTimestamp,
	}

	// struct to JSON
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return nil, err
	}

	// Create composite key
	compositeKey, err := ctx.GetStub().CreateCompositeKey(HistoryObjectType, []string{history.ProposalID, history.TaskID, history.OrgID})
	if err != nil {
		return nil, fmt.Errorf("error happened creating composite key for history: %v", err)
	}

	// Check whether there is the state on the state DB
	if !overwritable {
		obtainedJSON, err := ctx.GetStub().GetState(compositeKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read from world state: %v", err)
		}
		if obtainedJSON != nil {
			return nil, fmt.Errorf("the state is already exists: %v", history.OrgID)
		}
	}

	// Put state
	err = ctx.GetStub().PutState(compositeKey, historyJSON)
	if err != nil {
		return nil, fmt.Errorf("error happened marshalling the history: %v", err)
	}

	return history, nil
}

func (s *SmartContract) getMSPID(ctx contractapi.TransactionContextInterface) (string, error) {
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return "", fmt.Errorf("error happened reading the transaction creator: %v", err)
	}
	return getMSPID(creator)
}

func (s *SmartContract) canPropose(ctx contractapi.TransactionContextInterface, ChannelID string) (bool, error) {
	channelOpsArgs := util.ToChaincodeArgs("GetChannelType", ChannelID)
	response := ctx.GetStub().InvokeChaincode(channelOpsCCName(), channelOpsArgs, "")
	if response.Status != shim.OK {
		return false, fmt.Errorf("failed to call get channel type (code: %d, message: %v)",
			response.Status, response.Message)
	}
	channelType := string(response.Payload)
	if channelType == "system" || channelType == "disable" {
		return false, nil
	}
	return true, nil
}

// Utils
func getMSPID(creator []byte) (string, error) {
	identity := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(creator, identity); err != nil {
		return "", fmt.Errorf("error happened unmarshalling the creator: %v", err)
	}
	return identity.Mspid, nil
}

func getTxTimestampRFC3339(ctx contractapi.TransactionContextInterface) (string, error) {
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", err
	}
	tm := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	return tm.Format(time.RFC3339), nil
}

func channelOpsCCName() string {
	if os.Getenv(ChannelOpsChaincodeNameEnv) != "" {
		return os.Getenv(ChannelOpsChaincodeNameEnv)
	}
	return DefaultChannelOpsChaincodeName
}
