#!/bin/bash

RELATED_IMAGE_ARCH=''
RELATED_IMAGE_URL=''
RELATED_IMAGE_DIGEST=''
RELATED_IMAGE_STATE=0
grep -A 2 --no-group-separator 'RELATED_IMAGE' $1 | while read -r RELATED_IMAGE_LINE ; do

    if [ ${RELATED_IMAGE_STATE} -eq 0 ]; then
        if [[ "${RELATED_IMAGE_LINE}" =~ .*ppc64le ]]; then
            RELATED_IMAGE_ARCH='ppc64le'
        elif [[ "${RELATED_IMAGE_LINE}" =~ .*s390x ]]; then
            RELATED_IMAGE_ARCH='s390x'
        else
            RELATED_IMAGE_ARCH='amd64'
        fi
        RELATED_IMAGE_STATE=1
    elif [ ${RELATED_IMAGE_STATE} -eq 1 ]; then
        RELATED_IMAGE_URL="$(grep -Po 'registry.redhat.io.*' <<< ${RELATED_IMAGE_LINE})"
        RELATED_IMAGE_STATE=2
    elif [ ${RELATED_IMAGE_STATE} -eq 2 ]; then
        RELATED_IMAGE_DIGEST="$(grep -Po 'sha256:.*' <<< ${RELATED_IMAGE_LINE})"
        echo "Inspecting ${RELATED_IMAGE_ARCH} ${RELATED_IMAGE_URL} ${RELATED_IMAGE_DIGEST}"
        RELATED_IMAGE_REGISTRY_DIGEST="$(skopeo inspect --override-arch ${RELATED_IMAGE_ARCH} docker://${RELATED_IMAGE_URL} | jq -r '.Digest')"

        if [ "${RELATED_IMAGE_DIGEST}" != "${RELATED_IMAGE_REGISTRY_DIGEST}" ]; then
            echo "Digest mismatch: ${RELATED_IMAGE_DIGEST}/${RELATED_IMAGE_REGISTRY_DIGEST}"
        fi
        RELATED_IMAGE_STATE=0
    fi
done
