/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-config/configtx"

	"github.com/hyperledger/fabric/protoutil"
)

// OutputConfigTXToCreateChannel creates ConfigUpdate for creating the specified channel by using the specified profile and
// then outputs the ConfigUpdate to the specified output path with the specified output format.
func OutputConfigTXToCreateChannel(outputDir string, outputFile string, outputFormat string, channelID string, profilePath string) error {

	profile, err := loadChannelCreationProfile(profilePath)
	if err != nil {
		return errors.Wrap(err, "failed to load channel creation profile")
	}

	channel, err := createChannelConfig(profile)
	if err != nil {
		return errors.Wrap(err, "failed to create the channel configuration")
	}

	err = outputForChannelCreation(filepath.Join(outputDir, outputFile), outputFormat, channel, channelID)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func createChannelConfig(profile ChannelCreationProfile) (configtx.Channel, error) {

	logger.Info("create channel configuration")

	// Validate profile
	if profile.Consortium == "" {
		return configtx.Channel{}, errors.New("section 'Consortium' should be set")
	}
	if profile.Application == nil {
		return configtx.Channel{}, errors.New("section 'Application' is not found")
	}
	if profile.Application.Policies == nil || len(profile.Application.Policies) < 1 {
		return configtx.Channel{}, errors.New("application policies are not found")
	}

	// Create channel configuration from profile
	policies, err := newPolicies(profile.Application.Policies)
	if err != nil {
		return configtx.Channel{}, err
	}
	channel := configtx.Channel{
		Consortium: profile.Consortium,
		Application: configtx.Application{
			Organizations: newOrganizations(profile.Application.Organizations),
			ACLs:          profile.Application.ACLs,
			Policies:      policies,
			Capabilities:  profile.Application.Capabilities,
		},
	}

	return channel, nil
}

func outputForChannelCreation(outputConfigUpdatePath string, outputFormat string, channel configtx.Channel, channelName string) error {
	logger.Info("output configtx to file")
	delta, err := configtx.NewMarshaledCreateChannelTx(channel, channelName)
	if err != nil {
		return errors.Wrap(err, "failed to create the marshaled CreateChannelTx")
	}
	printDebug("output delta", delta)

	var outputData []byte
	switch outputFormat {
	case "delta":
		outputData = delta
	case "enveloped_delta":
		env, err := configtx.NewEnvelope(delta)
		if err != nil {
			return errors.Wrap(err, "failed to create the enveloped delta")
		}
		printDebug("output envelope", env)
		outputData = protoutil.MarshalOrPanic(env)
	default:
		return errors.New("invalid output format. It should be set 'delta' or 'enveloped_delta'")
	}

	// Write to file
	err = writeFile(outputConfigUpdatePath, outputData, 0640)
	if err != nil {
		return errors.Wrap(err, "failed to write the output file")
	}

	return nil
}
