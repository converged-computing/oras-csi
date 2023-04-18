# Run ORAS CSI in kind

This small tutorial will show you how to get ORAS CSI working in Kind.
We've provided the script [kind-with-registry.sh](kind-with-registry.sh)
that will create the kind cluster with a local registry:

```shell 
❯ ./kind-with-registry.sh
Creating cluster "oras" ...
 ✓ Ensuring node image (kindest/node:v1.25.3) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-oras"
You can now use your cluster with:

kubectl cluster-info --context kind-oras

Thanks for using kind! 😊
configmap/local-registry-hosting created
```

At this point you will have:

 - A kind cluster to mimic Kubernetees
 - A local OCI registry (aka Docker Registry) running on port 5001

And the next steps are to install ORAS CSI, which we can do with [helm](https://helm.sh/docs/intro/install/).

## Build and Deploy the ORAS-OCI driver

Now that we have a locally running registry, let's build the driver and push the resulting container to it!

```bash
$ make dev-helm DOCKER_REGISTRY=localhost:5001
```

After build you should see the image push to `localhost:5001/oras-csi-plugin:0.1.0-dev`,
and then (based on a custom command to helm) we install from there. You can expand
the details section below to see an example helm install command.

<details>

<summary>Example helm install</summary>

This is a derivation of the command run in the `Makefile` that installs the driver to your
cluster, customizing it to use your image pushed to you local registry.
This is provided for illustration purposes.

```bash
helm install --set node.csiOrasPlugin.image.repository="localhost:5001/oras-csi-plugin" \
              --set node.csiOrasPlugin.image.tag="latest" \
              --set node.csiOrasPlugin.imagePullPolicy="Always" \
              --set config.orasLogging="true" oras-csi ./charts
```

You can choose to uninstall `make helm-uninstall` and redeploy via the command above,
if you so choose.

</details>

You can validate if your development image is in the registry 

```shell
$ oras repo tags localhost:5001/oras-csi-plugin
0.1.0-dev
```

Also validate if your pods are running 

```shell
$ kubectl logs --follow \
    $(kubectl get pods -l 'app.kubernetes.io/part-of=csi-driver-oras' \
        --all-namespaces -o jsonpath='{.items[*].metadata.name}') \
    -c csi-oras-plugin \
    -n kube-system

time="2023-04-17T23:23:46Z" level=info
time="2023-04-17T23:23:46Z" level=info msg="██████╗ ██████╗  █████╗ ███████╗       ██████╗███████╗██╗"
time="2023-04-17T23:23:46Z" level=info msg="██╔═══██╗██╔══██╗██╔══██╗██╔════╝      ██╔════╝██╔════╝██║"
time="2023-04-17T23:23:46Z" level=info msg="██║   ██║██████╔╝███████║███████╗█████╗██║     ███████╗██║"
time="2023-04-17T23:23:46Z" level=info msg="██║   ██║██╔══██╗██╔══██║╚════██║╚════╝██║     ╚════██║██║"
time="2023-04-17T23:23:46Z" level=info msg="╚██████╔╝██║  ██║██║  ██║███████║      ╚██████╗███████║██║"
time="2023-04-17T23:23:46Z" level=info msg=" ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝       ╚═════╝╚══════╝╚═╝"
time="2023-04-17T23:23:46Z" level=info msg="Preparing artifact cache (mode: node; node-id: oras-control-plane; root-dir: /; plugin-data-dir: pv_data enforce-namespaces: true)"
time="2023-04-17T23:23:46Z" level=info msg="NewNodeService creation (rootDir /, pluginDataDir pv_data, nodeId oras-control-plane, handlersCount 1)"
...
```

## Pushing your images 

Images in the local kind registry can be reference using the `kind-registry:5000` as the registry host. 
Let's copy an image there from the GitHub container registry:

```shell
❯ oras copy ghcr.io/singularityhub/github-ci:latest localhost:5001/github-ci:latest
Copying acb1ec674e68 container.sif
Copied  acb1ec674e68 container.sif
Copied [registry] ghcr.io/singularityhub/github-ci:latest => [registry] localhost:5001/github-ci:latest
Digest: sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867 
```

And now deploy the sample pod and mount the artifact from local registry

```shell
❯ kubectl apply -f ./pod.yaml
```
And then check to see if the container.sif provided by the artifact is there:

```bash
$ kubectl exec -it my-csi-app-inline-on-kind -- ls /mnt/oras
container.sif
```

Success! When you are done, clean everything up.

```bash
$ docker stop kind-registry
$ docker rm kind-registry # note you can add the --rm flag in the script to do this automatically
$ kind delete cluster --name oras
```
