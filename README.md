# knoway

> An Envoy inspired, ultimate LLM-first gateway for LLM serving and downstream application developers and enterprises

## Description

Lite and easy dedicated Gateway with various of LLM specific optimizations and features. You can think of it as Nginx, but for LLMs, and upcoming supported models (such as Stable Diffusion, etc.).

## Features

- 💬 **LLM-first**: Designed for LLMs, with optimizations and features that are specific to LLMs.
- 🕸️ **Envoy Inspired**: Inspired by Envoy, similar architecture and features, if you are already familiar with Envoy, you will find it easy to use and understand this codebase.
- 💻 **Single command deployment**: Deploy the gateway with a single command, just like `nginx -c nginx.conf`.
- 🚢 **Kubernetes Native**: CRDs, control plane side implementations are batteries included, `helm install knoway` and you are ready to go.

Some of the LLMs specific optimizations and features include:

- 👷 **Serverless boot loader**: Able to boot up the upstream Pod of serving services on-demand, make LLM serving more cost-effective.
- ✅ **Fault tolerance**: Fault tolerance for LLMs, with the ability to retry, circuit breaking, etc. when dealing with external providers.
- 🚥 **Rate limiting**: Rate limiting based on tokens, prompts, etc., to protect the LLMs serving services from being abused.
- 📚 **Semantic Cache**: Cache based on the semantics of the prompts and tokens, CDN of the LLMs.
- 📖 **Semantic Route**: Route based on the difficulties, semantic meaning of prompts, etc., to make the LLMs serving services more efficient with right models.
- 🔍 **OpenTelemetry**: OpenTelemetry support, with the ability to trace the calls to LLMs, and the gateway itself.

## Getting Started

### Prerequisites

- `go` version v1.22.0+
- `docker` version 17.03+.
- `kubectl` version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

```console
$ export VERSION=v0.0.0-xxxx
$ helm repo add knoway https://release-ci.daocloud.io/chartrepo/knoway
$ helm repo update
$ helm upgrade --install --create-namespace -n knoway-system knoway knoway/knoway --version=${VERSION} \
    --set global.imageRegistry=release-ci.daocloud.io \
    --set global.config.auth_server_url="10.33.2.23:31960" \
    --set global.config.stats_server_url="10.33.2.23:31960"
```

### To Uninstall

```console
helm uninstall knoway -n knoway-system
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/knoway:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/knoway/<tag or branch>/dist/install.yaml
```

## Contributing

> [!NOTE]
> Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

