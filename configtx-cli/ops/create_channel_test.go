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

func TestOutputConfigTXToCreateChannel(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "configtx-create-channel-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseOutputFileName := "output.pb"
	baseOutputFormat := "delta"
	baseChannelID := "mychannel"
	baseProfilePath := "testdata/create_mychannel2_profile.yaml"

	// Case: Positive test cases
	cases := []struct {
		channelID string
	}{
		{channelID: "mychannel2"},
	}

	for _, c := range cases {
		outputFileName := fmt.Sprintf("output_%s.pb", c.channelID)
		err = OutputConfigTXToCreateChannel(tempDir, outputFileName, baseOutputFormat, c.channelID, baseProfilePath)
		require.NoError(t, err)

		expected, err := getConfigUpdateFromFile(fmt.Sprintf("testdata/create_%s_expected.pb", c.channelID))
		require.NoError(t, err)
		var expectedBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&expectedBuffer, expected)
		require.NoError(t, err)

		actual, err := getConfigUpdateFromFile(filepath.Join(tempDir, outputFileName))
		require.NoError(t, err)
		var actualBuffer bytes.Buffer
		err = protolator.DeepMarshalJSON(&actualBuffer, actual)
		require.NoError(t, err)

		require.JSONEq(t, string(expectedBuffer.String()), string(actualBuffer.String()), "Output configtx to set %s", c.channelID)
	}

	// Case: Using wrong profile path
	err = OutputConfigTXToCreateChannel(tempDir, baseOutputFileName, baseOutputFormat, baseChannelID, "testdata/wrong-path.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load channel creation profile")

	// Case: Using bad profile
	err = OutputConfigTXToCreateChannel(tempDir, baseOutputFileName, baseOutputFormat, "mychannel", "testdata/org3_profile.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create the channel configuration")

	// Case: Using wrong output format
	err = OutputConfigTXToCreateChannel(tempDir, baseOutputFileName, "wrong format", baseChannelID, baseProfilePath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to output")
}
