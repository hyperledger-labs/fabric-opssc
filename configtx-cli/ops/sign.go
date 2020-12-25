/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"io/ioutil"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/configtx"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
)

// OutputSign outputs the ConfigSignature signed by specified MSP to the specified ConfigUpdate using with the specified sign key and cert.
func OutputSign(configTxPath string, outputDirectory string, outputFile string, mspID string, keyPath string, certPath string) error {
	logger.Infof("create an ConfigSignature with %s", mspID)

	delta, err := getConfigUpdateFromConfigTxFile(configTxPath)
	if err != nil {
		return errors.Wrap(err, "failed to get the ConfigUpdate from the specified file")
	}

	cert, err := getCertificateFromFile(certPath)
	if err != nil {
		return errors.Wrap(err, "failed to get the sign certificate from the specified file")
	}

	key, err := getPrivateKeyFromFile(keyPath)
	if err != nil {
		return errors.Wrap(err, "failed to get the sign key from the specified file")
	}

	signingIdentity := configtx.SigningIdentity{
		MSPID:       mspID,
		Certificate: cert,
		PrivateKey:  key,
	}

	configSignature, err := signingIdentity.CreateConfigSignature(delta)
	if err != nil {
		return errors.Wrap(err, "failed to create ConfigSignature using the specified ConfigUpdate")
	}

	printDebug("Output ConfigSignature", configSignature)

	marshaledSignature, err := proto.Marshal(configSignature)
	if err != nil {
		return errors.Wrap(err, "failed marshaling ConfigSignature")
	}

	err = writeFile(filepath.Join(outputDirectory, outputFile), marshaledSignature, 0640)
	if err != nil {
		return errors.Wrap(err, "failed to output")
	}

	return nil
}

func getConfigUpdateFromConfigTxFile(configTxPath string) ([]byte, error) {
	logger.Info("getting ConfigUpdate from configtx file")
	configTxBin, err := ioutil.ReadFile(configTxPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the configtx file")
	}

	// To check that the config can be unmarshalled
	config := &cb.ConfigUpdate{}
	err = proto.Unmarshal(configTxBin, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed unmarshalling the ConfigUpdate")
	}

	return configTxBin, nil
}
