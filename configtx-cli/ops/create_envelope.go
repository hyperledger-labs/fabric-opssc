/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/configtx"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// ConfigTxProfile is a profile to output config envelope.
type ConfigTxProfile struct {
	ConfigUpdate string            `yaml:"configUpdate"`
	Signatures   map[string]string `yaml:"signatures"`
}

// OutputEnvelopedConfigTx outputs the config envelope by using the specified profile to the specified output path.
func OutputEnvelopedConfigTx(profile string, outputDirectory string, outputFile string) error {

	logger.Info("create an enveloped configtx")
	configTxProfile := ConfigTxProfile{}
	data, err := ioutil.ReadFile(profile)
	if err != nil {
		return errors.Wrap(err, "failed to get the ConfigTxProfile")
	}

	err = yaml.Unmarshal([]byte(data), &configTxProfile)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal the ConfigTxProfile")
	}

	if configTxProfile.ConfigUpdate == "" {
		return errors.New("the required parameter 'ConfigUpdate' is empty")
	}

	updateBin, err := base64.StdEncoding.DecodeString(configTxProfile.ConfigUpdate)
	if err != nil {
		return errors.Wrap(err, "failed to decode the ConfigUpdate to base64")
	}
	var update cb.ConfigUpdate
	if err := proto.Unmarshal(updateBin, &update); err != nil {
		return errors.Wrap(err, "failed to unmarshal the ConfigUpdate")
	}

	var signatures []*cb.ConfigSignature
	for _, s := range configTxProfile.Signatures {
		sigBin, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return errors.Wrap(err, "failed to decode the ConfigSignature to base64")
		}
		var sigPb cb.ConfigSignature
		if err := proto.Unmarshal(sigBin, &sigPb); err != nil {
			return errors.Wrap(err, "failed to unmarshal the ConfigSignature")
		}
		signatures = append(signatures, &sigPb)
	}

	env, err := configtx.NewEnvelope(updateBin, signatures...)
	if err != nil {
		return errors.Wrap(err, "failed to create the envelope using the ConfigUpdate and ConfigSignatures")
	}

	err = writeFile(filepath.Join(outputDirectory, outputFile), protoutil.MarshalOrPanic(env), 0640)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}
