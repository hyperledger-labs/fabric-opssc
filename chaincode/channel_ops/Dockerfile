# SPDX-License-Identifier: Apache-2.0

ARG GO_VER=1.17.5
ARG ALPINE_VER=3.14

FROM golang:${GO_VER}-alpine${ALPINE_VER} AS build

RUN apk add --no-cache \
	bash \
	binutils-gold \
  dumb-init \
	gcc \
	git \
	make \
	musl-dev

ADD ./ /go/src/github.com/chaincode
WORKDIR /go/src/github.com/chaincode

RUN go build -o chaincode-bin -v .

FROM golang:${GO_VER}-alpine${ALPINE_VER}

LABEL org.opencontainers.image.title "Channel Ops Chaincode"
LABEL org.opencontainers.image.description "Channel Ops Chaincode for OpsSC"
LABEL org.opencontainers.image.source "https://github.com/hyperledger-labs/fabric-opssc/tree/main/chaincode/channel_ops"

COPY --from=build /usr/bin/dumb-init /usr/bin/dumb-init
COPY --from=build /go/src/github.com/chaincode/chaincode-bin /usr/bin/chaincode-bin

WORKDIR /var/hyperledger/chaincode-bin
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["sh", "-c", "exec /usr/bin/chaincode-bin -peer.address=$CORE_PEER_ADDRESS"]