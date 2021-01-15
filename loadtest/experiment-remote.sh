#!/bin/bash -e
if [  $# -le 1 ]; then
  echo "Usage: $0 EXPERIMENT_DIR OUTPUT_DIR"
  echo "    EXPERIMENT_DIR: directory of the experiment, must have dsps.config.yaml file"
  exit 1
fi

# Variables required by loadtest script
for name in BASE_URL CHANNEL_ID_PREFIX
do
  if [ -z "${!name}" ]; then
    echo "Environment variable ${name} required."
    exit 1
  fi
done

EXPERIMENT_DIR="$(realpath "$1")"
OUTPUT_DIR="$(realpath "$2")"
cd "$(dirname "$0")"
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT

# Create output directories
LOG_DIR="${OUTPUT_DIR}/logs"
RESULT_DIR="${OUTPUT_DIR}/result"
test -d "${LOG_DIR}" && rm -r "${LOG_DIR}"
mkdir -p "${LOG_DIR}"
test -d "${RESULT_DIR}" && rm -r "${RESULT_DIR}"
mkdir -p "${RESULT_DIR}"
exec > >(tee -a "${LOG_DIR}/experiment-local.log") 2>&1  # Save all logs after this

# Show some limits
ulimit -a
sysctl kern.ipc.somaxconn || true

# Load test
echo "Starting loadtest... ($(LC_ALL=C date))"
time k6 run \
      "--summary-export=${RESULT_DIR}/summary.json" \
      --out "json=${RESULT_DIR}/data.json" \
      --summary-trend-stats="min,avg,med,max,p(90),p(95),p(99)" \
      ./loadtest.k6.js \
    | tee "${LOG_DIR}/k6.log"
echo "Experiment completed ($(LC_ALL=C date))."

echo "Compressing logs..."
gzip "${LOG_DIR}"/*.log
gzip "${RESULT_DIR}/data.json"
