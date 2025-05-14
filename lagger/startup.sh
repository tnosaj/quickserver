#!/bin/bash

if [[ "unset" == "${DEST:-unset}" ]]; then
  echo "unset DEST env var - exiting"
  exit 1
else
  # DEST=foo:80
  if [[ $DEST =~ ":" ]]; then
    arrIN=(${DEST//:/ })
    IP=echo ${arrIN[0]}
    DESTINATION_PORT=echo ${arrIN[1]}
    # destination ip (connect was a hostname of a docker container)
    DESTINATION_IP=$(dig $DEST +short)
  else
    DESTINATION_PORT=8080
    DESTINATION_IP=$(dig $DEST +short)
  fi
fi

if [[ "unset" == "${DELAY:-unset}" ]]; then
  DELAY=100ms
else
  # DELAY=100ms-10ms
  if [[ $DELAY =~ "-" ]]; then
    DELAY="${DELAY//-/ }"
  fi
fi

echo "Running for $DESTINATION_IP:$DESTINATION_PORT - with deplay: $DELAY"


# traffic control to create 100ms delay
tc qdisc add dev eth0 root netem delay "$DELAY"

# socat multipurpose relay
socat tcp-listen:8080,reuseaddr,fork tcp:$DESTINATION_IP:$DESTINATION_PORT


# traffic control to create 100ms delay
# tc qdisc add dev eth0 root netem delay 100ms

# change to random delay
# tc qdisc change dev eth0 root netem delay 100ms 10ms

# delete 
# tc qdisc del dev eth0 root netem

# add 250ms delay
# tc qdisc add dev eth0 root netem delay 250ms

# add random delay
# tc qdisc add dev eth0 root netem delay 100ms 10ms