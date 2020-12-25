/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"fmt"
	"path/filepath"

	"github.com/hyperledger/fabric-config/configtx"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// OutputConfigTXToDoMultipleOps creates ConfigUpdate for executing multiple operations to update the specified channel by using the specified profile and
// then outputs the ConfigUpdate to the specified output path with the specified output format.
func OutputConfigTXToDoMultipleOps(blockPath string, outputDir string, outputFile string, outputFormat string, channelID string, profile string) error {
	configtx, err := getConfigTxFromBlock(blockPath)
	if err != nil {
		return errors.Wrap(err, "failed to get configtx from block")
	}

	instructions, err := loadInstructionsProfile(profile)
	if err != nil {
		return errors.Wrap(err, "failed to load multiple ops profile")
	}

	for _, instruction := range instructions {
		switch instruction.Command {
		case "set-org":
			if err = setOrgAsPartOfMultipleOps(configtx, instruction); err != nil {
				return errors.Wrap(err, "failed to do 'set-org'")
			}
		case "remove-org":
			if err = removeOrgAsPartOfMultipleOps(configtx, instruction); err != nil {
				return errors.Wrap(err, "failed to do 'remove-org'")
			}
		case "set-channel":
			if err = setChannelAsPartOfMultipleOps(configtx, instruction); err != nil {
				return errors.Wrap(err, "failed to do 'set-channel'")
			}
		case "set-orderer":
			if err = setOrdererAsPartOfMultipleOps(configtx, instruction); err != nil {
				return errors.Wrap(err, "failed to do 'set-orderer'")
			}
		case "set-consenter":
			if err = setConsenterAsPartOfMultipleOps(configtx, instruction); err != nil {
				return errors.Wrap(err, "failed to do 'set-consenter'")
			}
		case "remove-consenter":
			if err = removeConsenterAsPartOfMultipleOps(configtx, instruction); err != nil {
				return errors.Wrap(err, "failed to do 'remove-consenter'")
			}
		default:
			return errors.New(fmt.Sprintf("command is not found: %s", instruction.Command))
		}
	}

	err = output(configtx, filepath.Join(outputDir, outputFile), outputFormat, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func setOrgAsPartOfMultipleOps(configtx configtx.ConfigTx, instruction Instruction) error {
	// Parse command specific parameters
	tmp, err := yaml.Marshal(instruction.Parameters)
	if err != nil {
		return errors.Wrap(err, "failed to marshal instruction parameters")
	}
	params := SetOrgParameters{}
	err = yaml.Unmarshal([]byte(tmp), &params)
	if err != nil {
		return errors.Wrap(err, "failed to extract set-org parameters")
	}
	if params.OrgType == "" {
		params.OrgType = ApplicationOrg
	}
	printDebug("SetOrgParameters", params)

	// Do instruction with parsed parameters
	if err = setOrg(configtx, params.Org, params.OrgType); err != nil {
		return errors.Wrap(err, "failed to do 'set-org' with the parsed parameters")
	}

	return nil
}

func removeOrgAsPartOfMultipleOps(configtx configtx.ConfigTx, instruction Instruction) error {
	// Parse command specific parameters
	tmp, err := yaml.Marshal(instruction.Parameters)
	if err != nil {
		return errors.Wrap(err, "failed to marshal instruction parameters")
	}
	params := RemoveOrgParameters{}
	err = yaml.Unmarshal([]byte(tmp), &params)
	if err != nil {
		return errors.Wrap(err, "failed to extract remove-org parameters")
	}
	if params.OrgType == "" {
		params.OrgType = ApplicationOrg
	}
	printDebug("RemoveOrgParameters", params)

	// Do instruction with parsed parameters
	if err = removeOrg(configtx, params.OrgName, params.OrgType); err != nil {
		return errors.Wrap(err, "failed to do 'remove-org' with the parsed parameters")
	}

	return nil
}

func setChannelAsPartOfMultipleOps(configtx configtx.ConfigTx, instruction Instruction) error {
	// Parse command specific parameters
	tmp, err := yaml.Marshal(instruction.Parameters)
	if err != nil {
		return errors.Wrap(err, "failed to marshal instruction parameters")
	}
	params := SetChannelParameters{}
	err = yaml.Unmarshal([]byte(tmp), &params)
	if err != nil {
		return errors.Wrap(err, "failed to extract set-channel parameters")
	}
	printDebug("SetChannelParameters", params)

	// Do instruction with parsed parameters
	if err = setChannel(configtx, params.Channel); err != nil {
		return errors.Wrap(err, "failed to do 'set-channel' with the parsed parameters")
	}

	return nil
}

func setOrdererAsPartOfMultipleOps(configtx configtx.ConfigTx, instruction Instruction) error {
	// Parse command specific parameters
	tmp, err := yaml.Marshal(instruction.Parameters)
	if err != nil {
		return errors.Wrap(err, "failed to marshal instruction parameters")
	}
	params := SetOrdererParameters{}
	err = yaml.Unmarshal([]byte(tmp), &params)
	if err != nil {
		return errors.Wrap(err, "failed to extract set-orderer parameters")
	}
	printDebug("SetOrdererParameters", params)

	// Do instruction with parsed parameters
	if err = setOrderer(configtx, params.Orderer); err != nil {
		return errors.Wrap(err, "failed to do 'set-orderer' with the parsed parameters")
	}

	return nil
}

func setConsenterAsPartOfMultipleOps(configtx configtx.ConfigTx, instruction Instruction) error {
	// Parse command specific parameters
	tmp, err := yaml.Marshal(instruction.Parameters)
	if err != nil {
		return errors.Wrap(err, "failed to marshal instruction parameters")
	}
	params := SetConsenterParameters{}
	err = yaml.Unmarshal([]byte(tmp), &params)
	if err != nil {
		return errors.Wrap(err, "failed to extract set-consenter parameters")
	}
	printDebug("SetConsenterParameters", params)

	// Do instruction with parsed parameters
	if err = setConsenter(configtx, params.Consenter); err != nil {
		return errors.Wrap(err, "failed to do 'set-consenter' with the parsed parameters")
	}

	return nil
}

func removeConsenterAsPartOfMultipleOps(configtx configtx.ConfigTx, instruction Instruction) error {
	// Parse command specific parameters
	tmp, err := yaml.Marshal(instruction.Parameters)
	if err != nil {
		return errors.Wrap(err, "failed to marshal instruction parameters")
	}
	params := RemoveConsenterParameters{}
	err = yaml.Unmarshal([]byte(tmp), &params)
	if err != nil {
		return errors.Wrap(err, "failed to extract remove-consenter parameters")
	}
	printDebug("RemoveConsenterParameters", params)

	// Do instruction with parsed parameters
	if err = removeConsenter(configtx, params.ConsenterAddress); err != nil {
		return errors.Wrap(err, "failed to do 'remove-consenter' with the parsed parameters")
	}

	return nil
}
