/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-config/configtx"
	"github.com/hyperledger/fabric-config/configtx/orderer"
)

// OutputConfigTXToRemoveConsenter creates ConfigUpdate for remove the specified consenter from the specified channel and
// then outputs the ConfigUpdate to the specified output path with the specified output format.
func OutputConfigTXToRemoveConsenter(blockPath string, outputDir string, outputFile string, outputFormat string, channelID string, address string) error {
	configtx, err := getConfigTxFromBlock(blockPath)
	if err != nil {
		return errors.Wrap(err, "failed to get configtx from block")
	}

	err = removeConsenter(configtx, address)
	if err != nil {
		return errors.Wrap(err, "failed to remove the consenter from the current config")
	}

	err = output(configtx, filepath.Join(outputDir, outputFile), outputFormat, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func removeConsenter(c configtx.ConfigTx, address string) error {

	logger.Info("remove consenter")

	// TODO: Should improve Regular expression to parse address
	regexp := regexp.MustCompile(`^(.+):(\d+)$`)
	result := regexp.FindAllStringSubmatch(address, -1)
	if len(result) != 1 {
		return errors.New("address parsing Failed (format should be: ^(.+):(\\d+)$")
	}

	port, err := strconv.Atoi(result[0][2])
	if err != nil {
		return errors.New("address's port parsing Failed")
	}

	targetAddress := orderer.EtcdAddress{
		Host: result[0][1],
		Port: port,
	}

	ordererGroup := c.Orderer()
	if ordererGroup == nil || reflect.DeepEqual(ordererGroup, &configtx.OrdererGroup{}) {
		return errors.New("the orderer is not found in the current configuration")
	}

	ordererConfig, err := ordererGroup.Configuration()
	if err != nil {
		return errors.Wrap(err, "failed to get the original orderer configuration")
	}

	// NOTE: Current configtx's removeConsenter() only supports complete matching (not only address but also TLS certs).
	// This function allows you to remove the consenter whose address matches only.
	consenters := ordererConfig.EtcdRaft.Consenters
	found := false
	logger.Debugf("original consenters: %v", consenters)
	for i, c := range consenters {
		if reflect.DeepEqual(c.Address, targetAddress) {
			ordererConfig.EtcdRaft.Consenters = append(consenters[:i], consenters[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return errors.New("the target consenter is not found in the original configuration")
	}
	logger.Debugf("updated consenters: %v", ordererConfig.EtcdRaft.Consenters)
	if err = ordererGroup.SetConfiguration(ordererConfig); err != nil {
		return errors.Wrap(err, "failed to set the orderer configuration")
	}
	return nil
}
