#!/usr/bin/env bash
set -euxo pipefail

if [[ -z "$PROJECT_ID" ]]; then
  echo "PROJECT_ID not provided"
  exit 1
fi

if [[ -z "$RUNNER_URL" ]]; then
  echo "RUNNER_URL not provided"
  exit 1
fi

if [[ -z "$DOCKER_REGISTRY" ]]; then
  echo "DOCKER_REGISTRY not provided"
  exit 1
fi

IMAGE_TAG="${IMAGE_TAG:=latest}"
REGION="${REGION:=europe-west2}"
CONCURRENCY="${CONCURRENCY=250}"
MIN_INSTANCES="${MIN_INSTANCES:=1}"
MAX_INSTANCES="${MAX_INSTANCES:=1}"
CPU="${CPU=1}"
MEMORY="${MEMORY=256M}"

gcloud beta run deploy eq-questionnaire-launcher \
    --project="${PROJECT_ID}" --region="${REGION}" --concurrency="${CONCURRENCY}" --min-instances="${MIN_INSTANCES}" --max-instances="${MAX_INSTANCES}" \
    --port=8000 --cpu="${CPU}" --memory="${MEMORY}" \
    --image="${DOCKER_REGISTRY}/eq-questionnaire-launcher:${IMAGE_TAG}" --platform=managed --allow-unauthenticated \
    --set-env-vars SURVEY_RUNNER_URL="${RUNNER_URL}"
