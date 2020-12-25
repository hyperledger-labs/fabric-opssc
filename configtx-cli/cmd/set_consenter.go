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

// SetConsenterCmd returns the cobra command for set-consenter.
func SetConsenterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-consenter",
		Short: "Output configtx to set consenter",
		Long:  "Output a config transaction to set (add or update) a specified consenter (Currently a consenter with the specified address is identified as the same)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputToSetConsenter(); err != nil {
				return err
			}
			if err := ops.OutputConfigTXToSetConsenter(blockPath, outputDirectory, outputFile, outputFormat, channelID, profile); err != nil {
				return errors.Wrap(err, "failed to output configtx to set consenter")
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

func validateInputToSetConsenter() error {
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
