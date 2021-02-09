/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/hyperledger-labs/fabric-opssc/configtx-cli/ops"
)

// CreateChannelCmd returns the cobra command for create-channel.
func CreateChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-channel",
		Short: "Output configtx to create channel",
		Long:  "Output a config transaction to create a channel",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputToCreateChannel(); err != nil {
				return err
			}
			if err := ops.OutputConfigTXToCreateChannel(outputDirectory, outputFile, outputFormat, channelID, profile); err != nil {
				return errors.Wrap(err, "failed to output configtx to create channel")
			}
			return nil
		},
	}

	flagList := []string{
		"channelID",
		"outputDir",
		"outputFile",
		"outputFormat",
		"profile",
	}
	attachFlags(cmd, flagList)

	return cmd
}

func validateInputToCreateChannel() error {
	if channelID == "" {
		return errors.New("the required parameter 'channelID' is empty")
	}
	if profile == "" {
		return errors.New("the required parameter 'profile' is empty")
	}
	return nil
}
