# Deployment with [Concourse](https://concourse-ci.org/)

To deploy this application with Concourse, you must have a Kubernetes cluster and be logged in to a Concourse instance.

---

The following environment variables must be set when deploying the app:
- PROJECT_ID
- DOCKER_REGISTRY
- RUNNER_URL

The following are optional variables that can also be set if needed:
- SERVICE_ACCOUNT_JSON
- REGION
- IMAGE_TAG

To deploy to a cluster you can run the following command

```sh
PROJECT_ID=<project_id> \
DOCKER_REGISTRY=<docker_registry> \
RUNNER_URL=<runner_instance_url> \
fly -t <target_concourse_instance> execute \
  --config ci/deploy.yaml
```