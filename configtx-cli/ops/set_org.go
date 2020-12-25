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

// OutputConfigTXToSetOrg creates ConfigUpdate for upserting the consenter to the specified channel by using the specified profile and
// then outputs the ConfigUpdate to the specified output path with the specified output format.
func OutputConfigTXToSetOrg(blockPath string, outputDir string, outputFile string, outputFormat string, channelID string, profilePath string, orgType string) error {
	configtx, err := getConfigTxFromBlock(blockPath)
	if err != nil {
		return errors.Wrap(err, "failed to get configtx from block")
	}

	profile, err := loadOrganizationProfile(profilePath)
	if err != nil {
		return errors.Wrap(err, "failed to load organization profile")
	}

	err = setOrg(configtx, profile, orgType)
	if err != nil {
		return errors.Wrap(err, "failed to set the organization to the current configuration")
	}

	err = output(configtx, filepath.Join(outputDir, outputFile), outputFormat, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func setOrg(c configtx.ConfigTx, newOrgProfile OrganizationProfile, orgType string) error {

	org, err := getOrganizationFromProfile(newOrgProfile)
	if err != nil {
		return errors.Wrap(err, "failed to get the organization data structure for configtx from the profile")
	}

	containCorrectType := false
	if strings.Contains(orgType, ApplicationOrg) {
		containCorrectType = true
		if err := setAppOrg(c, org); err != nil {
			return errors.Wrap(err, "failed to set the application organization")
		}
	}
	if strings.Contains(orgType, OrdererOrg) {
		containCorrectType = true
		if err := setOrdererOrg(c, org); err != nil {
			return errors.Wrap(err, "failed to set the orderer organization")
		}
	}
	if strings.Contains(orgType, ConsortiumsOrg) {
		containCorrectType = true
		if err := setConsortiumsOrg(c, org); err != nil {
			return errors.Wrap(err, "failed to set the consortiums organization")
		}
	}

	if !containCorrectType {
		return errors.New("correct orgType is not found")
	}

	return nil
}

func setAppOrg(c configtx.ConfigTx, org configtx.Organization) error {
	logger.Info("Set application organization")
	application := c.Application()
	if application == nil || reflect.DeepEqual(application, &configtx.ApplicationGroup{}) {
		return errors.New("the data 'Application' is not found in the current configuration")
	}

	logger.Infof("set target: %s", org.Name)
	if err := application.SetOrganization(org); err != nil {
		return errors.Wrap(err, "failed to set the organization as application organization")
	}
	return nil
}

func setOrdererOrg(c configtx.ConfigTx, org configtx.Organization) error {
	logger.Info("set orderer organization")
	orderer := c.Orderer()
	if orderer == nil || reflect.DeepEqual(orderer, &configtx.OrdererGroup{}) {
		return errors.New("the data 'Orderer' are not found in the current configuration")
	}

	logger.Infof("set target: %s", org.Name)
	if err := orderer.SetOrganization(org); err != nil {
		return errors.Wrap(err, "failed to set the organization as orderer organization")
	}
	return nil
}

func setConsortiumsOrg(c configtx.ConfigTx, org configtx.Organization) error {
	logger.Info("set consortiums organization")
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
		logger.Infof("set org '%s' to consortium '%s'", org.Name, consortiumConfig.Name)
		if err := consortium.SetOrganization(org); err != nil {
			return errors.Wrap(err, "failed to set the organization to the first consortium")
		}
	}

	return nil
}
