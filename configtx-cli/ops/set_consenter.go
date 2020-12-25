/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-config/configtx"
)

// OutputConfigTXToSetConsenter creates ConfigUpdate for upserting the consenter to the specified channel by using the specified profile and
// then outputs the ConfigUpdate to the specified output path with the specified output format.
func OutputConfigTXToSetConsenter(blockPath string, outputDir string, outputFile string, outputFormat string, channelID string, profilePath string) error {
	configtx, err := getConfigTxFromBlock(blockPath)
	if err != nil {
		return errors.Wrap(err, "failed to get configtx from block")
	}

	profile, err := loadConsenterProfile(profilePath)
	if err != nil {
		return errors.Wrap(err, "failed to load consenter profile")
	}

	err = setConsenter(configtx, profile)
	if err != nil {
		return errors.Wrap(err, "failed to set the consenter to the current config")
	}

	err = output(configtx, filepath.Join(outputDir, outputFile), outputFormat, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func setConsenter(c configtx.ConfigTx, profile Consenter) error {

	logger.Info("Set consenter")
	newConsenter, err := newConsenter(profile)
	if err != nil {
		return errors.Wrap(err, "failed to get the original orderer configuration")
	}

	ordererGroup := c.Orderer()
	if ordererGroup == nil || reflect.DeepEqual(ordererGroup, &configtx.OrdererGroup{}) {
		return errors.New("the orderer is not found in the current config")
	}

	ordererConfig, err := ordererGroup.Configuration()
	if err != nil {
		return errors.Wrap(err, "failed to get the original orderer configuration")
	}

	// NOTE: Current configtx only supports 'add consenter'.
	// To support update consenter, this func directly operates ordererConfig.EtcdRaft.Consenters field.
	// This function identifies consenters with matching addresses as the same.
	var update = false
	consenters := ordererConfig.EtcdRaft.Consenters
	for i, c := range consenters {
		if reflect.DeepEqual(c.Address, newConsenter.Address) {
			ordererConfig.EtcdRaft.Consenters[i] = newConsenter
			update = true
			break
		}
	}
	if !update {
		ordererConfig.EtcdRaft.Consenters = append(ordererConfig.EtcdRaft.Consenters, newConsenter)
	}

	if err = ordererGroup.SetConfiguration(ordererConfig); err != nil {
		return errors.Wrap(err, "failed to set the orderer configuration")
	}
	return nil
}
