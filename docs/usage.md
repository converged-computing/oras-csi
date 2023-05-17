# User Guide

## Setup

Before using the ORAS CSI Driver, you should read the instructions in [install](install.md).

## How does it work?

The driver can be configured by an administrator to customize control of saving paths and other metadata,
see the [install](install.md) documentation for that. Once you have the driver installed, you can
make storage requests for pods by way of `volumeAttributes` for it. Here is an example pod with
several options:


```yaml
kind: Pod
apiVersion: v1
metadata:
  name: my-csi-app-inline
spec:
  containers:
    - name: my-container
      image: ubuntu
      volumeMounts:
      - name: oras-inline
        mountPath: "/mnt/oras"
        readOnly: true
      command: ["sleep", "1000000"]
  volumes:
    - name: oras-inline
      csi:
        driver: csi.oras.land
        readOnly: true
        volumeAttributes:
          oras.artifact.reference: "ghcr.io/singularityhub/github-ci:latest"
          oras.options.plainhttp: true
          oras.options.insecure: true
```

In the above, we want to retrieve an artifact from the GitHub container registry, and extract
the layers (or blobs) to `/mnt/oras` in the container. Although this particular image doesn't require it,
we've added a few options for demonstration purposes only.

## Options

The full list of options you can define is provided in the table below.

| Name | Description | Required | Default |
|------|-------------|---------|----------|
| oras.artifact.reference | The artifact unique resource identifier (URI) | true | unset |
| oras.artifact.layers.mediatype | one or more (comma separated) media types to filter (include) | false | unset |
| oras.options.plainhttp | Allow pull of an artifact using plain http | false | false |
| oras.options.insecure | Allow insecure pull of an artifact (not implemented yet) | false | false |
| oras.options.concurrency | Concurrency to use for oras handler download | false | 1 |
| oras.options.pullalways | Always pull the artifact files afresh | false | false |

Of the above, the only required is the reference. When you provide only a reference, the entire
artifact will be extracted. We will have more examples and tests for each of the above soon!

## Examples

