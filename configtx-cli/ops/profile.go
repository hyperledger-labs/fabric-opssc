/*
Copyright IBM Corp. All Rights Reserved.

Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Ops profile structures is based on configgen's config and msp/configbuilder
// Ref: https://github.com/hyperledger/fabric/blob/master/internal/configtxgen/genesisconfig/config.go
// Ref: https://github.com/hyperledger/fabric/blob/master/msp/configbuilder.go

// OrganizationProfile is a profile to add or update an organization.
type OrganizationProfile struct {
	Name     string             `yaml:"Name"`
	ID       string             `yaml:"ID"`
	MSPDir   string             `yaml:"MSPDir"`
	MSP      MSP                `yaml:"MSP"`
	Policies map[string]*Policy `yaml:"Policies"`

	AnchorPeers      []*AnchorPeer `yaml:"AnchorPeers"`
	OrdererEndpoints []string      `yaml:"OrdererEndpoints"`
}

// MSP contains the configuration information for a MSP
type MSP struct {
	RootCerts                     []string       `yaml:"RootCerts"`
	IntermediateCerts             []string       `yaml:"IntermediateCerts"`
	Admins                        []string       `yaml:"Admins"`
	RevocationList                []string       `yaml:"RevocationList"`
	OrganizationalUnitIdentifiers []OUIdentifier `yaml:"OrganizationalUnitIdentifiers"`
	TLSRootCerts                  []string       `yaml:"TLSRootCerts"`
	TLSIntermediateCerts          []string       `yaml:"TLSIntermediateCerts"`
	NodeOUs                       NodeOUs        `yaml:"NodeOUs"`
}

// NodeOUs contains information on how to tell apart clients, peers and orderers
// based on OUs. If the check is enforced, by setting Enabled to true,
// the MSP will consider an identity valid if it is an identity of a client, a peer or
// an orderer. An identity should have only one of these special OUs.
type NodeOUs struct {
	Enable              bool          `yaml:"Enable"`
	ClientOUIdentifier  *OUIdentifier `yaml:"ClientOUIdentifier"`
	PeerOUIdentifier    *OUIdentifier `yaml:"PeerOUIdentifier"`
	AdminOUIdentifier   *OUIdentifier `yaml:"AdminOUIdentifier"`
	OrdererOUIdentifier *OUIdentifier `yaml:"OrdererOUIdentifier"`
}

// OUIdentifier is used to represent an OU and an associated trusted certificate.
type OUIdentifier struct {
	OrganizationalUnitIdentifier string `yaml:"OrganizationalUnitIdentifier"`
	Certificate                  string `yaml:"Certificate"`
}

// Policy contains the configuration information for a policy
type Policy struct {
	Type string `yaml:"Type"`
	Rule string `yaml:"Rule"`
}

// AnchorPeer contains the configuration information for a anchor peer
type AnchorPeer struct {
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`
}

// ChannelProfile is a profile to update the channel configuration on a channel.
type ChannelProfile struct {
	Application *Application `yaml:"Application"`
	Channel     *Channel     `yaml:"Channel"`
}

// ChannelCreationProfile is a profile to create a channel.
type ChannelCreationProfile struct {
	Consortium  string                         `yaml:"Consortium"`
	Application *ApplicationForChannelCreation `yaml:"Application"`
}

// ApplicationForChannelCreation contains the configuration information for a application channel.
// (This is only used for channel creation)
type ApplicationForChannelCreation struct {
	Policies      map[string]*Policy `yaml:"Policies"`
	ACLs          map[string]string  `yaml:"ACLs"`
	Capabilities  []string           `yaml:"Capabilities"`
	Organizations []string           `yaml:"Organizations"`
}

// Channel contains the configuration information for a system channel.
type Channel struct {
	Policies map[string]*Policy `yaml:"Policies"`
}

// Application contains the configuration information for a application channel.
// (This is only used for channel update)
type Application struct {
	Policies map[string]*Policy `yaml:"Policies"`
	ACLs     map[string]string  `yaml:"ACLs"`
}

// OrdererProfile is a profile to update the orderer configuration on a channel.
type OrdererProfile struct {
	OrdererType     string           `yaml:"OrdererType"`
	BatchTimeout    string           `yaml:"BatchTimeout"`
	BatchSize       BatchSize        `yaml:"BatchSize"`
	EtcdRaftOptions *EtcdRaftOptions `yaml:"EtcdRaftOptions"`
	MaxChannels     uint64           `yaml:"MaxChannels"`
	// Capabilities    map[string]bool    `yaml:"Capabilities"` // Out of scope
	Policies map[string]*Policy `yaml:"Policies"`
}

// BatchSize contains configuration affecting the size of batches.
type BatchSize struct {
	MaxMessageCount   uint32 `yaml:"MaxMessageCount"`
	AbsoluteMaxBytes  string `yaml:"AbsoluteMaxBytes"`
	PreferredMaxBytes string `yaml:"PreferredMaxBytes"`
}

// EtcdRaftOptions contains configuration for Raft orderer.
type EtcdRaftOptions struct {
	TickInterval         string `yaml:"TickInterval"`
	ElectionTick         uint32 `yaml:"ElectionTick"`
	HeartbeatTick        uint32 `yaml:"HeartbeatTick"`
	MaxInflightBlocks    uint32 `yaml:"MaxInflightBlocks"`
	SnapshotIntervalSize string `yaml:"SnapshotIntervalSize"`
}

// Consenter contains the configuration information for a consenter.
type Consenter struct {
	Host          string `yaml:"Host"`
	Port          int    `yaml:"Port"`
	ClientTLSCert string `yaml:"ClientTLSCert"`
	ServerTLSCert string `yaml:"ServerTLSCert"`
}

// Instruction represents a update operation.
type Instruction struct {
	Command    string      `yaml:"Command"`
	Parameters interface{} `yaml:"Parameters"`
}

// SetOrgParameters represents to do set-org.
type SetOrgParameters struct {
	OrgType string              `yaml:"OrgType"`
	Org     OrganizationProfile `yaml:"Org"`
}

// RemoveOrgParameters represents to do remove-org.
type RemoveOrgParameters struct {
	OrgType string `yaml:"OrgType"`
	OrgName string `yaml:"OrgName"`
}

// SetChannelParameters represents to do set-channel.
type SetChannelParameters struct {
	Channel ChannelProfile `yaml:"Channel"`
}

// SetOrdererParameters represents to do set-orderer.
type SetOrdererParameters struct {
	Orderer OrdererProfile `yaml:"Orderer"`
}

// SetConsenterParameters represents to do set-consenter.
type SetConsenterParameters struct {
	Consenter Consenter `yaml:"Consenter"`
}

// RemoveConsenterParameters represents to do remove-consenter.
type RemoveConsenterParameters struct {
	ConsenterAddress string `yaml:"ConsenterAddress"`
}

// loadOrganizationProfile loads the profile to set an organization from the specified file.
func loadOrganizationProfile(path string) (OrganizationProfile, error) {
	profile := OrganizationProfile{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return profile, errors.Wrap(err, "failed to read the organization profile")
	}

	err = yaml.Unmarshal([]byte(data), &profile)
	if err != nil {
		return profile, errors.Wrap(err, "failed to unmarshal the organization profile")
	}

	return profile, nil
}

// loadChannelProfile loads the profile to update channel configuration from the specified file.
func loadChannelProfile(configPath string) (ChannelProfile, error) {
	channelProfile := ChannelProfile{}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return channelProfile, errors.Wrap(err, "failed to read the channel profile")
	}

	err = yaml.Unmarshal([]byte(data), &channelProfile)
	if err != nil {
		return channelProfile, errors.Wrap(err, "failed to unmarshal the channel profile")
	}

	return channelProfile, nil
}

// loadChannelCreationProfile loads the profile to create an channel from the specified file.
func loadChannelCreationProfile(configPath string) (ChannelCreationProfile, error) {
	profile := ChannelCreationProfile{}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return profile, errors.Wrap(err, "failed to read the channel creation profile")
	}

	err = yaml.Unmarshal([]byte(data), &profile)
	if err != nil {
		return profile, errors.Wrap(err, "failed to unmarshal the channel creation profile")
	}

	return profile, nil
}

// loadConsenterProfile loads the profile to set a consenter from the specified file.
func loadConsenterProfile(configPath string) (Consenter, error) {
	consenterConfig := Consenter{}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return consenterConfig, errors.Wrap(err, "failed to read the consenter profile")
	}

	err = yaml.Unmarshal([]byte(data), &consenterConfig)
	if err != nil {
		return consenterConfig, errors.Wrap(err, "failed to unmarshal the consenter profile")
	}

	return consenterConfig, nil
}

// loadOrdererProfile loads the profile to set orderer configuration from the specified file.
func loadOrdererProfile(configPath string) (OrdererProfile, error) {

	ordererProfile := OrdererProfile{}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ordererProfile, errors.Wrap(err, "failed to read the orderer profile")
	}

	err = yaml.Unmarshal([]byte(data), &ordererProfile)
	if err != nil {
		return ordererProfile, errors.Wrap(err, "failed to unmarshal the orderer profile")
	}

	return ordererProfile, nil
}

// loadInstructionsProfile loads the profile to do multiple ops the from the specified file.
func loadInstructionsProfile(path string) ([]Instruction, error) {
	var instructions []Instruction
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return instructions, errors.Wrap(err, "failed to read the multiple ops profile")
	}

	err = yaml.Unmarshal([]byte(data), &instructions)
	if err != nil {
		return instructions, errors.Wrap(err, "failed to unmarshal the multiple ops profile")
	}

	return instructions, nil
}
