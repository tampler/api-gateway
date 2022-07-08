#!/usr/bin/bash

source ./scripts/services.mk

for i in ${SERVICE_LIST}
do
    echo "Updating service $i"
    cd internal/$i; go get -u; go mod tidy --compat=1.18; cd ../..
done

