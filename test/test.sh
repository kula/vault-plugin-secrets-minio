#!/bin/bash

PLUGIN_NAME=vault-plugin-secrets-minio
PLUGIN_PATH=minio
TOPDIR="$( git rev-parse --show-toplevel )"
PLUGIN="${TOPDIR}/${PLUGIN_NAME}"
WORK="${TOPDIR}/_workspace"
PLUG_DIR="${WORK}/plugins"
MINIO_DATA="${WORK}/minio-data"
TEST="${TOPDIR}/test"
VAULT_PID_FILE="${WORK}/vault.pid"
VAULT_LOG_FILE="${WORK}/vault.log"
MINIO_PID_FILE="${WORK}/minio.pid"
MINIO_LOG_FILE="${WORK}/minio.log"

. ${TEST}/env.sh


die() { 
    echo "### Error: $@" >&2
    exit 1
}

cleanup() {
    for PID_FILE in "${VAULT_PID_FILE}" "${MINIO_PID_FILE}"; do
	if [ -f "$PID_FILE" ]; then
	    kill -INT $(cat "$PID_FILE") >/dev/null 2>&1
	    rm -f "${PID_FILE}" >/dev/null 2>&1
	fi
    done
}

trap cleanup EXIT


which vault >/dev/null || die "Cannot find vault binary in your path"
which minio >/dev/null || die "Cannot fine minio binary in your path"
which mc >/dev/null || die "Cannot find mc binary in your path"

set -x

[[ -x "$PLUGIN" ]] || die "$PLUGIN_NAME doesn't exist, make it?"
mkdir -p "${PLUG_DIR}" || die "Cannot make work directory"
mkdir -p "${MINIO_DATA}" || die "Cannot make minio data dir"
rm -f "${PLUG_DIR}/${PLUGIN_NAME}" || die "Cannot delete old plugin"
rm -rf "${MINIO_DATA}/*" "${MINIO_DATA}/.minio.sys" || die "Cannot clean out old minio data"

echo "### Starting minio server"
nohup minio server --address ${MINIO_IP}:${MINIO_PORT} \
    "${MINIO_DATA}" >> "${MINIO_LOG_FILE}" 2>&1 &
echo $! > "${MINIO_PID_FILE}"

ps -p $(cat ${MINIO_PID_FILE} ) > /dev/null || die "Could not start minio"
echo 
echo "### Minio started"
echo "### Log file: ${MINIO_LOG_FILE}"
echo "### PID file: ${MINIO_PID_FILE}"
echo

echo "### Starting vault server"

nohup vault server -dev \
    -dev-listen-address="${VAULT_IP}:${VAULT_PORT}" \
    -dev-root-token-id="${VAULT_TOKEN}" \
    -dev-plugin-dir="${PLUG_DIR}" \
    -log-level=debug >> "${VAULT_LOG_FILE}" 2>&1 &
echo $! > "${VAULT_PID_FILE}"

ps -p $(cat ${VAULT_PID_FILE} ) >/dev/null || die "Could not start vault"

echo "### Vault started"
echo "### Log file: ${VAULT_LOG_FILE}"
echo "### Pid file: ${VAULT_PID_FILE}"
echo
echo "### Copying and registering plugin"

cp "${TOPDIR}/${PLUGIN_NAME}" "${PLUG_DIR}" || die "Cannot copy ${PLUGIN_NAME} into plugin dir"
SUM=$( sha256sum ${PLUG_DIR}/${PLUGIN_NAME} 2>/dev/null | cut -d " " -f 1 )
[[ -n "$SUM" ]] || die "Could not calculate plugin sha256 sum"

vault plugin register \
    -command="${PLUGIN_NAME}" \
    -sha256="${SUM}" \
    ${PLUGIN_NAME} || die "Could not register plugin in vault"

vault secrets enable \
    -path=${PLUGIN_PATH} \
    -plugin-name=${PLUGIN_NAME} \
    plugin || die "Could not enable plugin"


echo
echo "Plugin enabled at ${PLUGIN_NAME}/"
echo "Starting shell, exit to stop vault server"
echo
echo

PS1="vault-testing: " /bin/bash --noprofile --norc
