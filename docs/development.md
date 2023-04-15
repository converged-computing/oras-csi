# Development

For the tutorial below, we are using minikube, e.g., `minikube start`. A kind tutorial will come soon.

## Quick Start 

After developing local files, we have a single command to uninstall the plugin, build the binary, package it in
a container, and push to a registry. You should check the `Makefile` first an ensure that you are targeting
a registry you have push permissions for!

```bash
$ make dev
```

and then the command installs the configs in [deploy](deploy). You should inspect this logic in the [Makefile](Makefile) first.

## Detailed Start

Usually you'll want to develop local files, and then build the binary, package the container,
and push to a registry:

```bash
$ make
$ make build-dev
$ make push-dev
```

When you are ready, apply the configuration file:

```bash
$ kubectl apply -f deploy/kubernetes/csi-oras-config.yaml
```

This is the config map if you want to inspect it:

```bash
$ kubectl describe cm -n kube-system csi-oras-config 
$ kubectl get cm -n kube-system csi-oras-config 
NAME              DATA   AGE
csi-oras-config   5      20s
```

Then deploy the driver plugin with the CSI sidecar containers:

```bash
$ kubectl apply -f deploy/kubernetes/csi-oras.yaml
```

And then proceed with the regular usage tutorial in [post install](usage.md).


## Build Helm

You can build the helm chart as follows (recommended to delete first)

```bash
$ rm -rf ./chart
$ make helm
```
