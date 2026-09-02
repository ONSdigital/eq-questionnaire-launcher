# EQ Launcher

This project was copied from https://github.com/ONSdigital/go-launch-a-survey and should be used for v3 of runner.

## Building and Running
Install Go and ensure that your `GOPATH` env variable is set (usually it's `~/go`).

```
go get
go build
./eq-questionnaire-launcher

go run launch.go (Does both the build and run cmd above)
```

Open http://localhost:8000/

## Docker
Install Podman for your system as the container runtime, and make sure the Podman machine is
started every time you want to use container images:

```shell
podman machine start
```

Podman is API-compatible with Docker, so provide a `docker` command that points at it (the
`docker build`/`docker buildx`/`docker run` commands below rely on this):

```shell
mkdir -p ~/.local/bin
ln -s "$(which podman)" ~/.local/bin/docker
hash -r
```

`~/.local/bin` must be on your `PATH`. Verify:

```shell
docker --version
```

The dockerfile is a multistage dockerfile which can be built using:

```
docker build -t eq-questionnaire-launcher:latest .
```

You can also build for a specific platform using Docker’s extended build tool - buildx:

> Note: `podman` supports multi-platform builds via `podman build --platform`, but does not
> ship a `buildx` subcommand. If `docker buildx build` fails through the symlink, use
> `podman build --platform ...` directly instead of the commands below.

```
docker buildx build --platform linux/amd64 -t eq-questionnaire-launcher:latest .
```
Or for multiple platforms:

```
docker buildx build --platform linux/amd64,linux/arm64 -t eq-questionnaire-launcher:latest .
```

You can then run the image using `SURVEY_RUNNER_SCHEMA_URL` to point it at an instance of survey runner.

```
docker run -e SURVEY_RUNNER_SCHEMA_URL=http://localhost:5000 -it -p 8000:8000 onsdigital/eq-questionnaire-launcher:latest
```

The syntax for this will be slightly different on Mac

```
docker run -e SURVEY_RUNNER_SCHEMA_URL=http://host.docker.internal:5000 -it -p 8000:8000 onsdigital/eq-questionnaire-launcher:latest
```

You should then be able to access go launcher at `localhost:8000`

You can also run a Survey Register for launcher to load Schemas from

```
docker run -it -p 8080:8080 onsdigital/eq-survey-register:simple-rest-api
```

## Run Quick-Launch
If the schema specifies a `schema_name` field, that will be used as the schema_name claim. If not, the filename from the URL (before `.`) will be used.

Run Questionnaire Launcher
```
scripts/run_app.sh
```
Now run Go launcher and navigate to "http://localhost:8000/quick-launch?schema_url=" passing the url of the JSON
```
e.g."http://localhost:8000/quick-launch?schema_url=http://localhost:7777/1_0001.json"
```

The optional query parameter `version` can be added to the quick launch url which allows for the launch payload structure to be specified. If the parameter is not set then the default launch payload structure `v2` will be used.
Documentation on the `v2` structure can be found [here](https://github.com/ONSdigital/ons-schema-definitions/blob/v3/docs/rm_to_eq_runner_payload_v2.rst)
```
e.g."http://localhost:8000/quick-launch?schema_url=http://localhost:7777/1_0001.json&version=v1"
```

## Commands for Formatting & Linting

### Conda environment

Python, Node.js, Poetry, Go and golangci-lint are managed via the committed `environment.yml`,
matching `.python-version` and `.nvmrc` as closely as conda-forge availability allows:

```shell
conda env create -f environment.yml
conda activate eq-launcher
```

> Note: conda-forge does not publish every Node patch release.
> Where the exact `.nvmrc` version is unavailable, pin the closest available patch
> below it and note the substitution in `environment.yml`.

If `.python-version` or `.nvmrc` change, update `environment.yml` to match and run:

```shell
conda env update -f environment.yml --prune
```

### Poetry

Poetry must install into the conda environment rather than creating its own virtualenv. Set this
on the environment so no configuration file is left in the repository:

```shell
conda env config vars set POETRY_VIRTUALENVS_CREATE=false
conda deactivate && conda activate eq-launcher
```

Confirm it took effect: this must print `false`:

```shell
echo $POETRY_VIRTUALENVS_CREATE
```

With the environment active, install ESLint and Prettier for formatting and linting of static files:
``` shell
npm install
```
Install djLint for formatting and linting template files using:
```shell
poetry install
```

Before being able to run `lint-go`, ensure `golangci-lint` is available (installed via the conda
environment above; upgrade with `conda env update -f environment.yml --prune`).

| Command                 | Task                                                    |
|-------------------------|---------------------------------------------------------|
| `make format-static`    | Formats all static files (Javascipt and CSS)            |
| `make format-templates` | Formats all HTML files and shows the changes to be made |
| `make format-go`        | Formats all the Golang files                            |
| `make format`           | Formats all files listed above                          |
| `make lint-static`      | Lints all static files and reports any issues           |
| `make lint-templates`   | Lints all HTML files and reports any issues             |
| `make lint-go`          | Lints all Golang files using an external tool           |
| `make lint`             | Lints all files listed above                            |


## Design System
To update the design system version, you need to update the version within the CDN link, they are present in both template files ([layout](templates/layout.html:11) and [launch](templates/launch.html:381))

## Notes
* There are no unit tests yet
* JWT spec based on http://ons-schema-definitions.readthedocs.io/en/latest/jwt_profile.html

## Settings

| Environment Variable           | Meaning                                                             | Default                                                                |
|--------------------------------|---------------------------------------------------------------------|------------------------------------------------------------------------|
| GO_LAUNCH_A_SURVEY_LISTEN_HOST | Host address to listen on                                           | 0.0.0.0                                                                |
| GO_LAUNCH_A_SURVEY_LISTEN_PORT | Host port to listen on                                              | 8000                                                                   |
| SURVEY_RUNNER_URL              | URL of Questionnaire Runner to re-direct to when launching a survey | http://localhost:5000                                                  |
| SURVEY_REGISTER_URL            | URL of eq-survey-register to load schema list from                  | http://localhost:8080                                                  |
| SDS_API_BASE_URL               | URL of the SDS API to fetch supplementary data from                 | http://localhost:5003                                                  |
| JWT_ENCRYPTION_KEY_PATH        | Path to the JWT Encryption Key (PEM format)                         | jwt-test-keys/sdc-user-authentication-encryption-sr-public-key.pem     |
| JWT_SIGNING_KEY_PATH           | Path to the JWT Signing Key (PEM format)                            | jwt-test-keys/sdc-user-authentication-signing-launcher-private-key.pem |
| OIDC_TOKEN_BACKEND             | The backend to use when fetching the Open ID Connect token          | gcp                                                                    |
| OIDC_TOKEN_VALIDITY_IN_SECONDS | The time in seconds an OIDC token is valid                          | 3600                                                                   |
| OIDC_TOKEN_LEEWAY_IN_SECONDS   | The leeway to use when validating OIDC tokens                       | 300                                                                    |
| SDS_OAUTH2_CLIENT_ID           | The OAuth2 Client ID used when setting up IAP on the SDS            |                                                                        |
| CIR_OAUTH2_CLIENT_ID           | The OAuth2 Client ID used when setting up IAP on the CIR            |                                                                        |
| SDS_ENABLED_IN_ENV             | Signifies if the SDS service is enabled in the environment          | true                                                                   |
