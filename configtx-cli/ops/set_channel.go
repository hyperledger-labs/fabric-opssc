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

// OutputConfigTXToSetChannel creates ConfigUpdate for updating the channel configuration of the specified channel by using the specified profile and
// then outputs the ConfigUpdate to the specified output path with the specified output format.
func OutputConfigTXToSetChannel(blockPath string, outputDir string, outputFile string, outputFormat string, channelID string, profilePath string) error {
	configtx, err := getConfigTxFromBlock(blockPath)
	if err != nil {
		return errors.Wrap(err, "failed to get configtx from block")
	}

	profile, err := loadChannelProfile(profilePath)
	if err != nil {
		return errors.Wrap(err, "failed to load channel profile")
	}

	err = setChannel(configtx, profile)
	if err != nil {
		return errors.Wrap(err, "failed to set the channel configuration to the current configuration")
	}

	err = output(configtx, filepath.Join(outputDir, outputFile), outputFormat, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func setChannel(c configtx.ConfigTx, newChannelProfile ChannelProfile) error {

	logger.Info("Set channel")
	// Validate channel section
	if newChannelProfile.Channel == nil {
		return errors.New("section 'Channel' is not found")
	}
	if newChannelProfile.Channel.Policies == nil || len(newChannelProfile.Channel.Policies) < 1 {
		return errors.New("channel policies are not found")
	}

	// Replace Channel Policies
	channelGroup := c.Channel()
	newChannelPolicies, err := newPolicies(newChannelProfile.Channel.Policies)
	if err != nil {
		return errors.Wrap(err, "failed to create channel policies data structure for configtx from the profile")
	}
	if err = channelGroup.SetPolicies(configtx.AdminsPolicyKey, newChannelPolicies); err != nil {
		return errors.Wrap(err, "failed to set the channel policies to the current configuration")
	}

	appGroup := c.Application()
	if reflect.DeepEqual(appGroup, &configtx.ApplicationGroup{}) {
		logger.Warn("application group is not found")
		return nil
	}

	// Validate application section
	if newChannelProfile.Application == nil {
		return errors.New("section 'Application' is not found")
	}
	if newChannelProfile.Application.Policies == nil || len(newChannelProfile.Application.Policies) < 1 {
		return errors.New("application policies are not found")
	}

	// Replace Application Policies
	newAppPolicies, err := newPolicies(newChannelProfile.Application.Policies)
	if err != nil {
		return errors.Wrap(err, "failed to create application policies data structure for configtx from the profile")
	}
	if err = appGroup.SetPolicies(configtx.AdminsPolicyKey, newAppPolicies); err != nil {
		return errors.Wrap(err, "failed to set the application policies to the current configuration")
	}

	// Replace ACLs
	if err = appGroup.SetACLs(newChannelProfile.Application.ACLs); err != nil {
		return errors.Wrap(err, "failed to set the application ACLs to the current configuration")
	}

	return nil
}
