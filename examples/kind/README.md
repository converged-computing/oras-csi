# Run ORAS CSI driver in kind

Create the kind cluster with the a local registry using the following script 

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

## Update oras-csi deployment

```bash
helm install --set node.csiOrasPlugin.image.repository="localhost:5001/oras-csi-plugin" \
               --set node.csiOrasPlugin.image.tag="latest" \
               --set node.csiOrasPlugin.imagePullPolicy="Always" \
               --set config.orasLogging="true" oras-csi ./charts
```

## Deploying the driver

Make dev will use the kind cluster as the context and build and push to your local registry at `localhost:5001` 

```
make dev DOCKER_REGISTRY=localhost:5001
``` 

You can validate if your images are in the registry 

```shell
$ oras repo tags localhost:5001/oras-csi-plugin
latest
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
time="2023-04-17T23:23:46Z" level=info msg="Preparing artifact cache (mode: node; node-id: oras-control-plane; root-dir: /; plugin-data-dir: pv_data enforce-namespaces: %!s(bool=true))"
time="2023-04-17T23:23:46Z" level=info msg="NewNodeService creation (rootDir /, pluginDataDir pv_data, nodeId oras-control-plane, handlersCount 1)"
t
```

## Pushing your images 

Images in the local kind registry can be reference using the `kind-registry:5000` as the registry host. 

```shell
❯ oras copy ghcr.io/singularityhub/github-ci:latest localhost:5001/github-ci:latest
Copying acb1ec674e68 container.sif
Copied  acb1ec674e68 container.sif
Copied [registry] ghcr.io/singularityhub/github-ci:latest => [registry] localhost:5001/github-ci:latest
Digest: sha256:5d6742ff0b10c1196202765dafb43275259bcbdbd3868c19ba1d19476c088867 
```

- Deploy the sample pod and mount the artifact from local registry

```shell
kubectl apply -f ./pod.yaml
```


