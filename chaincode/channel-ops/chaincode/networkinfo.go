/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for operating channels
type SmartContract struct {
	contractapi.Contract
}

// Channel describes channel information which includes the channel ID and organizations participating in the channel.
type Channel struct {
	ObjectType    string            `json:"docType"`       //docType is used to distinguish the various types of objects in state database
	ID            string            `json:"ID"`            // Channel ID
	ChannelType   string            `json:"channelType"`   // Channel Type
	Organizations map[string]string `json:"organizations"` // Set of MSP IDs
}

// Object types
const (
	ChannelObjectType = "channel"
)

// Channel types
const (
	SystemChannelType      = "system"
	OpsChannelType         = "ops"
	ApplicationChannelType = "application"
	DisableChannelType     = "disable"
)

// ChaincodeEvents
const (
	NetworkUpdateEvent = "networkUpdateEvent"
)

// CreateChannel issues a new channel information.
//
// Arguments:
//   0: channelID - the channel ID to be recorded
//   1: channelType - the type of the channel (if this is empty, "application" is set as the default value)
//   2: mspIDs - the list of members of the channel
//
// Returns:
//   0: error
//
func (s *SmartContract) CreateChannel(ctx contractapi.TransactionContextInterface, channelID string, channelType string, mspIDs []string) error {

	exists, err := s.ChannelExists(ctx, channelID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the channel %s already exists", channelID)
	}

	if channelType == "" {
		channelType = ApplicationChannelType
	}
	if channelType != SystemChannelType && channelType != OpsChannelType && channelType != ApplicationChannelType {
		return fmt.Errorf("invalid channel type - expecting %s, %s or %s", SystemChannelType, OpsChannelType, ApplicationChannelType)
	}

	channel := &Channel{
		ObjectType:  ChannelObjectType,
		ID:          channelID,
		ChannelType: channelType,
	}

	if mspIDs != nil {
		channel.Organizations = make(map[string]string)

		for _, mspID := range mspIDs {
			channel.Organizations[mspID] = ""
		}
	}
	return s.putChannel(ctx, channel)
}

// UpdateChannelType replaces the type of the given channel.
//
// Arguments:
//   0: channelID - the target channel ID
//   1: channelType - the type of the channel
//
// Returns:
//   0: error
//
func (s *SmartContract) UpdateChannelType(ctx contractapi.TransactionContextInterface, channelID string, channelType string) error {

	channel, err := s.ReadChannel(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to read channel: %v", err)
	}

	if channelType != SystemChannelType && channelType != OpsChannelType && channelType != ApplicationChannelType && channelType != DisableChannelType {
		return fmt.Errorf("invalid channel type - expecting %s, %s, %s or %s", SystemChannelType, OpsChannelType, ApplicationChannelType, DisableChannelType)
	}

	channel.ChannelType = channelType
	return s.putChannel(ctx, channel)
}

// AddOrganization upserts an organization's MSP ID as a member of the given channel.
//
// Arguments:
//   0: channelID - the target channel ID
//   1: mspID - the MSP ID of newly added or updated member
//
// Returns:
//   0: error
//
func (s *SmartContract) AddOrganization(ctx contractapi.TransactionContextInterface, channelID string, mspID string) error {
	channel, err := s.ReadChannel(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to read channel: %v", err)
	}

	if channel.Organizations == nil {
		channel.Organizations = make(map[string]string)
	}
	channel.Organizations[mspID] = ""

	return s.putChannel(ctx, channel)
}

// SetOrganizations replaces the members of the given channel with the given MSP ID list.
//
// Arguments:
//   0: channelID - the target channel ID
//   1: mspIDs - the MSP ID list
//
// Returns:
//   0: error
//
func (s *SmartContract) SetOrganizations(ctx contractapi.TransactionContextInterface, channelID string, mspIDs []string) error {
	channel, err := s.ReadChannel(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to read channel: %v", err)
	}

	channel.Organizations = make(map[string]string)

	for _, mspID := range mspIDs {
		channel.Organizations[mspID] = ""
	}

	return s.putChannel(ctx, channel)
}

