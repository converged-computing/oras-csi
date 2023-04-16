# Install

Note that you need the `--allow-privileged=true` flag set for both API server and kubelet (default value for kubelet is `true`).
Also note that this is currently being developed on minikube, and we haven't gotten it working on kind (or interacting with
a local registry yet).

## Options

Whether you install via the helm chart or the included yaml configs, the following options are available to customize the driver:

| Name | Description | Default |
|------|-------------|---------|
| csi_root_dir | root directory for all claims | "/"     |                   
| driver_working_dir | ORAS directory (relative to csi_root_dir) for all driver data (effective working dir will be calculated as csi_root_dir/driver_working_dir | "pv_data" |
| handlers_count | Number of handlers for each node | "1" |
| enforce_namespaces | Enforce unique artifacts across namespaces (tradeoff between node storage space and security) | "true" | 
| oras_logging| Should driver log to csi_root_dir/driver_working_dir/logs directory? |  "true" |

## Helm Install

To install, you can use [helm](https://helm.sh): 

```bash
$ git clone https://github.com/converged-computing/oras-csi
$ cd oras-csi
$ helm install oras-csi ./charts
```

Note that for helm, you can see the values for the chart as follows:

```bash
$ helm show values ./charts
```

And then set any of them for an install:

```bash
$ helm install --set config.orasLogging="false" oras-csi ./charts
```

Or you can install directly from GitHub packages (an OCI registry):

```
# helm prior to v3.8.0
$ export HELM_EXPERIMENTAL_OCI=1
$ helm pull oci://ghcr.io/converged-computing/oras-csi-helm/chart
```
```console
Pulled: ghcr.io/converged-computing/oras-csi-helm/chart:0.1.0
```

And install!

```bash
$ helm install oras-oci chart-0.1.0.tgz 
```
```console
NAME: oras-csi
LAST DEPLOYED: Wed Apr 12 22:41:08 2023
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

## Config Install

You can also install from a config in [deploy](../deploy).

```bash
$ make install
```

which is the equivalent of:

```bash
kubectl apply -f deploy/driver-csi-oras.yaml
kubectl apply -f deploy/csi-oras-config.yaml
```

or the development config (along with your own kind cluster and registry) to build and deploy it first.
You'll want to change the `image` registry to be one you control.

```bash
kubectl apply -f deploy/dev-driver.yaml
kubectl apply -f deploy/csi-oras-config.yaml
```

Either way, if you have the repository handy, you can customize the [deploy/csi-oras-config.yaml](../deploy/csi-oras-config.yaml) to 
change any defaults. When you are done, see [post install usage](usage.md) for interacting with your driver
and using it.
