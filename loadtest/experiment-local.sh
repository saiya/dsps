#!/bin/bash -e
if [  $# -le 1 ]; then
  echo "Usage: $0 EXPERIMENT_DIR OUTPUT_DIR"
  echo "    EXPERIMENT_DIR: directory of the experiment, must have dsps.config.yaml file"
  exit 1
fi

EXPERIMENT_DIR="$(realpath "$1")"
OUTPUT_DIR="$(realpath "$2")"
cd "$(dirname "$0")"
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

# Ensure $1 points valid directory
if [ ! -f "${EXPERIMENT_DIR}/dsps.config.yaml" ]; then
  echo "DSPS config file not found: ${EXPERIMENT_DIR}/dsps.config.yaml" >&2
  exit 2
fi

# Show some limits
ulimit -a
sysctl kern.ipc.somaxconn || true

export REDIS1_PORT=$(($(($RANDOM%1000))+10000))
export REDIS2_PORT=$(($(($RANDOM%1000))+11000))
DSPS_PORT=$(($(($RANDOM%1000))+12000))

LOG_DIR="${OUTPUT_DIR}/logs"
RESULT_DIR="${OUTPUT_DIR}/result"
test -d "${LOG_DIR}" && rm -r "${LOG_DIR}"
mkdir -p "${LOG_DIR}"
test -d "${RESULT_DIR}" && rm -r "${RESULT_DIR}"
mkdir -p "${RESULT_DIR}"

echo "Starting redis servers... (port: ${REDIS1_PORT}, ${REDIS2_PORT})"
redis-server --port ${REDIS1_PORT} > "${LOG_DIR}/redis1.log" 2>&1 &
redis-server --port ${REDIS2_PORT} > "${LOG_DIR}/redis2.log" 2>&1 &

echo "Starting DSPS server... (port: ${DSPS_PORT})"
pushd ../server
cat "${EXPERIMENT_DIR}/dsps.config.yaml" | envsubst '${REDIS1_PORT} ${REDIS2_PORT}' | \
  go run main.go --port ${DSPS_PORT} --dump-config - > "${LOG_DIR}/dsps.log" 2>&1 &
popd

echo "Check DSPS server readiness..."
until curl "http://localhost:${DSPS_PORT}/probe/readiness"
do
  echo "Waiting until DSPS server get ready..."
  sleep 1
done

echo "Starting loadtest..."
BASE_URL="http://localhost:${DSPS_PORT}" \
    k6 run \
      "--summary-export=${RESULT_DIR}/summary.json" \
      --out "json=${RESULT_DIR}/data.json" \
      --summary-trend-stats="min,avg,med,max,p(90),p(95),p(99)" \
      ./loadtest.k6.js \
    | tee "${LOG_DIR}/k6.log"

echo "Experiment completed."

echo "Compressing logs..."
gzip "${LOG_DIR}"/*.log
gzip "${RESULT_DIR}/data.json"
