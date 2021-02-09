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

// MultipleOpsCmd returns the cobra command for execute-multiple-ops.
func MultipleOpsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-multiple-ops",
		Short: "Output configtx to do multiple operations on a channel",
		Long:  "Output a config transaction to do multiple operations (like set-org and set-channel) on a specified channel",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputForMultipleOps(); err != nil {
				return err
			}
			if err := ops.OutputConfigTXToDoMultipleOps(blockPath, outputDirectory, outputFile, outputFormat, channelID, profile); err != nil {
				return errors.Wrap(err, "failed to output configtx to execute multiple ops")
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

func validateInputForMultipleOps() error {
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
