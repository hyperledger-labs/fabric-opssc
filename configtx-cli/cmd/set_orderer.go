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

// SetOrdererCmd returns the cobra command for set-orderer.
func SetOrdererCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-orderer",
		Short: "Output configtx to update orderer configuration (excepting consenters)",
		Long:  "Output a config transaction to update orderer configuration for a channel\n(NOTE: Excepting consenters settings: set/remove-consenter should be used.\n       Only support for Raft orderer.)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputToSetOrderer(); err != nil {
				return err
			}
			if err := ops.OutputConfigTXToSetOrderer(blockPath, outputDirectory, outputFile, outputFormat, channelID, profile); err != nil {
				return errors.Wrap(err, "failed to output configtx to set orderer")
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

func validateInputToSetOrderer() error {
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
