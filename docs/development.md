# Development

For the tutorial below, we are using minikube. Building with kind is not
supported yet, because I'm not sure we can test CSI Drivers outside of VMs. Thus, for the time being,
the commands assume you are using minikube! 

## Quick Start 

Create a minikube cluster:

```bash
$ minikube start
```

We have a single command to uninstall the plugin, build the binary, package it in
a container, and push to this registry. You'll want to use a registry you have write permission
to, and that minikube can pull from. First, change the image: directive
in the development yaml `deploy/dev-driver.yaml` to your image:

```diff
- image: ghcr.io/converged-computing/oras-csi-plugin:0.1.0:dev
+ image: ghcr.io/myusername/oras-csi-plugin:0.1.0:dev
```

You should only change the registry name, and not the repository name or tag.
And then to deploy (as an example):

```bash
$ make dev DOCKER_REGISTRY=ghcr.io/myusername
```

Note that this command is installing the "dev" version of the config in [deploy](https://github.com/converged-computing/oras-csi/tree/main/deploy),
and you can tweak values there to change defaults. And then proceed with the regular usage tutorial in [post install](usage.md).

## Build Helm

> **warning** These charts have only been tested with MiniKube

I used helmify to generate the original charts, and then did a lot of manual fixes to namespaces,
names, and other variables that I didn't want to be customized. This means that (for now) the helm charts need
to be manually updated if any changes are added to the core in `deploy`. If you do this, please make a backup first
and then run!

You can build new draft helm chart as follows:

```bash
$ mv ./charts ./chart
$ make helm
```
