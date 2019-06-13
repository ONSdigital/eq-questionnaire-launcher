#!/usr/bin/env bash

set -e

if [[ -z "$RUNNER_URL" ]]; then
  echo "RUNNER_URL is mandatory"
  exit 1
fi

DOCKER_REGISTRY="${DOCKER_REGISTRY:-eu.gcr.io/census-eq-ci}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

helm tiller run \
    helm upgrade --install \
    survey-launcher \
    k8s/helm \
    --set surveyRunnerUrl=https://${RUNNER_URL} \
    --set image.repository=${DOCKER_REGISTRY}/eq-questionnaire-launcher \
    --set image.tag=${IMAGE_TAG}
