/*
Copyright 2009-2019 SAP SE or an SAP affiliate company. All Rights Reserved.

Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// Functionalities to communicate channel updates and signatures between different channel members.
// This code is based on the implementation of CMCC (Consortium Management Chaincode) in Fabric Interop Working Group.
// (Refer: https://wiki.hyperledger.org/display/fabric/Fabric+Interop+Working+Group)

// Proposal gathers all information of a proposed update, including all added signatures.
type Proposal struct {
	//docType is used to distinguish the various types of objects in state database
	ObjectType string `json:"docType"`

	// ID is the ID of the proposal
	ID string `json:"ID"`

	// ChannelID is the channel ID targeted by the proposal
	ChannelID string `json:"channelID"`

	// Description describes the proposal
	Description string `json:"description,omitempty" metadata:",optional"`

	// Creator describes the msp ID of the proposal creator
	Creator string `json:"creator"`

	// Action describes the action type of the proposed update ('update' or 'create')
	Action string `json:"action"`

	// Status is the status of the proposal
	Status string `json:"status"`

	// OpsProfile is the ops profile used for the input of configtx-cli to create the ConfigUpdate.
	// This is the information corresponding to Artifacts.ConfigUpdate.
	OpsProfile interface{} `json:"opsProfile"`

	// Artifacts contains the artifacts for the channel update proposal
	Artifacts Artifacts `json:"artifacts"`
}

// Artifacts contains artifacts for a channel update proposal
type Artifacts struct {
	// ConfigUpdate contains the base64 string representation of the common.ConfigUpdate
	ConfigUpdate string `json:"configUpdate"`

	// Signatures contains a map of signatures: mspID -> base64 string representation of common.ConfigSignature
	Signatures map[string]string `json:"signatures,omitempty" metadata:",optional"`
}

// ProposalInput represents a request input of a new proposal.
type ProposalInput struct {
	ID           string      `json:"ID"`
	ChannelID    string      `json:"channelID"`
	Description  string      `json:"description,omitempty" metadata:",optional"`
	Action       string      `json:"action,omitempty" metadata:",optional"`
	OpsProfile   interface{} `json:"opsProfile"`
	ConfigUpdate string      `json:"configUpdate"`
	Signature    string      `json:"signature"`
}

// Object types
const (
	ProposalObjectType = "proposal"
)

// Status for Proposal
const (
	Proposed  = "proposed"
	Approved  = "approved"
	Committed = "committed"
)

// Criteria for moving the next task
const (
	ALL      = "all"
	MAJORITY = "majority"
)

// EventDetail represents details of chaincode events which is issued in the chaincode.
type EventDetail struct {
	ProposalID       string   `json:"proposalID"`
	OperationTargets []string `json:"operationTargets"`
}

// Chaincode event names
const (
	NewProposalEvent         = "newProposalEvent"
	DeleteProposalEvent      = "deleteProposalEvent"
	NewVoteEvent             = "newVoteEvent"
	ReadyToUpdateConfigEvent = "readyToUpdateConfigEvent"
	UpdateConfigEvent        = "updateConfigEvent"
)

// Action types for channel operation
const (
	UpdateAction   = "update"
	CreationAction = "create"
)

var (
	// ErrProposalNotFound is returned when the requested object is not found.
	ErrProposalNotFound = fmt.Errorf("proposal not found")
	// ErrProposalIDAreadyInUse is returned when the requested proposal ID is already in use.
	ErrProposalIDAreadyInUse = fmt.Errorf("proposalID already in use")
	// ErrInconsistentChannelID is returned when the channel ID is inconsistent with the ID in the artifacts.
	ErrInconsistentChannelID = fmt.Errorf("channel ID is inconsistent with the ID in the artifacts")
)

// RequestProposal requests a new channel update proposal.
//
// Arguments:
//   0: input - the request input for the channel update proposal
//
// Returns:
//   0: the created proposal ID
//   1: error
//
// Events:
//   name: newProposalEvent(<proposalID>)
//   payload: the created proposal ID
//
func (s *SmartContract) RequestProposal(ctx contractapi.TransactionContextInterface, input ProposalInput) (string, error) {

	proposalID := input.ID
	action := input.Action
	if action == "" {
		action = UpdateAction
	}
	if action != UpdateAction && action != CreationAction {
		return "", fmt.Errorf("incorrect operation type - expecting %v or %v", UpdateAction, CreationAction)
	}

	if err := validateConfigUpdate(input); err != nil {
		return "", err
	}

	if err := validateConfigSignature(input.Signature); err != nil {
		return "", err
	}

	if p, _ := s.GetProposal(ctx, proposalID); p != nil {
		return "", ErrProposalIDAreadyInUse
	}

	mspID, err := s.getMSPID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get MSP ID: %v", err)
	}

	signatures := make(map[string]string)
	signatures[mspID] = input.Signature

	proposal := &Proposal{
		ObjectType:  ProposalObjectType,
		ID:          input.ID,
		ChannelID:   input.ChannelID,
		Description: input.Description,
		Creator:     mspID,
		Action:      action,
		Status:      Proposed,
		OpsProfile:  input.OpsProfile,
		Artifacts: Artifacts{
			ConfigUpdate: input.ConfigUpdate,
			Signatures:   signatures,
		},
	}

	if err = s.putProposal(ctx, proposal); err != nil {
		return "", fmt.Errorf("failed to put the proposal: %v", err)
	}

	if err = ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", NewProposalEvent, proposalID), []byte(proposalID)); err != nil {
		return "", fmt.Errorf("error happened emitting event: %v", err)
	}
	return proposalID, nil
}

// Vote votes for the channel update proposal.
// This function records the vote as a state into the ledger.
// Also, if the proposal is voted by MAJORITY, this changes the status of the proposal from proposed to approved.
//
// Arguments:
//   0: proposalID - the ID for voting for the channel update proposal
//   1: signature - the base64 string representation of common.ConfigSignature for signing by the creator to the ConfigUpdate
//
// Returns:
//   0: error
//
// Events:
//   (if the status is changed to approved)
//   name: ReadyToUpdateConfigEvent(<proposalID>)
//   payload: EventDetail
//   (else)
//   name: NewVoteEvent(<proposalID>)
//   payload: proposalID
//
func (s *SmartContract) Vote(ctx contractapi.TransactionContextInterface, proposalID, signature string) error {

	if err := validateConfigSignature(signature); err != nil {
		return err
	}

	mspID, err := s.getMSPID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	// fetch and update the state of the proposal
	proposal, err := s.GetProposal(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %v", err)
	}
	if proposal.Artifacts.Signatures == nil {
		proposal.Artifacts.Signatures = make(map[string]string)
	}
	proposal.Artifacts.Signatures[mspID] = signature

	// NewVoteEvent is the default event for when the proposal status does not change.
	eventName := fmt.Sprintf("%s.%s", NewVoteEvent, proposalID)
	eventPayload := []byte(proposalID)

	// If votes meet the criteria, it changes the proposal status to "approved" and sets ReadyToUpdateConfigEvent.
	if proposal.Status == Proposed {
		var satisfied = false
		if satisfied, err = s.passedByVoting(ctx, proposal); err != nil {
			return fmt.Errorf("fail to check whether the votes passed: %v", err)
		}
		if satisfied {
			proposal.Status = Approved

			eventDetail := EventDetail{
				ProposalID:       proposalID,
				OperationTargets: []string{mspID},
			}
			// struct to JSON
			eventDetailJSON, err := json.Marshal(eventDetail)
			if err != nil {
				return fmt.Errorf("error happened creating event detail: %v", err)
			}
			eventName = fmt.Sprintf("%s.%s", ReadyToUpdateConfigEvent, proposalID)
			eventPayload = []byte(eventDetailJSON)
		}
	}

	// Set event on the response of the transaction
	if err = ctx.GetStub().SetEvent(eventName, eventPayload); err != nil {
		return fmt.Errorf("error happened emitting event: %v", err)
	}

	// store the updated proposal
	if err = s.putProposal(ctx, proposal); err != nil {
		return fmt.Errorf("failed to put the proposal: %v", err)
	}
	return nil
}

// NotifyCommitResult records the result of the commit for the channel update proposal.
// This function records the vote as a state into the ledger.
// Also, this changes the status of the proposal from approved to committed.
// If the status is changed, this updates channel info in the states in the ledger based on the committed channel update.
//
// Arguments:
//   0: proposalID - the ID for voting for the channel update proposal
//   1: signature - the base64 string representation of common.ConfigSignature for signing by the creator to the ConfigUpdate
//
// Returns:
//   0: error
//
// Events:
//   (if the channel info is updated)
//   name: UpdateConfigEvent(<proposalID>)
//   payload: proposalID
//
func (s *SmartContract) NotifyCommitResult(ctx contractapi.TransactionContextInterface, proposalID string) error {

	proposal, err := s.GetProposal(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %v", err)
	}

	// Execute state transition
	if proposal.Status != Approved {
		return fmt.Errorf("proposal is not yet approved or already committed")
	}
	proposal.Status = Committed

	// Store the updated proposal
	if err = s.putProposal(ctx, proposal); err != nil {
		return fmt.Errorf("failed to put proposal: %v", err)
	}

	// Update organizations in the updated channel
	update, err := base64.StdEncoding.DecodeString(proposal.Artifacts.ConfigUpdate)
	if err != nil {
		return fmt.Errorf("error happened decoding the configUpdate base64 string: %v", err)
	}
	unmarshaledUpdate := common.ConfigUpdate{}
	if err := proto.Unmarshal(update, &unmarshaledUpdate); err != nil {
		return fmt.Errorf("error happened decoding common.ConfigUpdate: %v", err)
	}
	configGroups := unmarshaledUpdate.GetWriteSet().GetGroups()

	log.Printf("ProposalID: %v, Channel: %v", proposalID, proposal.ChannelID)

	// NOTE: Current implementation assumes that:
	// (1) Application and Consortium are not updated at the same time
	// (2) Orderer organizations are same as Application or Consortium ones
	// (3) One consortium is only used
	for key, group := range configGroups {
		if key == "Application" || key == "Consortiums" {
			organizations := []string{}
			organizationsGroup := group.Groups
			if key == "Consortiums" {
				if len(group.Groups) == 0 {
					break
				}
				//Pick up first consortium's organization Group
				for _, consortium := range group.Groups {
					organizationsGroup = consortium.Groups
					continue
				}
			}
			for org := range organizationsGroup {
				organizations = append(organizations, org)
			}
			log.Printf("Orgs: %v", organizations)

			// Check channel exists
			channelExists, err := s.ChannelExists(ctx, proposal.ChannelID)
			if err != nil {
				return fmt.Errorf("error happend querying ChannelExists: %v", err)
			}

			// Update channel info
			if channelExists {
				err = s.SetOrganizations(ctx, proposal.ChannelID, organizations)
				if err != nil {
					return fmt.Errorf("error happend updating channel info: %v", err)
				}
			} else {
				err = s.CreateChannel(ctx, proposal.ChannelID, "", organizations)
				if err != nil {
					return fmt.Errorf("error happend creating channel info: %v", err)
				}
			}

			if err = ctx.GetStub().SetEvent(fmt.Sprintf("%s.%s", UpdateConfigEvent, proposalID), []byte(proposalID)); err != nil {
				return fmt.Errorf("error happened emitting event: %v", err)
			}
		}
	}

	return nil
}

// GetAllProposals returns the all channel update proposals.
//
// Arguments: none
//
// Returns:
//   0: the map of the all channel update proposals
//   1: error
//
func (s *SmartContract) GetAllProposals(ctx contractapi.TransactionContextInterface) (map[string]*Proposal, error) {

	proposals := make(map[string]*Proposal)
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
		proposal := &Proposal{}
		if err = json.Unmarshal(proposalJSON.Value, proposal); err != nil {
			return nil, fmt.Errorf("error happened unmarshalling a proposal JSON representation to struct: %v", err)
		}
		proposals[proposal.ID] = proposal
	}

	return proposals, nil
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
func (s *SmartContract) GetProposal(ctx contractapi.TransactionContextInterface, proposalID string) (*Proposal, error) {

	compositeKey, err := s.createCompositeKeyForProposal(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("error happend creating composite key for proposal: %v", err)
	}
	proposalJSON, err := ctx.GetStub().GetState(compositeKey)
	if err != nil {
		return nil, fmt.Errorf("error happened reading proposal with id (%v): %v", proposalID, err)
	}
	if proposalJSON == nil {
		return nil, ErrProposalNotFound
	}

	var proposal Proposal
	err = json.Unmarshal(proposalJSON, &proposal)
	if err != nil {
		return nil, fmt.Errorf("error happened unmarshalling a proposal JSON representation to struct: %v", err)
	}
	return &proposal, nil
}

// Internal functions

func (s *SmartContract) getMSPID(ctx contractapi.TransactionContextInterface) (string, error) {
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return "", fmt.Errorf("error happened reading the transaction creator: %v", err)
	}
	return getMSPID(creator)
}

func getMSPID(creator []byte) (string, error) {
	identity := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(creator, identity); err != nil {
		return "", fmt.Errorf("error happened unmarshalling the creator: %v", err)
	}
	return identity.Mspid, nil
}

func validateConfigUpdate(input ProposalInput) error {
	// check if the configUpdate is in the correct format: base64 encoded proto/common.ConfigUpdate
	update, err := base64.StdEncoding.DecodeString(input.ConfigUpdate)
	if err != nil {
		return fmt.Errorf("error happened decoding the configUpdate base64 string: %v", err)
	}
	unmarshaledUpdate := common.ConfigUpdate{}
	if err := proto.Unmarshal(update, &unmarshaledUpdate); err != nil {
		return fmt.Errorf("error happened decoding common.ConfigUpdate: %v", err)
	}

	// check if the channel ID in ConfigUpdate is equals to the channelID in the proposal
	if input.ChannelID != unmarshaledUpdate.GetChannelId() {
		return ErrInconsistentChannelID
	}
	return nil
}

func validateConfigSignature(signature string) error {
	// check if the signature is in the correct format: base64 encoded proto/common.ConfigSignature
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("error happened decoding the signature base64 string: %v", err)
	}
	if err := proto.Unmarshal(sig, &common.ConfigSignature{}); err != nil {
		return fmt.Errorf("error happened decoding common.ConfigSignature: %v", err)
	}
	return nil
}

func (s *SmartContract) passedByVoting(ctx contractapi.TransactionContextInterface, proposal *Proposal) (bool, error) {
	channelID := proposal.ChannelID
	var err error
	// Use organizations in the system channel when creating a channel
	if proposal.Action == CreationAction {
		channelID, err = s.GetSystemChannelID(ctx)
		if err != nil {
			return false, fmt.Errorf("error happend getting the system channel ID: %v", err)
		}
	}
	satisfied, err := s.meetCriteria(ctx, proposal.Artifacts.Signatures, MAJORITY, channelID)
	if err != nil {
		return false, fmt.Errorf("error happened checking to meet criteria: %v", err)
	}
	return satisfied, nil
}

func (s *SmartContract) meetCriteria(ctx contractapi.TransactionContextInterface, signatures map[string]string, criteria string, targetChannel string) (bool, error) {

	criteriaNum, err := s.CountOrganizationsInChannel(ctx, targetChannel)
	if err != nil {
		return false, fmt.Errorf("fail to get the num of organizations: %v", err)
	}

	switch criteria {
	case MAJORITY:
		criteriaNum = criteriaNum/2 + 1
	case ALL:
		// use total org num
	default:
		return false, fmt.Errorf("unknown criteria: %s", criteria)
	}

	if len(signatures) >= criteriaNum {
		return true, nil
	}
	return false, nil
}

func (s *SmartContract) putProposal(ctx contractapi.TransactionContextInterface, proposal *Proposal) error {
	proposalJSON, err := json.Marshal(proposal)
	if err != nil {
		return fmt.Errorf("error happened marshalling the new proposal: %v", err)
	}
	compositeKey, err := s.createCompositeKeyForProposal(ctx, proposal.ID)
	if err != nil {
		return fmt.Errorf("error happend creating composite key for proposal: %v", err)
	}
	if err := ctx.GetStub().PutState(string(compositeKey), proposalJSON); err != nil {
		return fmt.Errorf("error happened persisting the new proposal on the ledger: %v", err)
	}
	return nil
}

func (s *SmartContract) createCompositeKeyForProposal(ctx contractapi.TransactionContextInterface, proposalID string) (string, error) {
	return ctx.GetStub().CreateCompositeKey(ProposalObjectType, []string{proposalID})
}
