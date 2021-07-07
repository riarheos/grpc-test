#! /bin/bash

set -uex

protoc -I $GOPATH/pkg/mod/github.com/googleapis/googleapis\@v0.0.0-20210703104452-ac5d0130755d \
       -I . \
       --go_out . \
       --go_opt paths=source_relative \
       --go-grpc_out . \
       --go-grpc_opt paths=source_relative \
       --grpc-gateway_out . \
       --grpc-gateway_opt logtostderr=true \
       --grpc-gateway_opt paths=source_relative \
       --grpc-gateway_opt generate_unbound_methods=true \
       proto/greet.proto