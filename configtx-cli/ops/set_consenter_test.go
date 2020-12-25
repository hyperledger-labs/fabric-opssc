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

func TestOutputConfigTXToSetConsenter(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "configtx-set-consenter-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseBlockPath := "testdata/mychannel_config.block"
	baseOutputFileName := "output.pb"
	baseOutputFormat := "delta"
	baseChannelID := "mychannel"
	baseProfilePath := "testdata/org3_consenter_profile.yaml"

	// Case: Positive test cases
	cases := []struct {
		channelID   string
		profilePath string
	}{
		{channelID: "mychannel", profilePath: baseProfilePath},
		{channelID: "mychannel", profilePath: "testdata/org3_consenter_profile_without_reading_certs.yaml"},
		{channelID: "system-channel", profilePath: baseProfilePath},
		{channelID: "system-channel", profilePath: "testdata/org3_consenter_profile_without_reading_certs.yaml"},
	}

	for _, c := range cases {
		outputFileName := fmt.Sprintf("output_%s.pb", c.channelID)
		err = OutputConfigTXToSetConsenter(fmt.Sprintf("testdata/%s_config.block", c.channelID), tempDir, outputFileName, baseOutputFormat, c.channelID, c.profilePath)
		require.NoError(t, err)

		expected, err := getConfigUpdateFromFile(fmt.Sprintf("testdata/set_org3_consenter_to_%s_expected.pb", c.channelID))
		require.NoError(t, err)
		var expectedBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&expectedBuffer, expected)
		require.NoError(t, err)

		actual, err := getConfigUpdateFromFile(filepath.Join(tempDir, outputFileName))
		require.NoError(t, err)
		var actualBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&actualBuffer, actual)
		require.NoError(t, err)

		require.JSONEq(t, string(expectedBuffer.String()), string(actualBuffer.String()), "Output configtx to org3's consenter to %s", c.channelID)
	}

	// Case: Using wrong block path
	err = OutputConfigTXToSetConsenter("testdata/wrong-path.block", tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, baseProfilePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get configtx from block")

	// Case: Using wrong profile path
	err = OutputConfigTXToSetConsenter(baseBlockPath, tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, "testdata/wrong-path.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load consenter profile")

	// Case: Using profile with wrong format
	err = OutputConfigTXToSetConsenter(baseBlockPath, tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, "testdata/org3_profile.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to set the consenter to the current config")

	// Case: Using wrong output format
	err = OutputConfigTXToSetConsenter(baseBlockPath, tempDir, baseOutputFileName, "wrong format", baseChannelID, baseProfilePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to output")
}
