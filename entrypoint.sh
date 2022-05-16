#!/bin/bash

# this file was copied and modified from
# https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/blob/master/entrypoint.sh

# echo commands to the terminal output
set -ex

# Check whether there is a passwd entry for the container UID
myuid=$(id -u)
mygid=$(id -g)
# turn off -e for getent because it will return error code in anonymous uid case
set +e
uidentry=$(getent passwd $myuid)
set -e

echo $myuid
echo $mygid
echo $uidentry

# If there is no passwd entry for the container UID, attempt to create one
if [[ -z "$uidentry" ]] ; then
    if [[ -w /etc/passwd ]] ; then
        echo "$myuid:x:$myuid:$mygid:anonymous uid:/usr:/bin/false" >> /etc/passwd
    else
        echo "Container ENTRYPOINT failed to add passwd entry for anonymous UID"
    fi
fi

exec /usr/bin/spark-ui-reverse-proxy "$@"
