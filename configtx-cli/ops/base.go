/*
Copyright IBM Corp. All Rights Reserved.

Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/configtx"
	"github.com/hyperledger/fabric-config/configtx/orderer"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("fabric-configtx-cli.ops")

const (
	// ApplicationOrg identifies that OrgType is application organization.
	ApplicationOrg = "Application"

	// OrdererOrg identifies that OrgType is application organization.
	OrdererOrg = "Orderer"

	// ConsortiumsOrg identifies that OrgType is consortiums organization.
	ConsortiumsOrg = "Consortiums"
)

// getConfigTxFromBlock returns ConfigTx object in fabric-config from the specified config block file.
func getConfigTxFromBlock(path string) (configtx.ConfigTx, error) {
	baseConfig, err := getConfigFromBlock(path)
	if err != nil {
		return configtx.ConfigTx{}, errors.Wrap(err, "failed to get config from the specified file")
	}
	c := configtx.New(baseConfig)
	return c, nil
}

func getConfigFromBlock(blockPath string) (*cb.Config, error) {
	logger.Info("getting config from block")
	blockBin, err := ioutil.ReadFile(blockPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the specified file")
	}

	block := &cb.Block{}
	err = proto.Unmarshal(blockBin, block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the config block")
	}

	blockDataEnvelope := &cb.Envelope{}
	err = proto.Unmarshal(block.Data.Data[0], blockDataEnvelope)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the data envelope in the config block")
	}

	blockDataPayload := &cb.Payload{}
	err = proto.Unmarshal(blockDataEnvelope.Payload, blockDataPayload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the block data payload in the config block")
	}

	config := &cb.ConfigEnvelope{}
	err = proto.Unmarshal(blockDataPayload.Data, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the config envelope in the config block")
	}

	return config.Config, nil
}

// getConfigUpdateFromFile returns ConfigUpdate from the specified file.
func getConfigUpdateFromFile(filePath string) (*cb.ConfigUpdate, error) {
	bin, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the specified file")
	}
	var update cb.ConfigUpdate
	if err := proto.Unmarshal(bin, &update); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the ConfigUpdate")
	}

	return &update, nil
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(filename)
	exists, err := dirExists(dirPath)
	if err != nil {
		return errors.Wrap(err, "failed to check dir exists")
	}
	if !exists {
		err = os.MkdirAll(dirPath, 0750)
		if err != nil {
			return errors.Wrap(err, "failed to mkdir for outputting to the specified path")
		}
	}
	if err = ioutil.WriteFile(filename, data, perm); err != nil {
		return errors.Wrap(err, "failed to output")
	}
	return nil
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getOrganizationFromProfile(profile OrganizationProfile) (configtx.Organization, error) {
	printDebug("profile", profile)
	if (profile.MSPDir != "" && !reflect.DeepEqual(profile.MSP, MSP{})) {
		return configtx.Organization{}, errors.New("multiple MSP settings are not accepted. Only MSPDir or MSP should be set")
	}

	var mspConfig configtx.MSP
	var err error

	if profile.MSPDir != "" {
		logger.Info("load MSP Config from MSP Dir")
		mspConfig, err = getMSPConfigFromDir(profile.MSPDir, profile.ID)
		if err != nil {
			return configtx.Organization{}, errors.Wrap(err, "failed to get MSP config from MSP dir in the organization profile")
		}
	} else {
		logger.Info("create MSP Config with MSP section")
		mspConfig, err = createMSPConfig(profile.ID, profile.MSP)
		if err != nil {
			return configtx.Organization{}, errors.Wrap(err, "failed to create MSP config with MSP section in the organization profile")
		}
	}

	policies, err := newPolicies(profile.Policies)
	if err != nil {
		return configtx.Organization{}, errors.Wrap(err, "failed to create polices")
	}

	org := configtx.Organization{
		Name:             profile.Name,
		MSP:              mspConfig,
		Policies:         policies,
		AnchorPeers:      newAnchorPeers(profile.AnchorPeers),
		OrdererEndpoints: profile.OrdererEndpoints,
	}

	return org, nil
}

// output outputs ConfigUpdate
func output(c configtx.ConfigTx, outputConfigUpdatePath string, outputFormat string, channelName string) error {
	logger.Info("output configtx to file")
	delta, err := c.ComputeMarshaledUpdate(channelName)
	if err != nil {
		return errors.Wrap(err, "failed to compute marshaled config update")
	}
	printDebug("Output delta", delta)

	var outputData []byte
	switch outputFormat {
	case "delta":
		outputData = delta
	case "enveloped_delta":
		env, err := configtx.NewEnvelope(delta)
		if err != nil {
			return errors.Wrap(err, "failed to create envelope for the config update")
		}
		printDebug("Output envelope", env)
		outputData = protoutil.MarshalOrPanic(env)
	default:
		return errors.New("invalid output format. It should be set 'delta' or 'enveloped_delta'")
	}

	// Write to file
	err = writeFile(outputConfigUpdatePath, outputData, 0640)
	if err != nil {
		return errors.New("failed to output the config update to the file")
	}

	return nil
}

func printDebug(label string, targetObject interface{}) {
	logger.Debugf("Debug: %s", label)

	json, err := json.Marshal(targetObject)
	if err != nil {
		logger.Debug("Fail to marshal")
		logger.Debug(err)
	}
	logger.Debug(string(json))
}

func newPolicies(policies map[string]*Policy) (map[string]configtx.Policy, error) {
	txPolicies := map[string]configtx.Policy{}
	for name, policy := range policies {
		if policy == nil {
			return txPolicies, errors.New(fmt.Sprintf("Policy %v has no definition", name))
		}
		txPolicies[name] = configtx.Policy{
			Type: policy.Type,
			Rule: policy.Rule,
		}
	}

	return txPolicies, nil
}

func newOrganizations(orgNames []string) []configtx.Organization {
	organizations := []configtx.Organization{}
	for _, orgName := range orgNames {
		organizations = append(organizations, configtx.Organization{
			Name: orgName,
		})
	}
	return organizations
}

func newAnchorPeers(anchorPeers []*AnchorPeer) []configtx.Address {
	txAnchorPeers := []configtx.Address{}
	for _, peer := range anchorPeers {
		txAnchorPeers = append(txAnchorPeers, configtx.Address{
			Host: peer.Host,
			Port: peer.Port,
		})
	}
	return txAnchorPeers
}

func newConsenter(consenter Consenter) (orderer.Consenter, error) {
	bytes, err := readPemFile(consenter.ClientTLSCert)
	if err != nil {
		bytes = []byte(consenter.ClientTLSCert)
	}
	clientCert, err := parseCertificateFromBytes(bytes)
	if err != nil {
		return orderer.Consenter{}, errors.New("ClientTLSCert is cannot read as PemFile and/or parse certificate")
	}

	bytes, err = readPemFile(consenter.ServerTLSCert)
	if err != nil {
		bytes = []byte(consenter.ServerTLSCert)
	}
	serverCert, err := parseCertificateFromBytes(bytes)
	if err != nil {
		return orderer.Consenter{}, errors.New("ServerTLSCert is cannot read as PemFile and/or parse certificate")
	}

	txConsenter := orderer.Consenter{
		Address: orderer.EtcdAddress{
			Host: consenter.Host,
			Port: int(consenter.Port),
		},
		ClientTLSCert: clientCert,
		ServerTLSCert: serverCert,
	}
	return txConsenter, nil
}

// This function is based on byteSizeDecodeHook() in common/viperutil/config_util.go
func decodeByteSize(raw string) (uint32, error) {
	if raw == "" {
		return 0, fmt.Errorf("size is empty")
	}
	var re = regexp.MustCompile(`^(?P<size>[0-9]+)\s*(?i)(?P<unit>(k|m|g))b?$`)
	if re.MatchString(raw) {
		size, err := strconv.ParseUint(re.ReplaceAllString(raw, "${size}"), 0, 64)
		if err != nil {
			return 0, fmt.Errorf("value '%s' cannot be parsed with uint32", raw)
		}
		unit := re.ReplaceAllString(raw, "${unit}")
		switch strings.ToLower(unit) {
		case "g":
			size = size << 10
			fallthrough
		case "m":
			size = size << 10
			fallthrough
		case "k":
			size = size << 10
		}
		if size > math.MaxUint32 {
			return 0, fmt.Errorf("value '%s' overflows uint32", raw)
		}
		return uint32(size), nil
	}
	return 0, fmt.Errorf("value '%s' cannot be compiled", raw)
}
