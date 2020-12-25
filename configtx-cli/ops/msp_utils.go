/*
Copyright IBM Corp. All Rights Reserved.

Copyright 2020 Hitachi America, Ltd. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ops

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/configtx"
	"github.com/hyperledger/fabric-config/configtx/membership"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/bccsp"
	mspConfigBuilder "github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// This code is based on msp/configbuilder in hyperledger/fabric
// Ref: https://github.com/hyperledger/fabric/blob/master/msp/configbuilder.go

// ProviderType indicates the type of an identity provider
type ProviderType int

// The ProviderType of a member relative to the member API
const (
	FABRIC ProviderType = iota // MSP is of FABRIC type
	IDEMIX                     // MSP is of IDEMIX type
	OTHER                      // MSP is of OTHER TYPE

	// NOTE: as new types are added to this set,
	// the mspTypes map below must be extended

	cacerts              = "cacerts"
	admincerts           = "admincerts"
	signcerts            = "signcerts"
	keystore             = "keystore"
	intermediatecerts    = "intermediatecerts"
	crlsfolder           = "crls"
	configfilename       = "config.yaml"
	tlscacerts           = "tlscacerts"
	tlsintermediatecerts = "tlsintermediatecerts"
)

// getMSPConfigFromDir returns an MSP config given directory and ID.
// Currently, this only supports FABRIC msp and does not support IDEMIX.
func getMSPConfigFromDir(dir string, ID string) (configtx.MSP, error) {
	config, err := getMspConfig(dir, ID)
	if err != nil {
		return configtx.MSP{}, err
	}

	return getMSPConfig(config)
}

func getMSPConfig(config *msp.MSPConfig) (configtx.MSP, error) {
	fabricMSPConfig := &msp.FabricMSPConfig{}

	err := proto.Unmarshal(config.Config, fabricMSPConfig)
	if err != nil {
		return configtx.MSP{}, fmt.Errorf("unmarshaling fabric msp config: %v", err)
	}

	// ROOT CERTS
	rootCerts, err := parseCertificateListFromBytes(fabricMSPConfig.RootCerts)
	if err != nil {
		return configtx.MSP{}, fmt.Errorf("parsing root certs: %v", err)
	}

	// INTERMEDIATE CERTS
	intermediateCerts, err := parseCertificateListFromBytes(fabricMSPConfig.IntermediateCerts)
	if err != nil {
		return configtx.MSP{}, fmt.Errorf("parsing intermediate certs: %v", err)
	}

	// ADMIN CERTS
	adminCerts, err := parseCertificateListFromBytes(fabricMSPConfig.Admins)
	if err != nil {
		return configtx.MSP{}, fmt.Errorf("parsing admin certs: %v", err)
	}

	// REVOCATION LIST
	revocationList, err := parseCRL(fabricMSPConfig.RevocationList)
	if err != nil {
		return configtx.MSP{}, err
	}

	// OU IDENTIFIERS
	ouIdentifiers, err := parseOUIdentifiers(fabricMSPConfig.OrganizationalUnitIdentifiers)
	if err != nil {
		return configtx.MSP{}, fmt.Errorf("parsing ou identifiers: %v", err)
	}

	// TLS ROOT CERTS
	tlsRootCerts, err := parseCertificateListFromBytes(fabricMSPConfig.TlsRootCerts)
	if err != nil {
		return configtx.MSP{}, fmt.Errorf("parsing tls root certs: %v", err)
	}

	// TLS INTERMEDIATE CERTS
	tlsIntermediateCerts, err := parseCertificateListFromBytes(fabricMSPConfig.TlsIntermediateCerts)
	if err != nil {
		return configtx.MSP{}, fmt.Errorf("parsing tls intermediate certs: %v", err)
	}

	// NODE OUS
	var (
		clientOUIdentifierCert  *x509.Certificate
		peerOUIdentifierCert    *x509.Certificate
		adminOUIdentifierCert   *x509.Certificate
		ordererOUIdentifierCert *x509.Certificate
		nodeOUs                 membership.NodeOUs
	)
	if fabricMSPConfig.FabricNodeOus != nil {
		clientOUIdentifierCert, err = parseCertificateFromBytes(fabricMSPConfig.FabricNodeOus.ClientOuIdentifier.Certificate)
		if err != nil {
			return configtx.MSP{}, fmt.Errorf("parsing client ou identifier cert: %v", err)
		}

		peerOUIdentifierCert, err = parseCertificateFromBytes(fabricMSPConfig.FabricNodeOus.PeerOuIdentifier.Certificate)
		if err != nil {
			return configtx.MSP{}, fmt.Errorf("parsing peer ou identifier cert: %v", err)
		}

		adminOUIdentifierCert, err = parseCertificateFromBytes(fabricMSPConfig.FabricNodeOus.AdminOuIdentifier.Certificate)
		if err != nil {
			return configtx.MSP{}, fmt.Errorf("parsing admin ou identifier cert: %v", err)
		}

		ordererOUIdentifierCert, err = parseCertificateFromBytes(fabricMSPConfig.FabricNodeOus.OrdererOuIdentifier.Certificate)
		if err != nil {
			return configtx.MSP{}, fmt.Errorf("parsing orderer ou identifier cert: %v", err)
		}

		nodeOUs = membership.NodeOUs{
			Enable: fabricMSPConfig.FabricNodeOus.Enable,
			ClientOUIdentifier: membership.OUIdentifier{
				Certificate:                  clientOUIdentifierCert,
				OrganizationalUnitIdentifier: fabricMSPConfig.FabricNodeOus.ClientOuIdentifier.OrganizationalUnitIdentifier,
			},
			PeerOUIdentifier: membership.OUIdentifier{
				Certificate:                  peerOUIdentifierCert,
				OrganizationalUnitIdentifier: fabricMSPConfig.FabricNodeOus.PeerOuIdentifier.OrganizationalUnitIdentifier,
			},
			AdminOUIdentifier: membership.OUIdentifier{
				Certificate:                  adminOUIdentifierCert,
				OrganizationalUnitIdentifier: fabricMSPConfig.FabricNodeOus.AdminOuIdentifier.OrganizationalUnitIdentifier,
			},
			OrdererOUIdentifier: membership.OUIdentifier{
				Certificate:                  ordererOUIdentifierCert,
				OrganizationalUnitIdentifier: fabricMSPConfig.FabricNodeOus.OrdererOuIdentifier.OrganizationalUnitIdentifier,
			},
		}
	}

	return configtx.MSP{
		Name:                          fabricMSPConfig.Name,
		RootCerts:                     rootCerts,
		IntermediateCerts:             intermediateCerts,
		Admins:                        adminCerts,
		RevocationList:                revocationList,
		OrganizationalUnitIdentifiers: ouIdentifiers,
		CryptoConfig: membership.CryptoConfig{
			SignatureHashFamily:            fabricMSPConfig.CryptoConfig.SignatureHashFamily,
			IdentityIdentifierHashFunction: fabricMSPConfig.CryptoConfig.IdentityIdentifierHashFunction,
		},
		TLSRootCerts:         tlsRootCerts,
		TLSIntermediateCerts: tlsIntermediateCerts,
		NodeOUs:              nodeOUs,
	}, nil
}

func getMspConfig(dir string, ID string) (*msp.MSPConfig, error) {
	cacertDir := filepath.Join(dir, cacerts)
	admincertDir := filepath.Join(dir, admincerts)
	intermediatecertsDir := filepath.Join(dir, intermediatecerts)
	crlsDir := filepath.Join(dir, crlsfolder)
	configFile := filepath.Join(dir, configfilename)
	tlscacertDir := filepath.Join(dir, tlscacerts)
	tlsintermediatecertsDir := filepath.Join(dir, tlsintermediatecerts)

	cacerts, err := getPemMaterialFromDir(cacertDir)
	if err != nil || len(cacerts) == 0 {
		return nil, errors.WithMessagef(err, "could not load a valid ca certificate from directory %s", cacertDir)
	}

	admincert, err := getPemMaterialFromDir(admincertDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.WithMessagef(err, "could not load a valid admin certificate from directory %s", admincertDir)
	}

	intermediatecerts, err := getPemMaterialFromDir(intermediatecertsDir)
	if !os.IsNotExist(err) {
		return nil, errors.WithMessagef(err, "failed loading intermediate ca certs at [%s]", intermediatecertsDir)
	}

	tlsCACerts, err := getPemMaterialFromDir(tlscacertDir)
	tlsIntermediateCerts := [][]byte{}
	if os.IsNotExist(err) {
	} else if err != nil {
		return nil, errors.WithMessagef(err, "failed loading TLS ca certs at [%s]", tlsintermediatecertsDir)
	} else if len(tlsCACerts) != 0 {
		tlsIntermediateCerts, err = getPemMaterialFromDir(tlsintermediatecertsDir)
		if os.IsNotExist(err) {
		} else if err != nil {
			return nil, errors.WithMessagef(err, "failed loading TLS intermediate ca certs at [%s]", tlsintermediatecertsDir)
		}
	}

	crls, err := getPemMaterialFromDir(crlsDir)
	if os.IsNotExist(err) {
	} else if err != nil {
		return nil, errors.WithMessagef(err, "failed loading crls at [%s]", crlsDir)
	}

	// Load configuration file
	// if the configuration file is there then load it
	// otherwise skip it
	var ouis []*msp.FabricOUIdentifier
	var nodeOUs *msp.FabricNodeOUs
	_, err = os.Stat(configFile)
	if err == nil {
		// load the file, if there is a failure in loading it then
		// return an error
		raw, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed loading configuration file at [%s]", configFile)
		}

		configuration := mspConfigBuilder.Configuration{}
		err = yaml.Unmarshal(raw, &configuration)
		if err != nil {
			return nil, errors.Wrapf(err, "failed unmarshalling configuration file at [%s]", configFile)
		}

		// Prepare OrganizationalUnitIdentifiers
		if len(configuration.OrganizationalUnitIdentifiers) > 0 {
			for _, ouID := range configuration.OrganizationalUnitIdentifiers {
				f := filepath.Join(dir, ouID.Certificate)
				raw, err = readFile(f)
				if err != nil {
					return nil, errors.Wrapf(err, "failed loading OrganizationalUnit certificate at [%s]", f)
				}

				oui := &msp.FabricOUIdentifier{
					Certificate:                  raw,
					OrganizationalUnitIdentifier: ouID.OrganizationalUnitIdentifier,
				}
				ouis = append(ouis, oui)
			}
		}

		// Prepare NodeOUs
		if configuration.NodeOUs != nil && configuration.NodeOUs.Enable {
			nodeOUs = &msp.FabricNodeOUs{
				Enable: true,
			}
			if configuration.NodeOUs.ClientOUIdentifier != nil && len(configuration.NodeOUs.ClientOUIdentifier.OrganizationalUnitIdentifier) != 0 {
				nodeOUs.ClientOuIdentifier = &msp.FabricOUIdentifier{OrganizationalUnitIdentifier: configuration.NodeOUs.ClientOUIdentifier.OrganizationalUnitIdentifier}
			}
			if configuration.NodeOUs.PeerOUIdentifier != nil && len(configuration.NodeOUs.PeerOUIdentifier.OrganizationalUnitIdentifier) != 0 {
				nodeOUs.PeerOuIdentifier = &msp.FabricOUIdentifier{OrganizationalUnitIdentifier: configuration.NodeOUs.PeerOUIdentifier.OrganizationalUnitIdentifier}
			}
			if configuration.NodeOUs.AdminOUIdentifier != nil && len(configuration.NodeOUs.AdminOUIdentifier.OrganizationalUnitIdentifier) != 0 {
				nodeOUs.AdminOuIdentifier = &msp.FabricOUIdentifier{OrganizationalUnitIdentifier: configuration.NodeOUs.AdminOUIdentifier.OrganizationalUnitIdentifier}
			}
			if configuration.NodeOUs.OrdererOUIdentifier != nil && len(configuration.NodeOUs.OrdererOUIdentifier.OrganizationalUnitIdentifier) != 0 {
				nodeOUs.OrdererOuIdentifier = &msp.FabricOUIdentifier{OrganizationalUnitIdentifier: configuration.NodeOUs.OrdererOUIdentifier.OrganizationalUnitIdentifier}
			}

			// Read certificates, if defined

			// ClientOU
			if nodeOUs.ClientOuIdentifier != nil {
				nodeOUs.ClientOuIdentifier.Certificate = loadCertificateAt(dir, configuration.NodeOUs.ClientOUIdentifier.Certificate, "ClientOU")
			}
			// PeerOU
			if nodeOUs.PeerOuIdentifier != nil {
				nodeOUs.PeerOuIdentifier.Certificate = loadCertificateAt(dir, configuration.NodeOUs.PeerOUIdentifier.Certificate, "PeerOU")
			}
			// AdminOU
			if nodeOUs.AdminOuIdentifier != nil {
				nodeOUs.AdminOuIdentifier.Certificate = loadCertificateAt(dir, configuration.NodeOUs.AdminOUIdentifier.Certificate, "AdminOU")
			}
			// OrdererOU
			if nodeOUs.OrdererOuIdentifier != nil {
				nodeOUs.OrdererOuIdentifier.Certificate = loadCertificateAt(dir, configuration.NodeOUs.OrdererOUIdentifier.Certificate, "OrdererOU")
			}
		}
	}

	// Set FabricCryptoConfig
	cryptoConfig := &msp.FabricCryptoConfig{
		SignatureHashFamily:            bccsp.SHA2,
		IdentityIdentifierHashFunction: bccsp.SHA256,
	}

	// Compose FabricMSPConfig
	fmspconf := &msp.FabricMSPConfig{
		Admins:                        admincert,
		RootCerts:                     cacerts,
		IntermediateCerts:             intermediatecerts,
		Name:                          ID,
		OrganizationalUnitIdentifiers: ouis,
		RevocationList:                crls,
		CryptoConfig:                  cryptoConfig,
		TlsRootCerts:                  tlsCACerts,
		TlsIntermediateCerts:          tlsIntermediateCerts,
		FabricNodeOus:                 nodeOUs,
	}

	fmpsjs, _ := proto.Marshal(fmspconf)

	mspconf := &msp.MSPConfig{Config: fmpsjs, Type: int32(FABRIC)}

	return mspconf, nil
}

// getCertificateFromFile returns the x509 certificate from the specified file.
func getCertificateFromFile(certificatePath string) (*x509.Certificate, error) {
	bytes, err := readPemFile(certificatePath)
	if err != nil {
		return nil, err
	}
	return parseCertificateFromBytes(bytes)
}

func loadCertificateAt(dir, certificatePath string, ouType string) []byte {

	if certificatePath == "" {
		logger.Debugf("Specific certificate for %s is not configured", ouType)
		return nil
	}

	f := filepath.Join(dir, certificatePath)
	raw, err := readFile(f)
	if err != nil {
	} else {
		return raw
	}

	return nil
}

func readFile(file string) ([]byte, error) {
	fileCont, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read file %s", file)
	}

	return fileCont, nil
}

func readPemFile(file string) ([]byte, error) {
	bytes, err := readFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "reading from file %s failed", file)
	}

	b, _ := pem.Decode(bytes)
	if b == nil { // TODO: also check that the type is what we expect (cert vs key..)
		return nil, errors.Errorf("no pem content for file %s", file)
	}

	return bytes, nil
}

func getPemMaterialFromDir(dir string) ([][]byte, error) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, err
	}

	content := make([][]byte, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read directory %s", dir)
	}

	for _, f := range files {
		fullName := filepath.Join(dir, f.Name())

		f, err := os.Stat(fullName)
		if err != nil {
			continue
		}
		if f.IsDir() {
			continue
		}

		item, err := readPemFile(fullName)
		if err != nil {
			continue
		}

		content = append(content, item)
	}

	return content, nil
}

func parseCertificateFromBytes(cert []byte) (*x509.Certificate, error) {
	pemBlock, _ := pem.Decode(cert)
	if pemBlock == nil {
		return &x509.Certificate{}, errors.New("failed to decode PEM block")
	}

	certificate, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return &x509.Certificate{}, err
	}

	return certificate, nil
}

func parseCertificateListFromBytes(certs [][]byte) ([]*x509.Certificate, error) {
	certificateList := []*x509.Certificate{}

	for _, cert := range certs {
		certificate, err := parseCertificateFromBytes(cert)
		if err != nil {
			return certificateList, err
		}

		certificateList = append(certificateList, certificate)
	}

	return certificateList, nil
}

func parseCRL(crls [][]byte) ([]*pkix.CertificateList, error) {
	certificateLists := []*pkix.CertificateList{}

	for _, crl := range crls {
		pemBlock, _ := pem.Decode(crl)

		certificateList, err := x509.ParseCRL(pemBlock.Bytes)
		if err != nil {
			return certificateLists, fmt.Errorf("parsing crl: %v", err)
		}

		certificateLists = append(certificateLists, certificateList)
	}

	return certificateLists, nil
}

func parseOUIdentifiers(identifiers []*msp.FabricOUIdentifier) ([]membership.OUIdentifier, error) {
	fabricIdentifiers := []membership.OUIdentifier{}

	for _, identifier := range identifiers {
		cert, err := parseCertificateFromBytes(identifier.Certificate)
		if err != nil {
			return fabricIdentifiers, err
		}

		fabricOUIdentifier := membership.OUIdentifier{
			Certificate:                  cert,
			OrganizationalUnitIdentifier: identifier.OrganizationalUnitIdentifier,
		}

		fabricIdentifiers = append(fabricIdentifiers, fabricOUIdentifier)
	}

	return fabricIdentifiers, nil
}

func toBytes(source []string) [][]byte {
	list := [][]byte{}
	for _, e := range source {
		list = append(list, []byte(e))
	}
	return list
}

// createMSPConfig creates MSP for fabric-config from the specified MSP profile.
func createMSPConfig(mspID string, msp MSP) (configtx.MSP, error) {

	rCerts, err := parseCertificateListFromBytes(toBytes(msp.RootCerts))
	if err != nil {
		return configtx.MSP{}, err
	}
	iCerts, err := parseCertificateListFromBytes(toBytes(msp.IntermediateCerts))
	if err != nil {
		return configtx.MSP{}, err
	}
	admins, err := parseCertificateListFromBytes(toBytes(msp.Admins))
	if err != nil {
		return configtx.MSP{}, err
	}
	crls, err := parseCRL(toBytes(msp.RevocationList))
	if err != nil {
		return configtx.MSP{}, err
	}
	tlsRCerts, err := parseCertificateListFromBytes(toBytes(msp.TLSRootCerts))
	if err != nil {
		return configtx.MSP{}, err
	}
	tlsICerts, err := parseCertificateListFromBytes(toBytes(msp.TLSIntermediateCerts))
	if err != nil {
		return configtx.MSP{}, err
	}

	clientOUID, err := newOUIdentifier(msp.NodeOUs.ClientOUIdentifier)
	if err != nil {
		return configtx.MSP{}, err
	}
	peerOUID, err := newOUIdentifier(msp.NodeOUs.PeerOUIdentifier)
	if err != nil {
		return configtx.MSP{}, err
	}
	adminOUID, err := newOUIdentifier(msp.NodeOUs.AdminOUIdentifier)
	if err != nil {
		return configtx.MSP{}, err
	}
	ordererOUID, err := newOUIdentifier(msp.NodeOUs.OrdererOUIdentifier)
	if err != nil {
		return configtx.MSP{}, err
	}
	OUIDs, err := newOUIdentifiers(msp.OrganizationalUnitIdentifiers)
	if err != nil {
		return configtx.MSP{}, err
	}

	return configtx.MSP{
		Name:                          mspID,
		RootCerts:                     rCerts,
		IntermediateCerts:             iCerts,
		Admins:                        admins,
		RevocationList:                crls,
		OrganizationalUnitIdentifiers: OUIDs,
		CryptoConfig: membership.CryptoConfig{
			SignatureHashFamily:            "SHA2",
			IdentityIdentifierHashFunction: "SHA256",
		},
		TLSRootCerts:         tlsRCerts,
		TLSIntermediateCerts: tlsICerts,
		NodeOUs: membership.NodeOUs{
			Enable:              msp.NodeOUs.Enable,
			ClientOUIdentifier:  clientOUID,
			PeerOUIdentifier:    peerOUID,
			AdminOUIdentifier:   adminOUID,
			OrdererOUIdentifier: ordererOUID,
		},
	}, nil
}

func newOUIdentifiers(identifiers []OUIdentifier) ([]membership.OUIdentifier, error) {
	list := []membership.OUIdentifier{}
	for _, id := range identifiers {
		new, err := newOUIdentifier(&id)
		if err != nil {
			return list, err
		}
		list = append(list, new)
	}
	return list, nil
}

func newOUIdentifier(identifier *OUIdentifier) (membership.OUIdentifier, error) {
	if identifier == nil {
		return membership.OUIdentifier{}, nil
	}
	var cert *x509.Certificate
	var err error
	if identifier.Certificate != "" {
		cert, err = parseCertificateFromBytes([]byte(identifier.Certificate))
		if err != nil {
			return membership.OUIdentifier{}, err
		}
	}
	return membership.OUIdentifier{
		Certificate:                  cert,
		OrganizationalUnitIdentifier: identifier.OrganizationalUnitIdentifier,
	}, nil
}

// getPrivateKeyFromFile returns the private key from the specified file.
func getPrivateKeyFromFile(path string) (crypto.PrivateKey, error) {
	binaries, err := readPemFile(path)
	if err != nil {
		return nil, err
	}
	return parsePrivateKey(binaries)
}

// Based on fabric/cmd/common/signer/signer.go:
func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {

	pemBlock, _ := pem.Decode(der)

	// OpenSSL 1.0.0 generates PKCS#8 keys.
	if key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes); err == nil {
		switch key := key.(type) {
		// Fabric only supports ECDSA at the moment.
		case *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, errors.Errorf("found unknown private key type (%T) in PKCS#8 wrapping", key)
		}
	}

	// OpenSSL ecparam generates SEC1 EC private keys for ECDSA.
	key, err := x509.ParseECPrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, errors.Errorf("failed to parse private key: %v", err)
	}

	return key, nil
}
