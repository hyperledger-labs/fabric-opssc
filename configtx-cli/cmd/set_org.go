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

// SetOrgCmd returns the cobra command for set-org.
func SetOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-org",
		Short: "Output configtx to set org",
		Long:  "Output a config transaction to set (add or update) a specified organization (Currently support for application, orderer, and consortium organization)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputToSetOrg(); err != nil {
				return err
			}
			if err := ops.OutputConfigTXToSetOrg(blockPath, outputDirectory, outputFile, outputFormat, channelID, profile, orgType); err != nil {
				return errors.Wrap(err, "failed to output configtx to set org")
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
		"orgType",
	}
	attachFlags(cmd, flagList)

	return cmd
}

func validateInputToSetOrg() error {
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
