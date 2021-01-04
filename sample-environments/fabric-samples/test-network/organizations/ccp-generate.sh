#!/bin/bash

function one_line_pem {
    echo "`awk 'NF {sub(/\\n/, ""); printf "%s\\\\\\\n",$0;}' $1`"
}

function json_ccp {
    local PP=$(one_line_pem $5)
    local OP=$(one_line_pem $6)
    local CP=$(one_line_pem $7)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${P0PORT}/$2/" \
        -e "s/\${O0PORT}/$3/" \
        -e "s/\${CAPORT}/$4/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        -e "s#\${ORDERERPEM}#$OP#" \
        organizations/ccp-template.json
}

function yaml_ccp {
    local PP=$(one_line_pem $5)
    local OP=$(one_line_pem $6)
    local CP=$(one_line_pem $7)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${P0PORT}/$2/" \
        -e "s/\${O0PORT}/$3/" \
        -e "s/\${CAPORT}/$4/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        -e "s#\${ORDERERPEM}#$OP#" \
        organizations/ccp-template.yaml | sed -e $'s/\\\\n/\\\n        /g'
}

ORG=1
P0PORT=7051
O0PORT=7050
CAPORT=7054
PEERPEM=organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
ORDERERPEM=organizations/peerOrganizations/org1.example.com/orderers/orderer0.org1.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
CAPEM=organizations/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

echo "$(json_ccp $ORG $P0PORT $O0PORT $CAPORT $PEERPEM $ORDERERPEM $CAPEM)" > organizations/peerOrganizations/org1.example.com/connection-org1.json
echo "$(yaml_ccp $ORG $P0PORT $O0PORT $CAPORT $PEERPEM $ORDERERPEM $CAPEM)" > organizations/peerOrganizations/org1.example.com/connection-org1.yaml

ORG=2
P0PORT=9051
O0PORT=9050
CAPORT=8054
PEERPEM=organizations/peerOrganizations/org2.example.com/tlsca/tlsca.org2.example.com-cert.pem
ORDERERPEM=organizations/peerOrganizations/org2.example.com/orderers/orderer0.org2.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
CAPEM=organizations/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem

echo "$(json_ccp $ORG $P0PORT $O0PORT $CAPORT $PEERPEM $ORDERERPEM $CAPEM)" > organizations/peerOrganizations/org2.example.com/connection-org2.json
echo "$(yaml_ccp $ORG $P0PORT $O0PORT $CAPORT $PEERPEM $ORDERERPEM $CAPEM)" > organizations/peerOrganizations/org2.example.com/connection-org2.yaml
