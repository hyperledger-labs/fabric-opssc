# fabric-configtx-cli

## Overview

This is a tiny channel / organization management CLI tool for Hyperledger Fabric v2.x.
- This tool outputs a new config transaction that controls channel and organization with the following inputs:
  - The latest config block
  - `configtx.yaml`-like profile
  - Parameters to the CLI
- This tool internally uses [Config Transaction Library](https://github.com/hyperledger/fabric-config).
- This tool is internally used by OpsSC agents and API servers.

## Prerequisites / Assumptions

- go 1.14
- Hyperledger Fabric 2.2.1 or later

## How to use

### Setup

```sh
$ make build

$ ls bin/fabric-configtx-cli
bin/fabric-configtx-cli
```

### Commands

`fabric-configtx-cli` provides subcommands to control channel configurations.

The subcommands are:

```sh
$ bin/fabric-configtx-cli
Usage:
  fabric-configtx-cli [command]

Available Commands:
  create-channel       Output configtx to create channel
  create-envelope      Output an enveloped configtx with ConfigUpdate and ConfigSignitures
  execute-multiple-ops Output configtx to do multiple operations on a channel
  help                 Help about any command
  remove-consenter     Output configtx to remove consenter
  remove-org           Output configtx to remove org
  set-channel          Output configtx to set channel (excepting organization parts)
  set-consenter        Output configtx to set consenter
  set-orderer          Output configtx to update orderer configuration (excepting consenters)
  set-org              Output configtx to set org
  sign                 Output ConfigSignature to sign the specified configTx

Flags:
  -h, --help   help for fabric-configtx-cli

Use "fabric-configtx-cli [command] --help" for more information about a command.
```

Each subcommand is explained below:

#### fabric-configtx-cli set-org

This subcommand outputs a config transaction to set a specified organization.
If the organization is not exist, the subcommand creates a config transaction to add the organization.
If it is exist, the subcommand creates a config transaction to update the organization.

```sh
$ bin/fabric-configtx-cli set-org --help
Output a config transaction to set (add or update) a specified organization (Currently support for application, orderer, and consortium organization)

Usage:
  fabric-configtx-cli set-org [flags]

Flags:
      --blockPath string      The path to read the config block
  -C, --channelID string      The channel ID to use in the configtx
  -h, --help                  help for set-org
      --orgType string        The organization type which is operated (Option: Application, Orderer, Consortiums) (default "Application")
      --outputDir string      The path to write the configtx (default "artifacts")
      --outputFile string     The file name to write the configtx (default "output.pb")
      --outputFormat string   The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta') (default "enveloped_delta")
      --profile string        configtx.yaml-like profile to control configtx (the format depends on subcommands)
```

The subcommand expects a profile with almost the same format as the Organizations section of `configtx.yaml`.
For more information on the format, see [a sample profile to set an org (Org3MSP) with reading MSP Dir](./ops/testdata/org3_profile.yaml) and [a sample profile to set an org (Org3MSP) without reading MSP Dir](./ops/testdata/org3_profile_without_reading_mspdir.yaml).

#### fabric-configtx-cli remove-org

```sh
$ bin/fabric-configtx-cli remove-org --help
Output a config transaction to remove a specified organization (Currently support for application, orderer, and consortium organization)

Usage:
  fabric-configtx-cli remove-org [flags]

Flags:
      --blockPath string      The path to read the config block
  -C, --channelID string      The channel ID to use in the configtx
  -h, --help                  help for remove-org
      --orgName string        The organization name (OrgName) which is operated
      --orgType string        The organization type which is operated (Option: Application, Orderer, Consortiums) (default "Application")
      --outputDir string      The path to write the configtx (default "artifacts")
      --outputFile string     The file name to write the configtx (default "output.pb")
      --outputFormat string   The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta') (default "enveloped_delta")
```

#### fabric-configtx-cli set-consenter

```sh
$ bin/fabric-configtx-cli set-consenter --help
Output a config transaction to set (add or update) a specified consenter (Currently a consenter with the specified address is identified as the same)

Usage:
  fabric-configtx-cli set-consenter [flags]

Flags:
      --blockPath string      The path to read the config block
  -C, --channelID string      The channel ID to use in the configtx
  -h, --help                  help for set-consenter
      --outputDir string      The path to write the configtx (default "artifacts")
      --outputFile string     The file name to write the configtx (default "output.pb")
      --outputFormat string   The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta') (default "enveloped_delta")
      --profile string        configtx.yaml-like profile to control configtx (the format depends on subcommands)
```

The subcommand expects a profile with almost the same format as the Consenters parameters in the Orderer section of `configtx.yaml`.
For more information on the format, see [a sample profile to set a consenter (orderer0.org3.example.com) with reading cert files](./ops/testdata/org3_consenter_profile.yaml) and [a sample profile to set a consenter (orderer0.org3.example.com) without reading cert files](./ops/testdata/org3_consenter_profile_without_reading_certs.yaml).

#### fabric-configtx-cli remove-consenter

```sh
$ bin/fabric-configtx-cli remove-consenter --help
Output a config transaction to remove a specified consenter (Currently a consenter with the specified address is identified as the same)

Usage:
  fabric-configtx-cli remove-consenter [flags]

Flags:
      --blockPath string          The path to read the config block
  -C, --channelID string          The channel ID to use in the configtx
      --consenterAddress string   The consenter address should be removed (e.g.,: 'orderer.example.com:7050')
  -h, --help                      help for remove-consenter
      --outputDir string          The path to write the configtx (default "artifacts")
      --outputFile string         The file name to write the configtx (default "output.pb")
      --outputFormat string       The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta') (default "enveloped_delta")
```

#### fabric-configtx-cli set-orderer

```sh
$ bin/fabric-configtx-cli set-orderer --help
Output a config transaction to update orderer configuration for a channel
(NOTE: Excepting consenters settings: set/remove-consenter should be used.
       Only support for Raft orderer.)

Usage:
  fabric-configtx-cli set-orderer [flags]

Flags:
      --blockPath string      The path to read the config block
  -C, --channelID string      The channel ID to use in the configtx
  -h, --help                  help for set-orderer
      --outputDir string      The path to write the configtx (default "artifacts")
      --outputFile string     The file name to write the configtx (default "output.pb")
      --outputFormat string   The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta') (default "enveloped_delta")
      --profile string        configtx.yaml-like profile to control configtx (the format depends on subcommands)
```

The subcommand expects a profile with almost the same format as the Orderer section excepting the Consenters parameters of `configtx.yaml`.
For more information on the format, see [a sample profile to update orderer configuration](./ops/testdata/updated_orderer_profile.yaml).

#### fabric-configtx-cli set-channel

```sh
$ bin/fabric-configtx-cli set-channel --help
Output a config transaction to update a specified channel (excepting organization parts)

Usage:
  fabric-configtx-cli set-channel [flags]

Flags:
      --blockPath string      The path to read the config block
  -C, --channelID string      The channel ID to use in the configtx
  -h, --help                  help for set-channel
      --outputDir string      The path to write the configtx (default "artifacts")
      --outputFile string     The file name to write the configtx (default "output.pb")
      --outputFormat string   The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta') (default "enveloped_delta")
      --profile string        configtx.yaml-like profile to control configtx (the format depends on subcommands)
```

The subcommand expects a profile with almost the same format as the Application and Channel sections excepting the Organizations parameters of `configtx.yaml`.
For more information on the format, see [a sample profile to update channel configuration](./ops/testdata/updated_mychannel_profile.yaml).

#### fabric-configtx-cli execute-multiple-ops

```sh
$ bin/fabric-configtx-cli execute-multiple-ops --help
Output a config transaction to do multiple operations (like set-org and set-channel) on a specified channel

Usage:
  fabric-configtx-cli execute-multiple-ops [flags]

Flags:
      --blockPath string      The path to read the config block
  -C, --channelID string      The channel ID to use in the configtx
  -h, --help                  help for execute-multiple-ops
      --outputDir string      The path to write the configtx (default "artifacts")
      --outputFile string     The file name to write the configtx (default "output.pb")
      --outputFormat string   The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta') (default "enveloped_delta")
      --profile string        configtx.yaml-like profile to control configtx (the format depends on subcommands)
```

The subcommand expects a profile with almost the same format as profiles for the other subcommands for update merged.
For more information on the format, see [a sample profile to update an application channel with reading the other files](./ops/testdata/multiple_ops_profile_for_mychannel.yaml) and [without reading the other files](./ops/testdata/multiple_ops_profile_for_mychannel_without_reading_other_files.yaml) and [a sample profile to update the system channel with reading the other files](./ops/testdata/multiple_ops_profile_for_system-channel.yaml).

#### fabric-configtx-cli create-channel

```sh
$ bin/fabric-configtx-cli create-channel --help
Output a config transaction to create a channel

Usage:
  fabric-configtx-cli create-channel [flags]

Flags:
  -C, --channelID string      The channel ID to use in the configtx
  -h, --help                  help for create-channel
      --outputDir string      The path to write the configtx (default "artifacts")
      --outputFile string     The file name to write the configtx (default "output.pb")
      --outputFormat string   The output format for ConfigUpdate (Option: 'delta' or 'enveloped_delta') (default "enveloped_delta")
      --profile string        configtx.yaml-like profile to control configtx (the format depends on subcommands)
```

The subcommand expects creating a channel based on an existing consortium. Also, it expects a profile with almost the same format as the Application section of `configtx.yaml`.
For more information on the format, see [a sample profile to create an application channel](./ops/testdata/create_mychannel2_profile.yaml).

#### fabric-configtx-cli sign

```sh
$ bin/fabric-configtx-cli sign --help
Output ConfigSignature to sign the specified configTx with the specified key/cert

Usage:
  fabric-configtx-cli sign [flags]

Flags:
      --certPath string       The path to read the certificate to be used to sign ConfigUpdate
      --configTxPath string   The path to read the ConfigUpdate (assuming delta not enveloped)
  -h, --help                  help for sign
      --keyPath string        The path to read the key to be used to sign ConfigUpdate
      --mspID string          MSP ID to be used to sign ConfigUpdate
      --outputDir string      The path to write the configtx (default "artifacts")
      --outputFile string     The file name to write the configtx (default "output.pb")
```

#### fabric-configtx-cli create-envelope

```sh
$ bin/fabric-configtx-cli create-envelope --help
Output an enveloped configuration transaction with ConfigUpdate and ConfigSignatures specified in the profile

Usage:
  fabric-configtx-cli create-envelope [flags]

Flags:
  -h, --help                help for create-envelope
      --outputDir string    The path to write the configtx (default "artifacts")
      --outputFile string   The file name to write the configtx (default "output.pb")
      --profile string      configtx.yaml-like profile to control configtx (the format depends on subcommands)
```

The subcommand expects a profile with base64 encoded `configUpdate` and `signature` by each organizations.
For more information on the format, see [a sample profile to create an envelope](./ops/testdata/configtx_profile.yaml).
This format is compatible with the `Artifacts` in `Proposal` in `channel-ops` Ops chaincode.

### Examples of how to use

See the test code for more information on how to use this command.

### Logging control

Logging in the tool is provided by the `hyperledger/fabric/common/flogging` package.
The logging level of the commands is controlled by a logging specification, which is set via the `FABRIC_LOGGING_SPEC` environment variable.

So, for example, you can change the logging level when running the commands as follows:

```sh
$ FABRIC_LOGGING_SPEC=DEBUG bin/fabric-configtx-cli xxxxx
(...) // This tool outputs debug logs to help to know the internal behaviors
2020-07-30 18:46:45.622 UTC [fabric-configtx-cli.common] PrintDebug -> DEBU 007 Debug: Output envelope
2020-07-30 18:46:45.622 UTC [fabric-configtx-cli.common] PrintDebug -> DEBU 008 {"payload": ...
(...)
```

Logging severity levels are specified using case-insensitive strings chosen from:

```
   FATAL | PANIC | ERROR | WARNING | INFO | DEBUG
```

See [Hyperledger Fabric >> Operations Guides >> Logging Control](https://hyperledger-fabric.readthedocs.io/en/latest/logging-control.html) for more detail.

## Design memo: Current scope

This tool focuses on major operations to *update* channels.

- Scope:
  - Update channels (both application and system)
    - Add / remove organizations on a channel
    - Add / remove etcdraft consenters
    - Update organization configuration (e.g., MSP settings, policies, anchor peers, orderer endpoints)
    - Update consenter configuration (e.g., address, TLS certificates)
    - Update channel configuration (e.g., policies)
    - Update orderer configuration (e.g., policies, block size)
  - Create application channels
- Out of scope:
  - Create system genesis block
    - Assume using `configtxgen` for initialization
  - Fetch the latest config blocks
    - Assume using `peer channel fetch config`
  - Minor operations
    - Add an consortium, update consortium name (the tool only supports add/remove consortium org)
  - Setting Kafka / solo orderers (these types of orderers are deprecated)

## TODO

- Improve test coverage
- Validation improvements (on both CLI parameters and profiles)
- Support channel management without the system channel
