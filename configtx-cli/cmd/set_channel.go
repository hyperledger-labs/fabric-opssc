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

// SetChannelCmd returns the cobra command for set-channel.
func SetChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-channel",
		Short: "Output configtx to set channel (excepting organization parts)",
		Long:  "Output a config transaction to update a specified channel (excepting organization parts)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputToSetChannel(); err != nil {
				return err
			}
			if err := ops.OutputConfigTXToSetChannel(blockPath, outputDirectory, outputFile, outputFormat, channelID, profile); err != nil {
				return errors.Wrap(err, "failed to output configtx to set channel")
			}
			return nil
		},
	}

	flagList := []string{
		"channelID",
		"blockPath",
		"outputDir",
		"outputFile",
		"outputFormat",
		"profile",
	}
	attachFlags(cmd, flagList)

	return cmd
}

func validateInputToSetChannel() error {
	if channelID == "" {
		return errors.New("the required parameter 'channelID' is empty")
	}
	if blockPath == "" {
		return errors.New("the required parameter 'blockPath' is empty")
	}
	if profile == "" {
		return errors.New("the required parameter 'profile' is empty")
	}
	return nil
}
