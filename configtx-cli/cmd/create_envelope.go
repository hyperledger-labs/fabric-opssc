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

// CreateEnvelopeCmd returns the cobra command for create-envelope.
func CreateEnvelopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-envelope",
		Short: "Output an enveloped configtx with ConfigUpdate and ConfigSignitures",
		Long:  "Output an enveloped configuration transaction with ConfigUpdate and ConfigSignatures specified in the profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateInputToCreateEnvelope(); err != nil {
				return err
			}
			if err := ops.OutputEnvelopedConfigTx(profile, outputDirectory, outputFile); err != nil {
				return errors.Wrap(err, "failed to output configtx to create envelope")
			}
			return nil
		},
	}

	flagList := []string{
		"profile",
		"outputDir",
		"outputFile",
	}
	attachFlags(cmd, flagList)

	return cmd
}

func validateInputToCreateEnvelope() error {
	if profile == "" {
		return errors.New("the required parameter 'profile' is empty")
	}
	return nil
}
