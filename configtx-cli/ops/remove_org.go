/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-config/configtx"
)

// OutputConfigTXToRemoveOrg creates ConfigUpdate for removing the specified organization from the specified channel and
// then outputs the ConfigUpdate to the specified output path with the specified output format.
func OutputConfigTXToRemoveOrg(blockPath string, outputDir string, outputFile string, outputFormat string, channelID string, orgName string, orgType string) error {
	configtx, err := getConfigTxFromBlock(blockPath)
	if err != nil {
		return errors.Wrap(err, "failed to get configtx from block")
	}

	err = removeOrg(configtx, orgName, orgType)
	if err != nil {
		return errors.Wrap(err, "failed to remove the organization from the current config")
	}

	err = output(configtx, filepath.Join(outputDir, outputFile), outputFormat, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func removeOrg(c configtx.ConfigTx, orgName string, orgType string) error {

	if strings.Contains(orgType, ApplicationOrg) {
		if err := removeAppOrg(c, orgName); err != nil {
			return errors.Wrap(err, "failed to remove the application organization")
		}
	}
	if strings.Contains(orgType, OrdererOrg) {
		if err := removeOrdererOrg(c, orgName); err != nil {
			return errors.Wrap(err, "failed to remove the orderer organization")
		}
	}
	if strings.Contains(orgType, ConsortiumsOrg) {
		if err := removeConsortiumsOrg(c, orgName); err != nil {
			return errors.Wrap(err, "failed to remove the consortiums organization")
		}
	}

	return nil
}

func removeAppOrg(c configtx.ConfigTx, orgName string) error {
	logger.Infof("remove application organization: %s", orgName)
	application := c.Application()
	if application == nil || reflect.DeepEqual(application, &configtx.ApplicationGroup{}) {
		return errors.New("the data 'Application' is not found in the current configuration")
	}

	appOrg := c.Application().Organization(orgName)
	if appOrg == nil {
		return errors.New("the target organization is not found in the current configuration")
	}
	application.RemoveOrganization(orgName)
	return nil
}

func removeOrdererOrg(c configtx.ConfigTx, orgName string) error {
	logger.Infof("remove orderer organization: %s", orgName)
	orderer := c.Orderer()
	if orderer == nil || reflect.DeepEqual(orderer, &configtx.OrdererGroup{}) {
		return errors.New("the data 'Orderer' are not found in the current configuration")
	}

	ordererOrg := c.Orderer().Organization(orgName)
	if ordererOrg == nil {
		return errors.New("the target organization is not found in the current configuration")
	}
	orderer.RemoveOrganization(orgName)
	return nil
}

func removeConsortiumsOrg(c configtx.ConfigTx, orgName string) error {
	logger.Infof("remove consortiums organization: %s", orgName)
	consortiums := c.Consortiums()
	if consortiums == nil || reflect.DeepEqual(consortiums, &configtx.ConsortiumsGroup{}) {
		return errors.New("the data 'Consortiums' are not found in the current configuration")
	}

	consortiumConfigs, err := consortiums.Configuration()
	if err != nil {
		return errors.Wrap(err, "failed to set the consortiums configuration")
	}

	for _, consortiumConfig := range consortiumConfigs {
		consortium := c.Consortium(consortiumConfig.Name)
		logger.Infof("remove org '%s' to consortium '%s'", orgName, consortiumConfig.Name)

		consortiumOrg := consortium.Organization(orgName)
		if consortiumOrg == nil {
			logger.Warn("the target organization is not found in the current configuration")
		}
		consortium.RemoveOrganization(orgName)
	}

	return nil
}
