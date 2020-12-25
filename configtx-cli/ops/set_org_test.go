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

func TestOutputConfigTXToSetOrg(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "configtx-set-org-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseBlockPath := "testdata/mychannel_config.block"
	baseOutputFileName := "output.pb"
	baseOutputFormat := "delta"
	baseChannelID := "mychannel"
	baseProfilePath := "testdata/org3_profile.yaml"
	baseOrgType := "Application|Orderer"

	// Case: Positive test cases
	cases := []struct {
		channelID   string
		orgType     string
		profilePath string
	}{
		{channelID: "mychannel", orgType: "Application|Orderer", profilePath: baseProfilePath},
		{channelID: "mychannel", orgType: "Application|Orderer", profilePath: "testdata/org3_profile_without_reading_mspdir.yaml"},
		{channelID: "system-channel", orgType: "Consortiums|Orderer", profilePath: baseProfilePath},
		{channelID: "system-channel", orgType: "Consortiums|Orderer", profilePath: "testdata/org3_profile_without_reading_mspdir.yaml"},
	}

	for _, c := range cases {
		outputFileName := fmt.Sprintf("output_%s.pb", c.channelID)
		err = OutputConfigTXToSetOrg(fmt.Sprintf("testdata/%s_config.block", c.channelID), tempDir, outputFileName, baseOutputFormat, c.channelID, c.profilePath, c.orgType)
		require.NoError(t, err)

		expected, err := getConfigUpdateFromFile(fmt.Sprintf("testdata/set_org3_to_%s_expected.pb", c.channelID))
		require.NoError(t, err)
		var expectedBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&expectedBuffer, expected)
		require.NoError(t, err)

		actual, err := getConfigUpdateFromFile(filepath.Join(tempDir, outputFileName))
		require.NoError(t, err)
		var actualBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&actualBuffer, actual)
		require.NoError(t, err)

		require.JSONEq(t, string(expectedBuffer.String()), string(actualBuffer.String()), "Output configtx to set org3 to %s", c.channelID)
	}

	// Case: Using wrong block path
	err = OutputConfigTXToSetOrg("testdata/wrong-path.block", tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, baseProfilePath, baseOrgType)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get configtx from block")

	// Case: Using wrong profile path
	err = OutputConfigTXToSetOrg(baseBlockPath, tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, "testdata/wrong-path.yaml", baseOrgType)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load organization profile")

	// Case: Using wrong OrgType
	err = OutputConfigTXToSetOrg(baseBlockPath, tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, baseProfilePath, "WrongOrgType")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to set the organization to the current config")

	// Case: Using wrong output format
	err = OutputConfigTXToSetOrg(baseBlockPath, tempDir, baseOutputFileName, "wrong format", baseChannelID, baseProfilePath, baseOrgType)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to output")
}