// ReadChannel returns the channel information stored in the ledger with the given channel ID.
//
// Arguments:
//   0: channelID - the target channel ID
//
// Returns:
//   0: the channel with the given ID
//   1: error
//
func (s *SmartContract) ReadChannel(ctx contractapi.TransactionContextInterface, channelID string) (*Channel, error) {
	compositeKey, err := s.createCompositeKeyForChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key for channel: %v", err)
	}
	channelJSON, err := ctx.GetStub().GetState(compositeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if channelJSON == nil {
		return nil, fmt.Errorf("the channel %s does not exist", channelID)
	}

	var channel Channel
	err = json.Unmarshal(channelJSON, &channel)
	if err != nil {
		return nil, err
	}

	return &channel, nil
}

// ChannelExists returns true when the channel with the given ID exists in the ledger.
//
// Arguments:
//   0: channelID - the target channel ID
//
// Returns:
//   0: whether that the channel with the given ID exists or not
//   1: error
//
func (s *SmartContract) ChannelExists(ctx contractapi.TransactionContextInterface, channelID string) (bool, error) {
	compositeKey, err := s.createCompositeKeyForChannel(ctx, channelID)
	if err != nil {
		return false, fmt.Errorf("failed to create composite key for channel: %v", err)
	}
	channelJSON, err := ctx.GetStub().GetState(compositeKey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return channelJSON != nil, nil
}

// CountOrganizationsInChannel returns the number of organizations participating in the given channel.
//
// Arguments:
//   0: channelID - the target channel ID
//
// Returns:
//   0: the number of organizations participating in the given channel
//   1: error
//
func (s *SmartContract) CountOrganizationsInChannel(ctx contractapi.TransactionContextInterface, channelID string) (int, error) {
	channel, err := s.ReadChannel(ctx, channelID)
	if err != nil {
		return 0, fmt.Errorf("failed to read channel: %v", err)
	}
	return len(channel.Organizations), nil
}

// GetOrganizationsInChannel returns the list of organizations participating in the given channel.
//
// Arguments:
//   0: channelID - the target channel ID
//
// Returns:
//   0: the list of organizations participating in given channel
//   1: error
//
func (s *SmartContract) GetOrganizationsInChannel(ctx contractapi.TransactionContextInterface, channelID string) ([]string, error) {
	channel, err := s.ReadChannel(ctx, channelID)
	oList := []string{}
	if err != nil {
		return oList, fmt.Errorf("failed to read channel: %v", err)
	}
	for orgMSP := range channel.Organizations {
		oList = append(oList, orgMSP)
	}
	return oList, nil
}

// GetChannelType returns the channel type of the given channel
//
// Arguments:
//   0: channelID - the target channel ID
//
// Returns:
//   0: the channel type
//   1: error
//
func (s *SmartContract) GetChannelType(ctx contractapi.TransactionContextInterface, channelID string) (string, error) {
	channel, err := s.ReadChannel(ctx, channelID)
	if err != nil {
		return "", fmt.Errorf("failed to read channel: %v", err)
	}
	return channel.ChannelType, nil
}

// GetAllChannels returns the all channel information stored in the ledger.
//
// Arguments: none
//
// Returns:
//   0: the list of the all channel information
//   1: error
//
func (s *SmartContract) GetAllChannels(ctx contractapi.TransactionContextInterface) ([]*Channel, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(ChannelObjectType, []string{})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var channels []*Channel
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var channel Channel
		err = json.Unmarshal(queryResponse.Value, &channel)
		if err != nil {
			return nil, err
		}
		channels = append(channels, &channel)
	}

	return channels, nil
}

// GetSystemChannelID returns the system channel ID stored in the ledger
//
// Arguments: none
//
// Returns:
//   0: the system channel ID
//   1: error
//
func (s *SmartContract) GetSystemChannelID(ctx contractapi.TransactionContextInterface) (string, error) {

	channels, err := s.GetAllChannels(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get system channel: %v", err)
	}

	for _, channel := range channels {
		if channel.ChannelType == SystemChannelType {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("system channel is not found")
}

// Internal functions
func (s *SmartContract) putChannel(ctx contractapi.TransactionContextInterface, channel *Channel) error {
	channelJSON, err := json.Marshal(channel)
	if err != nil {
		return err
	}
	compositeKey, err := s.createCompositeKeyForChannel(ctx, channel.ID)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(compositeKey, channelJSON)
}

func (s *SmartContract) createCompositeKeyForChannel(ctx contractapi.TransactionContextInterface, channelID string) (string, error) {
	return ctx.GetStub().CreateCompositeKey(ChannelObjectType, []string{channelID})
}
