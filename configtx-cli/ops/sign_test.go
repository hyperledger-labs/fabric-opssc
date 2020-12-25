/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/require"
)

func TestOutputSign(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "configtx-sign-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseConfigTxPath := "testdata/set_mychannel_expected.pb"
	baseOutputFileName := "output.pb"
	baseMSPID := "Org1MSP"
	baseOrgLabel := "org1"
	baseSignKeyPath := fmt.Sprintf("testdata/peerOrganizations/%s.example.com/users/Admin@%s.example.com/msp/keystore/priv_sk", baseOrgLabel, baseOrgLabel)
	baseSignCertPath := fmt.Sprintf("testdata/peerOrganizations/%s.example.com/users/Admin@%s.example.com/msp/signcerts/cert.pem", baseOrgLabel, baseOrgLabel)

	// Case: Positive test cases
	cases := []struct {
		channelID string
	}{
		{channelID: "mychannel"},
	}

	for _, c := range cases {
		outputFileName := fmt.Sprintf("output_%s.pb", c.channelID)
		err = OutputSign(fmt.Sprintf("testdata/set_%s_expected.pb", c.channelID), tempDir, outputFileName, baseMSPID, baseSignKeyPath, baseSignCertPath)
		require.NoError(t, err)

		b, err := ioutil.ReadFile(filepath.Join(tempDir, outputFileName))
		require.NoError(t, err)
		var actual cb.ConfigSignature
		err = proto.Unmarshal(b, &actual)
		require.NoError(t, err)
	}

	// Case: Using wrong configtx path
	err = OutputSign("testdata/wrong-path.pb", tempDir, baseOutputFileName, baseMSPID, baseSignKeyPath, baseSignCertPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get the ConfigUpdate from the specified file")

	// Case: Using no output file name
	err = OutputSign(baseConfigTxPath, tempDir, "", baseMSPID, baseSignKeyPath, baseSignCertPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to output")

	// Case: Using wrong key path
	err = OutputSign(baseConfigTxPath, tempDir, baseOutputFileName, baseMSPID, "wrong-key-path", baseSignCertPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get the sign key from the specified file")

	// Case: Using wrong cert path
	err = OutputSign(baseConfigTxPath, tempDir, baseOutputFileName, baseMSPID, baseSignKeyPath, "wrong-cert-path")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get the sign certificate from the specified file")
}
