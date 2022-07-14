#!/bin/sh

days=64

echo "generating credentials and private key for server and client..."
openssl genrsa -out ./certs/server.key 2048

openssl req -new -x509 -days ${days} \
  -subj "/C=GB/L=Russia/O=NWS-server/CN=localhost" \
  -key ./certs/server.key -out ./certs/server.crt

openssl genrsa -out ./certs/client.key 2048

openssl req -new -x509 -days ${days} \
  -subj "/C=GB/L=Russia/O=NWS-client/CN=localhost" \
  -key ./certs/client.key -out ./certs/client.crt