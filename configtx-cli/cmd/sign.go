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

// SignCmd returns the cobra command for sign.
func SignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Output ConfigSignature to sign the specified configTx",
		Long:  "Output ConfigSignature to sign the specified configTx with the specified key/cert",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputToSign(); err != nil {
				return err
			}
			if err := ops.OutputSign(configTxPath, outputDirectory, outputFile, mspID, keyPath, certPath); err != nil {
				return errors.Wrap(err, "failed to output configtx to sign")
			}
			return nil
		},
	}

	flagList := []string{
		"configTxPath",
		"mspID",
		"keyPath",
		"certPath",
		"outputDir",
		"outputFile",
	}
	attachFlags(cmd, flagList)

	return cmd
}

func validateInputToSign() error {
	if mspID == "" {
		return errors.New("the required parameter 'mspID' is empty")
	}
	if keyPath == "" {
		return errors.New("the required parameter 'keyPath' is empty")
	}
	if certPath == "" {
		return errors.New("the required parameter 'certPath' is empty")
	}
	if configTxPath == "" {
		return errors.New("the required parameter 'configTxPath' is empty")
	}
	return nil
}
