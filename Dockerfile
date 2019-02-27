#------------------------------------------------------------------------------
# Base image and required packages
#------------------------------------------------------------------------------

FROM alpine:3.7
RUN apk add -U --no-cache ca-certificates

#------------------------------------------------------------------------------
# Build and install:
#------------------------------------------------------------------------------

ENV GOPATH="/go"
RUN apk add -U --no-cache -t dev git go musl-dev \
    && go get github.com/prg3/kubewatch-telegram \
    && cp ${GOPATH}/bin/kubewatch /usr/local/bin \
    && apk del --purge dev && rm -rf /tmp/* /go \
	&& rm -rf /var/cache/apk/*

#------------------------------------------------------------------------------
# Entrypoint:
#------------------------------------------------------------------------------

ENTRYPOINT [ "kubewatch" ]
