/*
Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/stretchr/testify/require"
)

func TestOutputEnvelopedConfigTx(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "configtx-create-envelope-test")
	require.NoErrorf(t, err, "Error creating temp dir")
	defer os.RemoveAll(tempDir)

	baseProfilePath := "testdata/configtx_profile.yaml"
	baseOutputFileName := "output.pb"

	err = OutputEnvelopedConfigTx(baseProfilePath, tempDir, baseOutputFileName)
	require.NoError(t, err)

	b, err := ioutil.ReadFile(filepath.Join(tempDir, baseOutputFileName))
	require.NoError(t, err)
	var actual cb.Envelope
	err = proto.Unmarshal(b, &actual)
	require.NoError(t, err)

	// Case: Using wrong configtx path
	err = OutputEnvelopedConfigTx("testdata/wrong-path.yaml", tempDir, baseOutputFileName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get the ConfigTxProfile")

	// Case: Using the profile which has not contains the required parameters
	err = OutputEnvelopedConfigTx("testdata/org3_profile.yaml", tempDir, baseOutputFileName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "the required parameter 'ConfigUpdate' is empty")

	// Case: Using the profile that the config is not base64
	err = OutputEnvelopedConfigTx("testdata/configtx_profile_that_config_is_not_base64.yaml", tempDir, baseOutputFileName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode the ConfigUpdate to base64")

	// Case: Using the profile that the config has wrong format
	err = OutputEnvelopedConfigTx("testdata/configtx_profile_that_config_has_wrong_format.yaml", tempDir, baseOutputFileName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal the ConfigUpdate")

	// Case: Using the profile that the sign is not base64
	err = OutputEnvelopedConfigTx("testdata/configtx_profile_that_sign_is_not_base64.yaml", tempDir, baseOutputFileName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode the ConfigSignature to base64")

	// Case: Using the profile that the sign has wrong format
	err = OutputEnvelopedConfigTx("testdata/configtx_profile_that_sign_has_wrong_format.yaml", tempDir, baseOutputFileName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal the ConfigSignature")

	// Case: Using no output file name
	err = OutputEnvelopedConfigTx(baseProfilePath, tempDir, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to output")
}
