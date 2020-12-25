/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-config/protolator"
	"github.com/stretchr/testify/require"
)

func TestOutputConfigTXToRemoveOrg(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "configtx-remove-org-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseChannelID := "mychannel"
	baseBlockPath := "testdata/mychannel_config.block"
	baseOutputFileName := "output.pb"
	baseOutputFormat := "delta"
	baseTargetOrg := "Org2MSP"
	baseOrgType := "Application|Orderer"

	// Case: Positive test cases
	cases := []struct {
		channelID string
		orgType   string
	}{
		{channelID: "mychannel", orgType: "Application|Orderer"},
		{channelID: "system-channel", orgType: "Consortiums|Orderer"},
	}

	for _, c := range cases {
		outputFileName := fmt.Sprintf("output_%s.pb", c.channelID)
		err = OutputConfigTXToRemoveOrg(fmt.Sprintf("testdata/%s_config.block", c.channelID), tempDir, outputFileName, baseOutputFormat, c.channelID, baseTargetOrg, c.orgType)
		require.NoError(t, err)

		expected, err := getConfigUpdateFromFile(fmt.Sprintf("testdata/remove_org2_from_%s_expected.pb", c.channelID))
		require.NoError(t, err)
		var expectedBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&expectedBuffer, expected)
		require.NoError(t, err)
		actual, err := getConfigUpdateFromFile(filepath.Join(tempDir, outputFileName))
		require.NoError(t, err)

		var actualBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&actualBuffer, actual)
		require.NoError(t, err)

		require.JSONEq(t, string(expectedBuffer.String()), string(actualBuffer.String()), "Output configtx to remove org2 to %s", c.channelID)
	}

	// Case: Using wrong block path
	err = OutputConfigTXToRemoveOrg("testdata/wrong-path.block", tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, baseTargetOrg, baseOrgType)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get configtx from block")

	// Case: Using missing Org
	err = OutputConfigTXToRemoveOrg(baseBlockPath, tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, "WrongOrg", baseOrgType)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to remove the organization from the current config")

	// Case: Using wrong output format
	err = OutputConfigTXToRemoveOrg(baseBlockPath, tempDir, baseOutputFileName, "wrong format", baseChannelID, baseTargetOrg, baseOrgType)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to output")
}
