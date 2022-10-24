/*
Copyright 2020-2022  Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/hyperledger-labs/fabric-opssc/chaincode/channel-ops/chaincode"
	"github.com/hyperledger-labs/fabric-opssc/chaincode/channel-ops/chaincode/mocks"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/stretchr/testify/require"
)

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

func TestCreateChannel(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	chaincodeStub.CreateCompositeKeyStub = createComposeKey

	sc := chaincode.SmartContract{}

	// Create a application channel without specifying the channel type and the initial organizations
	err := sc.CreateChannel(transactionContext, "mychannel", "", nil)
	require.NoError(t, err)
	key, state := chaincodeStub.PutStateArgsForCall(0)
	require.Equal(t, "channel_mychannel", key)
	expected := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
	}
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))

	// Create a ops channel with specifying initial organizations
	err = sc.CreateChannel(transactionContext, "ops-channel", chaincode.OpsChannelType, []string{"Org1MSP", "Org2MSP"})
	require.NoError(t, err)
	key, state = chaincodeStub.PutStateArgsForCall(1)
	require.Equal(t, "channel_ops-channel", key)
	expected = chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "ops-channel",
		ChannelType: chaincode.OpsChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	expectedJSON, err = json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))

	// Create a system channel with specifying initial organizations
	err = sc.CreateChannel(transactionContext, "system-channel", chaincode.SystemChannelType, []string{"Org1MSP", "Org2MSP"})
	require.NoError(t, err)

	// Case: Create a channel with an invalid channel type
	err = sc.CreateChannel(transactionContext, "mychannel", "invalid", []string{"Org1MSP", "Org2MSP"})
	require.EqualError(t, err, "invalid channel type - expecting system, ops or application")

	// Case: Create a duplicated channel
	chaincodeStub.GetStateReturns([]byte{}, nil)
	err = sc.CreateChannel(transactionContext, "mychannel", chaincode.ApplicationChannelType, []string{"Org1MSP", "Org2MSP"})
	require.EqualError(t, err, "the channel mychannel already exists")

	// Case: Internal state read error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve channel"))
	err = sc.CreateChannel(transactionContext, "mychannel", chaincode.ApplicationChannelType, []string{"Org1MSP", "Org2MSP"})
	require.EqualError(t, err, "failed to read from world state: unable to retrieve channel")
}

func TestReadChannel(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	// Case: Read a channel
	expected := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(expectedJSON, nil)
	sc := chaincode.SmartContract{}
	actual, err := sc.ReadChannel(transactionContext, "mychannel")
	require.NoError(t, err)
	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(actualJSON))

	// Case: Internal state read error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve channel"))
	_, err = sc.ReadChannel(transactionContext, "mychannel")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve channel")

	// Case: Fail when the channel does not exist
	chaincodeStub.GetStateReturns(nil, nil)
	actual, err = sc.ReadChannel(transactionContext, "mychannel")
	require.EqualError(t, err, "the channel mychannel does not exist")
	require.Nil(t, actual)

	// Case: Fail to create composite key
	chaincodeStub.CreateCompositeKeyReturns("", fmt.Errorf("failed to create composite key"))
	actual, err = sc.ReadChannel(transactionContext, "mychannel")
	require.EqualError(t, err, "failed to create composite key for channel: failed to create composite key")
	require.Nil(t, actual)
}

func TestChannelExists(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	// Case: A channel is found
	expected := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(expectedJSON, nil)
	sc := chaincode.SmartContract{}
	actual, err := sc.ChannelExists(transactionContext, "mychannel")
	require.NoError(t, err)
	require.Equal(t, true, actual)

	// Case: Fail to create composite key
	chaincodeStub.CreateCompositeKeyReturns("", fmt.Errorf("failed to create composite key"))
	_, err = sc.ChannelExists(transactionContext, "mychannel")
	require.EqualError(t, err, "failed to create composite key for channel: failed to create composite key")
}

func TestAddOrganization(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	// Case: Add an organization to a channel with some organizations
	channel := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	bytes, err := json.Marshal(channel)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	sc := chaincode.SmartContract{}
	err = sc.AddOrganization(transactionContext, "mychannel", "Org3MSP")
	require.NoError(t, err)
	_, state := chaincodeStub.PutStateArgsForCall(0)
	expected := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": "",
			"Org3MSP": ""},
	}
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))

	// Case: Add an organization to a channel without organization
	channel = chaincode.Channel{
		ObjectType:    chaincode.ChannelObjectType,
		ID:            "mychannel",
		ChannelType:   chaincode.ApplicationChannelType,
		Organizations: nil,
	}
	bytes, err = json.Marshal(channel)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)
	err = sc.AddOrganization(transactionContext, "mychannel", "Org3MSP")
	require.NoError(t, err)
	_, state = chaincodeStub.PutStateArgsForCall(1)
	expected = chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org3MSP": ""},
	}
	expectedJSON, err = json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))

	// Case: Fail when the channel does not exist
	chaincodeStub.GetStateReturns(nil, nil)
	err = sc.AddOrganization(transactionContext, "mychannel", "Org3MSP")
	require.EqualError(t, err, "failed to read channel: the channel mychannel does not exist")

	// Case: Internal state read error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve channel"))
	err = sc.AddOrganization(transactionContext, "mychannel", "Org3MSP")
	require.EqualError(t, err, "failed to read channel: failed to read from world state: unable to retrieve channel")
}

func TestSetOrganizations(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	channel := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": ""},
	}
	bytes, err := json.Marshal(channel)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	sc := chaincode.SmartContract{}
	err = sc.SetOrganizations(transactionContext, "mychannel", []string{"Org2MSP", "Org3MSP"})
	require.NoError(t, err)
	_, state := chaincodeStub.PutStateArgsForCall(0)
	expected := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org2MSP": "",
			"Org3MSP": ""},
	}
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))

	chaincodeStub.GetStateReturns(nil, nil)
	err = sc.SetOrganizations(transactionContext, "mychannel", []string{"Org2MSP", "Org3MSP"})
	require.EqualError(t, err, "failed to read channel: the channel mychannel does not exist")

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve channel"))
	err = sc.SetOrganizations(transactionContext, "mychannel", []string{"Org2MSP", "Org3MSP"})
	require.EqualError(t, err, "failed to read channel: failed to read from world state: unable to retrieve channel")
}

func TestCountOrganizationsInChannel(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	channel := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	bytes, err := json.Marshal(channel)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	sc := chaincode.SmartContract{}
	count, err := sc.CountOrganizationsInChannel(transactionContext, "mychannel")
	require.NoError(t, err)
	require.Equal(t, 2, count)

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve channel"))
	_, err = sc.CountOrganizationsInChannel(transactionContext, "mychannel")
	require.EqualError(t, err, "failed to read channel: failed to read from world state: unable to retrieve channel")

	chaincodeStub.GetStateReturns(nil, nil)
	_, err = sc.CountOrganizationsInChannel(transactionContext, "mychannel")
	require.EqualError(t, err, "failed to read channel: the channel mychannel does not exist")
}

func TestGetOrganizationsInChannel(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	channel := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	bytes, err := json.Marshal(channel)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	sc := chaincode.SmartContract{}
	orgList, err := sc.GetOrganizationsInChannel(transactionContext, "mychannel")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"Org1MSP", "Org2MSP"}, orgList)

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve channel"))
	_, err = sc.GetOrganizationsInChannel(transactionContext, "mychannel")
	require.EqualError(t, err, "failed to read channel: failed to read from world state: unable to retrieve channel")

	chaincodeStub.GetStateReturns(nil, nil)
	_, err = sc.GetOrganizationsInChannel(transactionContext, "mychannel")
	require.EqualError(t, err, "failed to read channel: the channel mychannel does not exist")
}

func TestGetAllChannels(t *testing.T) {
	channel1 := &chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "system-channel",
		ChannelType: chaincode.SystemChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	channel1JSON, err := json.Marshal(channel1)
	require.NoError(t, err)

	channel2 := &chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	channel2JSON, err := json.Marshal(channel2)
	require.NoError(t, err)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, true)
	iterator.HasNextReturnsOnCall(2, false)
	iterator.NextReturnsOnCall(0, &queryresult.KV{Value: channel1JSON}, nil)
	iterator.NextReturnsOnCall(1, &queryresult.KV{Value: channel2JSON}, nil)

	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	sc := &chaincode.SmartContract{}
	channels, err := sc.GetAllChannels(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Channel{channel1, channel2}, channels)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	channels, err = sc.GetAllChannels(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, channels)

	chaincodeStub.GetStateByPartialCompositeKeyReturns(nil, fmt.Errorf("failed retrieving all channels"))
	channels, err = sc.GetAllChannels(transactionContext)
	require.EqualError(t, err, "failed retrieving all channels")
	require.Nil(t, channels)
}

func TestGetSystemChannelID(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	// Positive case
	channel := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "system-channel",
		ChannelType: chaincode.SystemChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	bytes, err := json.Marshal(channel)
	require.NoError(t, err)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	sc := chaincode.SmartContract{}
	actual, err := sc.GetSystemChannelID(transactionContext)
	require.NoError(t, err)
	require.Equal(t, "system-channel", actual)

	// Error case: Only application channels exist
	channel = chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "mychannel",
		ChannelType: chaincode.ApplicationChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	bytes, err = json.Marshal(channel)
	require.NoError(t, err)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)
	chaincodeStub.GetStateByPartialCompositeKeyReturns(iterator, nil)
	_, err = sc.GetSystemChannelID(transactionContext)
	require.EqualError(t, err, "system channel is not found")

	// Error case: Fail to get channels internally
	chaincodeStub.GetStateByPartialCompositeKeyReturns(nil, fmt.Errorf("failed retrieving all channels"))
	_, err = sc.GetSystemChannelID(transactionContext)
	require.EqualError(t, err, "failed to get system channel: failed retrieving all channels")
}

func TestUpdateChannelType(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	// Positive case
	channel := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "system-channel",
		ChannelType: chaincode.SystemChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	bytes, err := json.Marshal(channel)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)

	sc := chaincode.SmartContract{}

	err = sc.UpdateChannelType(transactionContext, "system-channel", chaincode.DisableChannelType)
	require.NoError(t, err)

	_, state := chaincodeStub.PutStateArgsForCall(0)
	expected := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "system-channel",
		ChannelType: chaincode.DisableChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(state))

	// Error case: Update a channel to an invalid type
	err = sc.UpdateChannelType(transactionContext, "system-channel", "invalid-type")
	require.EqualError(t, err, "invalid channel type - expecting system, ops, application or disable")

	// Error case: Update an unavailable channel
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve channel"))
	err = sc.UpdateChannelType(transactionContext, "system-channel", chaincode.DisableChannelType)
	require.EqualError(t, err, "failed to read channel: failed to read from world state: unable to retrieve channel")
}

func TestGetChannelType(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	// Positive case
	channel := chaincode.Channel{
		ObjectType:  chaincode.ChannelObjectType,
		ID:          "system-channel",
		ChannelType: chaincode.SystemChannelType,
		Organizations: map[string]string{
			"Org1MSP": "",
			"Org2MSP": ""},
	}
	bytes, err := json.Marshal(channel)
	require.NoError(t, err)
	chaincodeStub.GetStateReturns(bytes, nil)

	sc := chaincode.SmartContract{}

	channelType, err := sc.GetChannelType(transactionContext, "system-channel")
	require.NoError(t, err)
	require.Equal(t, chaincode.SystemChannelType, channelType)

	// Error case: Update an unavailable channel
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve channel"))
	_, err = sc.GetChannelType(transactionContext, "system-channel")
	require.EqualError(t, err, "failed to read channel: failed to read from world state: unable to retrieve channel")
}
