#!/usr/bin/bash

source ./scripts/apigen.mk

for i in ${API_LIST}
do
    IFS=','
    set -- $i # convert the "tuple" into the param args $1 $2...
    # echo $2
    # echo $2
    oapi-codegen --config ${CFGDIR}/$1 ${APIDIR}/$2
done

