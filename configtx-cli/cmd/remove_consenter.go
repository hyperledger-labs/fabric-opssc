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

// RemoveConsenterCmd returns the cobra command for remove-consenter.
func RemoveConsenterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-consenter",
		Short: "Output configtx to remove consenter",
		Long:  "Output a config transaction to remove a specified consenter (Currently a consenter with the specified address (hostname:port) is identified as the same)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputToRemoveConsenter(); err != nil {
				return err
			}
			logger.Infof("outputFormat: %s", outputFormat)
			if err := ops.OutputConfigTXToRemoveConsenter(blockPath, outputDirectory, outputFile, outputFormat, channelID, consenterAddress); err != nil {
				return errors.Wrap(err, "failed to output configtx to remove consenter")
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
		"consenterAddress",
	}
	attachFlags(cmd, flagList)

	return cmd
}

func validateInputToRemoveConsenter() error {
	if channelID == "" {
		return errors.New("the required parameter 'channelID' is empty")
	}
	if blockPath == "" {
		return errors.New("the required parameter 'blockPath' is empty")
	}
	if consenterAddress == "" {
		return errors.New("the required parameter 'consenterAddress' is empty")
	}
	return nil
}
