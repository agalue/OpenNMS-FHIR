#!/bin/bash -e

join() {
  local IFS="$1"; shift; echo "$*";
}

get_key() {
  echo "$1" | cut -d_ -f2- | tr '[:upper:]' '[:lower:]' | tr _ .
}

IFS=$'\n'
CONSUMER=("acks=1")
for VAR in $(env)
do
  env_var=$(echo "$VAR" | cut -d= -f1)
  if [[ $env_var =~ ^CONSUMER_ ]]; then
    echo "[configuring consumer] processing $env_var"
    key=$(get_key $env_var)
    echo "[configuring consumer] key: $key"
    val=${!env_var}
    echo "[configuring consumer] value: $val"
    CONSUMER+=("$key=$val")
  fi
done

exec /eventhub-forwarder \
  -bootstrap "${BOOTSTRAP_SERVERS}" \
  -source-topic "${SOURCE_TOPIC}" \
  -group-id "${GROUP_ID-opennms}" \
  -connection-str "${EVENT_HUB_CONNECTION_STR}" \
  -consumer-params "$(join , ${CONSUMER[@]})"