The following full examples are provided alongside the repository:

 - [basic/pod](https://github.com/converged-computing/oras-csi/tree/main/examples/basic/pod): the tutorial here, along with a basic mediaType filter
 - [kind](https://github.com/converged-computing/oras-csi/tree/main/examples/kind): build and deploy to a local registry using kind
 - [java](https://github.com/converged-computing/oras-csi/tree/main/examples/java): assemble a functioning Java app from base and app images

For each of the examples above, follow the respective links.
Here we will provide a basic example to create a pod with a mounted artifact! We will add more
tutorials as needed. Regardless of how you apply the configs during [install](install.md), you can see the containers (plugins on each node) as follows:

```bash
$ kubectl get pods -n kube-system | grep csi-oras
```

To see logs for the `oras-csi-plugin` (usually for debugging) since it's a container in a pod, ask to see them:

```bash
$ kubectl logs -n kube-system csi-oras-node-pkdwh csi-oras-plugin -f
```
```console
time="2023-04-11T23:59:29Z" level=info msg="Preparing artifact cache (mode: node; node-id: minikube; root-dir: /; plugin-data-dir: pv_data)"
time="2023-04-11T23:59:29Z" level=info msg="NewNodeService creation (rootDir /, pluginDataDir pv_data, nodeId minikube, mountPointsCount 1)"
time="2023-04-11T23:59:29Z" level=info msg="Setting up ORAS Logging. ORAS path: /pv_data/logs"
time="2023-04-11T23:59:29Z" level=info msg="ORAS Logging set up!"
time="2023-04-11T23:59:29Z" level=info msg="StartService - endpoint unix:///csi/csi.sock"
time="2023-04-11T23:59:29Z" level=info msg=CreategRPCServer
time="2023-04-11T23:59:29Z" level=info msg="CreateListener - endpoint unix:///csi/csi.sock"
time="2023-04-11T23:59:29Z" level=info msg="CreateListener - Removing socket /csi/csi.sock"
time="2023-04-11T23:59:29Z" level=info msg="StartService - Registering node service"
time="2023-04-11T23:59:29Z" level=info msg="StartService - Starting to serve!"
time="2023-04-11T23:59:30Z" level=info msg=GetPluginInfo
time="2023-04-11T23:59:32Z" level=info msg=NodeGetInfo
```

Or you can use the [Kubernetes app label](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/) for "part-of" to filter pods:

```bash
$ kubectl get pods -l 'app.kubernetes.io/part-of=csi-driver-oras' --all-namespaces
```

And finally, see the storage class (that we've made the default):

```bash
$ kubectl get storageclass
NAME                     PROVISIONER                RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
oras-storage (default)   csi.oras.land              Delete          Immediate           true                   6m49s
standard (default)       k8s.io/minikube-hostpath   Delete          Immediate           false                  21d
```

We plan to make this customizable when we add a helm chart!

## Testing

Let's try creating a pod that has a volume using our storage class:

```bash
$ kubectl apply -f examples/basic/pod/pod.yaml
```

Note that if you check the output of the csi plugin now, it should be a lot more verbose.
I like to keep this command running in a second terminal to watch the logs as they appear!

```bash
$ kubectl logs -n kube-system csi-oras-node-bb7mw csi-oras-plugin -f
```

<details>

<summary>More verbose output</summary>

```console
time="2023-05-17T09:59:58Z" level=info msg="Preparing artifact cache (mode: node; node-id: oras-csi-control-plane; root-dir: /; plugin-data-dir: pv_data enforce-namespaces: true)"
time="2023-05-17T09:59:58Z" level=info msg="NewNodeService creation (rootDir /, pluginDataDir pv_data, nodeId oras-csi-control-plane, handlersCount 1)"
time="2023-05-17T09:59:58Z" level=info msg="Setting up ORAS Logging. ORAS path: /pv_data/logs"
time="2023-05-17T09:59:58Z" level=info msg="ORAS Logging set up!"
time="2023-05-17T09:59:58Z" level=info msg="StartService - endpoint unix:///csi/csi.sock"
time="2023-05-17T09:59:58Z" level=info msg=CreategRPCServer
time="2023-05-17T09:59:58Z" level=info msg="CreateListener - endpoint unix:///csi/csi.sock"
time="2023-05-17T09:59:58Z" level=info msg="CreateListener - Removing socket /csi/csi.sock"
time="2023-05-17T09:59:58Z" level=info msg="StartService - Registering node service"
time="2023-05-17T09:59:58Z" level=info msg="StartService - Starting to serve!"
time="2023-05-17T09:59:58Z" level=info msg=GetPluginInfo
time="2023-05-17T09:59:58Z" level=info msg=NodeGetInfo
time="2023-05-17T10:01:49Z" level=info msg="NodePublishVolume - VolumeId: csi-b2e6bcaddfcc84d4434e8bc126bf2a225009e95cef2215e57a1b4e1f277bf900, Readonly: true, VolumeCapability mount:<> access_mode:<mode:SINGLE_NODE_WRITER >  TargetPath /var/lib/kubelet/pods/44bb7b41-7854-4bdf-a2c5-d3552aee35f5/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-05-17T10:01:49Z" level=info msg="Looking for volume context...."
time="2023-05-17T10:01:49Z" level=info msg="Oras - container: ghcr.io/singularityhub/github-ci, target: /mnt/oras-csi-control-plane"
time="2023-05-17T10:01:49Z" level=info msg="Enforce namespaces: true"
time="2023-05-17T10:01:49Z" level=info msg="Enforcing artifact namespace to be under default"
time="2023-05-17T10:01:49Z" level=info msg="Remote repository ghcr.io/singularityhub/github-ci:latest will be proxied by /pv_data/default/ghcr-io-singularityhub-github-ci-latest"
time="2023-05-17T10:01:49Z" level=info msg="Resolving manifest descriptor for ghcr.io/singularityhub/github-ci:latest"
time="2023-05-17T10:01:50Z" level=info msg="Fetching manifest {\"MediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"Digest\":\"sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867\",\"Size\":402}"
time="2023-05-17T10:01:50Z" level=info msg="Uncached fetching : {\"MediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"Digest\":\"sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867\",\"Size\":402}"
time="2023-05-17T10:01:50Z" level=info msg="Pulling sha256:acb1ec674e686f4ba7a0e5c0ce1d41b6c2a5f5f1b9b9baca9c612f794faa3f8e, 1 of 1"
time="2023-05-17T10:01:50Z" level=info msg="Uncached fetching : {\"MediaType\":\"application/vnd.sylabs.sif.layer.v1.sif\",\"Digest\":\"sha256:acb1ec674e686f4ba7a0e5c0ce1d41b6c2a5f5f1b9b9baca9c612f794faa3f8e\",\"Size\":798720}"
time="2023-05-17T10:01:50Z" level=info msg="OCI: Writing sha256:acb1ec674e686f4ba7a0e5c0ce1d41b6c2a5f5f1b9b9baca9c612f794faa3f8e to /pv_data/default/ghcr-io-singularityhub-github-ci-latest/container.sif"
time="2023-05-17T10:01:50Z" level=info msg="Oras artifact root: /pv_data/default/ghcr-io-singularityhub-github-ci-latest"
time="2023-05-17T10:01:50Z" level=info msg="Found artifact asset: container.sif"
time="2023-05-17T10:01:50Z" level=info msg="volume source directory:/pv_data/default/ghcr-io-singularityhub-github-ci-latest"
time="2023-05-17T10:01:50Z" level=info msg="volume target directory:/var/lib/kubelet/pods/44bb7b41-7854-4bdf-a2c5-d3552aee35f5/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-05-17T10:01:50Z" level=info msg="volume options:[ro]"
```
Notice that the `Uncached fetching` indicates that manifest and sif layer blob are both pulled from the remote registry.

</details>

Note that by default, `enforceNamespaces` is set to true, meaning that the artifacts are stored on the driver and organized by namespace.
If you disable this mode (not the default of the driver), the artifacts will be shared across namespaces, something you should do with
caution only if you trust users and applications between them. The diff below shows the change when you disable the default.

```diff
- time="2023-04-16T23:31:43Z" level=info msg="volume source directory:/pv_data/default/ghcr-io-singularityhub-github-ci-latest"
+ time="2023-04-16T23:31:43Z" level=info msg="volume source directory:/pv_data/ghcr-io-singularityhub-github-ci-latest"
```

Since this is an ephermal volume, this means that we don't create PVC or PV, we simply ask for what we need
in the pod. Note that the pod wants to pull the artifact `ghcr.io/singularityhub/github-ci:latest`
and provide the contents (a container.sif in this case) to a volume mounted at `/mnt/oras':

```bash
kind: Pod
apiVersion: v1
metadata:
  name: my-csi-app-inline
spec:
  containers:
    - name: my-container
      image: ubuntu
      volumeMounts:
      - name: oras-inline
        mountPath: "/mnt/oras"
        readOnly: true
      command: [ "sleep", "1000000" ]
  volumes:
    - name: oras-inline
      csi:
        driver: csi.oras.land
        readOnly: true
        volumeAttributes:
          oras.artifact.reference: "ghcr.io/singularityhub/github-ci:latest"
```

Let's shell into the pod and see if it's there!

```bash
$ kubectl exec -it my-csi-app-inline -- bash
```
```console
root@my-csi-app-inline:/# ls /mnt/oras/
container.sif
```

Yay! Try copying your pod into a second (identical one) and creating a mount of the same container at a different location
(we have provided this second pod file to make it easy):

```bash
$ kubectl apply -f examples/basic/pod/second-pod.yaml
$ kubectl exec -it my-second-csi-app-inline -- bash
```
```console
root@my-second-csi-app-inline:/# ls /mnt/second-oras/
container.sif
```

Importantly, in the logs we see an indication that the manifest and sif blob are pulled from the OCI layout cache (our original goal): `Cached fetching :`

```console
time="2023-05-17T10:03:11Z" level=info msg="Remote repository ghcr.io/singularityhub/github-ci:latest will be proxied by /pv_data/default/ghcr-io-singularityhub-github-ci-latest"
time="2023-05-17T10:03:11Z" level=info msg="Manifest cached for ghcr.io/singularityhub/github-ci:latest"
time="2023-05-17T10:03:11Z" level=info msg="Fetching manifest {\"MediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"Digest\":\"sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867\",\"Size\":402}"
time="2023-05-17T10:03:11Z" level=info msg="Cached fetching : {\"MediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"Digest\":\"sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867\",\"Size\":402}"
time="2023-05-17T10:03:11Z" level=info msg="Pulling sha256:acb1ec674e686f4ba7a0e5c0ce1d41b6c2a5f5f1b9b9baca9c612f794faa3f8e, 1 of 1"
time="2023-05-17T10:03:11Z" level=info msg="Cached fetching : {\"MediaType\":\"application/vnd.sylabs.sif.layer.v1.sif\",\"Digest\":\"sha256:acb1ec674e686f4ba7a0e5c0ce1d41b6c2a5f5f1b9b9baca9c612f794faa3f8e\",\"Size\":798720}"
time="2023-05-17T10:03:11Z" level=info msg="OCI: Writing sha256:acb1ec674e686f4ba7a0e5c0ce1d41b6c2a5f5f1b9b9baca9c612f794faa3f8e to /pv_data/default/ghcr-io-singularityhub-github-ci-latest/container.sif"
time="2023-05-17T10:03:11Z" level=info msg="Oras artifact root: /pv_data/default/ghcr-io-singularityhub-github-ci-latest"
time="2023-05-17T10:03:11Z" level=info msg="Found artifact asset: container.sif"
time="2023-05-17T10:03:11Z" level=info msg="volume source directory:/pv_data/default/ghcr-io-singularityhub-github-ci-latest"
time="2023-05-17T10:03:11Z" level=info msg="volume target directory:/var/lib/kubelet/pods/eff62b75-3449-470c-af17-c4ba06adcf41/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-05-17T10:03:11Z" level=info msg="volume options:[ro]"
time="2023-05-17T10:03:11Z" level=info msg="BindMount - source: /pv_data/default/ghcr-io-singularityhub-github-ci-latest, target: /var/lib/kubelet/pods/eff62b75-3449-470c-af17-c4ba06adcf41/volumes/kubernetes.io~csi/oras-inline/mount, options: [ro]"
time="2023-05-17T10:03:11Z" level=info msg="mount -o bind /pv_data/default/ghcr-io-singularityhub-github-ci-latest /var/lib/kubelet/pods/eff62b75-3449-470c-af17-c4ba06adcf41/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-05-17T10:03:11Z" level=info msg="Successfully mounted /pv_data/default/ghcr-io-singularityhub-github-ci-latest to /var/lib/kubelet/pods/eff62b75-3449-470c-af17-c4ba06adcf41/volumes/kubernetes.io~csi/oras-inline/mount"
```

Let's try testing that the artifact remains persistent on the node and delete both pods, and also
test deletion. Deletion should run an unmount command, but should not delete anything from our actual node!
To delete:

```bash
$ kubectl delete -f examples/basic/pod/pod.yaml
$ kubectl delete -f examples/basic/pod/second-pod.yaml
```

These RPC actions hit the "NodeUnpublishVolume" endpoint and (I've noticed) at least for MiniKube they take
a hot minute to run. When it's run you'll see:

```console
time="2023-04-12T20:49:29Z" level=info msg="NodeUnpublishVolume - VolumeId: csi-990169a3156cb7579912440f451f5ff18d46b71d775aa25e7fac2698e0971955, TargetPath: /var/lib/kubelet/pods/ce9055f0-02e3-493f-9bf2-203a96a9d051/volumes/kubernetes.io~csi/oras-inline/mount)"
time="2023-04-12T20:49:29Z" level=info msg="BindUMount - target: /var/lib/kubelet/pods/ce9055f0-02e3-493f-9bf2-203a96a9d051/volumes/kubernetes.io~csi/oras-inline/mount"
```

We can actually sanity check that the artifacts still persist on the plugin as follows:

```bash
$ kubectl exec -it -n kube-system csi-oras-node-ghnkj -c csi-oras-plugin -- bash
```
```console
root@minikube:/# ls /pv_data/
ghcr-io-singularityhub-github-ci-latest
root@minikube:/# ls /pv_data/ghcr-io-singularityhub-github-ci-latest/
container.sif
```

Now we can theoretically create a pod again, and that same container.sif should be used (you should see the message about it already existing):

```bash
$ kubectl apply -f examples/basic/pod/pod.yaml
```
```console
time="2023-05-17T10:03:11Z" level=info msg="Manifest cached for ghcr.io/singularityhub/github-ci:latest"
time="2023-05-17T10:03:11Z" level=info msg="Fetching manifest {\"MediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"Digest\":\"sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867\",\"Size\":402}"
time="2023-05-17T10:03:11Z" level=info msg="Cached fetching : {\"MediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"Digest\":\"sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867\",\"Size\":402}"
time="2023-05-17T10:03:11Z" level=info msg="Pulling sha256:acb1ec674e686f4ba7a0e5c0ce1d41b6c2a5f5f1b9b9baca9c612f794faa3f8e, 1 of 1"
time="2023-05-17T10:03:11Z" level=info msg="Cached fetching : {\"MediaType\":\"application/vnd.sylabs.sif.layer.v1.sif\",\"Digest\":\"sha256:acb1ec674e686f4ba7a0e5c0ce1d41b6c2a5f5f1b9b9baca9c612f794faa3f8e\",\"Size\":798720}"
```

And that's it! We very likely should have an attribute that specifies for this to be cleaned up and re-created. E.g., you
can imagine using the same reference that has changed in the registry that you want to pull again.

Finally, clean up the whole thing and pod again!

```bash
$ kubectl delete -f examples/basic/pod/pod.yaml
$ make clean
```
