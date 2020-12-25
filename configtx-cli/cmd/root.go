/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var logger = flogging.MustGetLogger("fabric-configtx-cli.cmd")

var flags *pflag.FlagSet

var (
	outputDirectory  string
	outputFile       string
	profile          string
	blockPath        string
	channelID        string
	orgName          string
	orgType          string
	consenterAddress string

	outputFormat string

	mspID        string
	keyPath      string
	certPath     string
	configTxPath string
)

// NewRootCmd returns the cobra command for fabric-configtx-cli root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fabric-configtx-cli",
	}

	resetFlags()
	cmd.AddCommand(SetOrgCmd())
	cmd.AddCommand(RemoveOrgCmd())
	cmd.AddCommand(SetConsenterCmd())
	cmd.AddCommand(RemoveConsenterCmd())
	cmd.AddCommand(SetChannelCmd())
	cmd.AddCommand(SetOrdererCmd())
	cmd.AddCommand(MultipleOpsCmd())
	cmd.AddCommand(CreateChannelCmd())
	cmd.AddCommand(SignCmd())
	cmd.AddCommand(CreateEnvelopeCmd())

	return cmd
}

// Execute is called from main
func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		logger.Error(fmt.Errorf("fail to execute command: %v", err))
		os.Exit(1)
	}
}

func resetFlags() {
	flags = &pflag.FlagSet{}
	flags.StringVar(&blockPath, "blockPath", "", "The path to read the config block")
	flags.StringVar(&outputDirectory, "outputDir", "artifacts", "The path to write the configtx")
	flags.StringVar(&outputFile, "outputFile", "output.pb", "The file name to write the configtx")
	flags.StringVarP(&channelID, "channelID", "C", "", "The channel ID to use in the configtx")
	flags.StringVar(&profile, "profile", "", "configtx.yaml-like profile to control configtx (the format depends on subcommands)")
	flags.StringVar(&orgName, "orgName", "", "The organization name (OrgName) which is operated")
	flags.StringVar(&orgType, "orgType", "Application", "The organization type which is operated (Option: Application, Orderer, Consortiums)")
	flags.StringVar(&consenterAddress, "consenterAddress", "", "The consenter address should be removed (e.g.,: 'orderer.example.com:7050')")

	flags.StringVar(&outputFormat, "outputFormat", "enveloped_delta", "The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta')")

	flags.StringVar(&mspID, "mspID", "", "MSP ID to be used to sign ConfigUpdate")
	flags.StringVar(&keyPath, "keyPath", "", "The path to read the key to be used to sign ConfigUpdate")
	flags.StringVar(&certPath, "certPath", "", "The path to read the certificate to be used to sign ConfigUpdate")
	flags.StringVar(&configTxPath, "configTxPath", "", "The path to read the ConfigUpdate (assuming delta not enveloped)")
}

func init() {
}

func attachFlags(cmd *cobra.Command, names []string) {
	cmdFlags := cmd.Flags()
	for _, name := range names {
		if flag := flags.Lookup(name); flag != nil {
			cmdFlags.AddFlag(flag)
		} else {
			logger.Fatalf("Could not find flag '%s' to attach to command '%s'", name, cmd.Name())
		}
	}
}
