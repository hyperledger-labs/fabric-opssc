/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/satota2/fabric-opssc/configtx-cli/ops"
)

// RemoveOrgCmd returns the cobra command for remove-org.
func RemoveOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-org",
		Short: "Output configtx to remove org",
		Long:  "Output a config transaction to remove a specified organization (Currently support for application, orderer, and consortium organization)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputForRemoveOrg(); err != nil {
				return err
			}
			if err := ops.OutputConfigTXToRemoveOrg(blockPath, outputDirectory, outputFile, outputFormat, channelID, orgName, orgType); err != nil {
				return errors.Wrap(err, "failed to output configtx to remove org")
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
		"orgName",
		"orgType",
	}
	attachFlags(cmd, flagList)

	return cmd
}

func validateInputForRemoveOrg() error {
	if channelID == "" {
		return errors.New("the required parameter 'channelID' is empty")
	}
	if blockPath == "" {
		return errors.New("the required parameter 'blockPath' is empty")
	}
	if orgName == "" {
		return errors.New("the required parameter 'orgName' is empty")
	}
	return nil
}
