/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"path/filepath"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-config/configtx"
	"github.com/hyperledger/fabric-config/configtx/orderer"
)

// OutputConfigTXToSetOrderer creates ConfigUpdate for updating the orderer configuration of the specified channel by using the specified profile and
// then outputs the ConfigUpdate to the specified output path with the specified output format.
func OutputConfigTXToSetOrderer(blockPath string, outputDir string, outputFile string, outputFormat string, channelID string, profilePath string) error {
	configtx, err := getConfigTxFromBlock(blockPath)
	if err != nil {
		return errors.Wrap(err, "failed to get configtx from block")
	}

	profile, err := loadOrdererProfile(profilePath)
	if err != nil {
		return errors.Wrap(err, "failed to load orderer profile")
	}

	err = setOrderer(configtx, profile)
	if err != nil {
		return errors.Wrap(err, "failed to set the orderer configuration to the current configuration")
	}

	err = output(configtx, filepath.Join(outputDir, outputFile), outputFormat, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func setOrderer(c configtx.ConfigTx, newOrdererProfile OrdererProfile) error {

	logger.Info("Set orderer")
	ordererGroup := c.Orderer()
	if ordererGroup == nil || reflect.DeepEqual(ordererGroup, &configtx.OrdererGroup{}) {
		return errors.New("the orderer is not found in the current configuration")
	}

	originalOrdererConfig, err := ordererGroup.Configuration()
	if err != nil {
		return errors.Wrap(err, "failed to get the original orderer configuration")
	}

	// Validate new orderer profile and convert to configtx.Orderer object
	if newOrdererProfile.OrdererType != orderer.ConsensusTypeEtcdRaft {
		return errors.New("this tool only supports " + orderer.ConsensusTypeEtcdRaft)
	}
	if newOrdererProfile.Policies == nil || len(newOrdererProfile.Policies) < 1 {
		return errors.New("orderer policies are not found")
	}
	if newOrdererProfile.EtcdRaftOptions == nil {
		return errors.New("the parameter 'EtcdRaftOptions' is not found")
	}
	absoluteMaxBytes, err := decodeByteSize(newOrdererProfile.BatchSize.AbsoluteMaxBytes)
	if err != nil {
		return errors.Wrap(err, "can not parse BatchSize.AbsoluteMaxBytes")
	}
	preferredMaxBytes, err := decodeByteSize(newOrdererProfile.BatchSize.PreferredMaxBytes)
	if err != nil {
		return errors.Wrap(err, "can not parse BatchSize.PreferredMaxBytes")
	}
	batchTimepout, err := time.ParseDuration(newOrdererProfile.BatchTimeout)
	if err != nil {
		return errors.Wrap(err, "can not parse BatchTimeout")
	}
	snapshotIntervalSize, err := decodeByteSize(newOrdererProfile.EtcdRaftOptions.SnapshotIntervalSize)
	if err != nil {
		return errors.Wrap(err, "can not parse EtcdRaftOptions.SnapshotIntervalSize")
	}
	newPolicies, err := newPolicies(newOrdererProfile.Policies)
	if err != nil {
		return errors.Wrap(err, "failed to create policies data structure for configtx from the profile")
	}

	newOrdererConfig := configtx.Orderer{
		OrdererType:  orderer.ConsensusTypeEtcdRaft,
		Capabilities: originalOrdererConfig.Capabilities,
		BatchTimeout: batchTimepout,
		BatchSize: orderer.BatchSize{
			AbsoluteMaxBytes:  absoluteMaxBytes,
			MaxMessageCount:   newOrdererProfile.BatchSize.MaxMessageCount,
			PreferredMaxBytes: preferredMaxBytes,
		},
		EtcdRaft: orderer.EtcdRaft{
			Consenters: originalOrdererConfig.EtcdRaft.Consenters,
			Options: orderer.EtcdRaftOptions{
				TickInterval:         newOrdererProfile.EtcdRaftOptions.TickInterval,
				ElectionTick:         newOrdererProfile.EtcdRaftOptions.ElectionTick,
				HeartbeatTick:        newOrdererProfile.EtcdRaftOptions.HeartbeatTick,
				MaxInflightBlocks:    newOrdererProfile.EtcdRaftOptions.MaxInflightBlocks,
				SnapshotIntervalSize: snapshotIntervalSize,
			},
		},
		MaxChannels: newOrdererProfile.MaxChannels,
		Policies:    newPolicies,
		State:       originalOrdererConfig.State,
	}
	err = ordererGroup.SetConfiguration(newOrdererConfig)
	if err != nil {
		return errors.Wrap(err, "failed to set the orderer configuration to the current configuration")
	}

	// Replace Orderer Policies
	// NOTE: ordererGroup.SetConfiguration() skips updating policies.
	// So, this function replaces the policies as follows.
	if err = ordererGroup.SetPolicies(configtx.AdminsPolicyKey, newPolicies); err != nil {
		return errors.Wrap(err, "failed to set the orderer policies to the current configuration")
	}

	return nil
}
