#!/usr/bin/bash
outpath="./"

# Generate Cloud Control
protoc  -I. \
					--go_out=${outpath} \
					--go-grpc_out=${outpath} \
					protos/*.proto


# Generate all services
protoc  -I. \
					--go_out=${outpath} \
					--go-grpc_out=${outpath} \
					protos/services/*.proto
					