# Post Install

Regardless of how you apply the configs during [install](install.md), you can see the containers (plugins on each node) as follows:

```bash
$ kubectl get pods -n kube-system | grep csi-oras
```

To see logs for the `oras-csi-plugin` (usually for debugging) since it's a container in a pod, ask to see them:

```bash
$ kubectl logs -n kube-system  csi-oras-node-pkdwh csi-oras-plugin -f
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
$ kubectl apply -f examples/kubernetes/pod/pod.yaml
```

Note that if you check the output of the csi plugin now, it should be a lot more verbose.
I like to keep this command running in a second terminal to watch the logs as they appear!

```bash
$ kubectl logs -n kube-system csi-oras-node-bb7mw csi-oras-plugin -f
```

<details>

<summary>More verbose output</summary>

```console
time="2023-04-12T20:31:46Z" level=info msg="Preparing artifact cache (mode: node; node-id: minikube; root-dir: /; plugin-data-dir: pv_data)"
time="2023-04-12T20:31:46Z" level=info msg="NewNodeService creation (rootDir /, pluginDataDir pv_data, nodeId minikube, mountPointsCount 1)"
time="2023-04-12T20:31:46Z" level=info msg="Setting up ORAS Logging. ORAS path: /pv_data/logs"
time="2023-04-12T20:31:46Z" level=info msg="ORAS Logging set up!"
time="2023-04-12T20:31:46Z" level=info msg="StartService - endpoint unix:///csi/csi.sock"
time="2023-04-12T20:31:46Z" level=info msg=CreategRPCServer
time="2023-04-12T20:31:46Z" level=info msg="CreateListener - endpoint unix:///csi/csi.sock"
time="2023-04-12T20:31:46Z" level=info msg="CreateListener - Removing socket /csi/csi.sock"
time="2023-04-12T20:31:46Z" level=info msg="StartService - Registering node service"
time="2023-04-12T20:31:46Z" level=info msg="StartService - Starting to serve!"
time="2023-04-12T20:31:46Z" level=info msg=GetPluginInfo
time="2023-04-12T20:31:47Z" level=info msg=NodeGetInfo
time="2023-04-12T20:32:43Z" level=info msg="NodePublishVolume - VolumeId: csi-997d0afca658f39939fafc20ffaf7b059be2f70940b594cba6cbfde715670fc1, Readonly: true, VolumeContext map[container:ghcr.io/singularityhub/github-ci:latest csi.storage.k8s.io/ephemeral:true csi.storage.k8s.io/pod.name:my-csi-app-inline csi.storage.k8s.io/pod.namespace:default csi.storage.k8s.io/pod.uid:77268dbb-b6a8-4b73-a580-677d4eb93178 csi.storage.k8s.io/serviceAccount.name:default], PublishContext map[], VolumeCapability mount:<> access_mode:<mode:SINGLE_NODE_WRITER >  TargetPath /var/lib/kubelet/pods/77268dbb-b6a8-4b73-a580-677d4eb93178/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-04-12T20:32:43Z" level=info msg="Looking for volume context...."
time="2023-04-12T20:32:43Z" level=info msg="map[container:ghcr.io/singularityhub/github-ci:latest csi.storage.k8s.io/ephemeral:true csi.storage.k8s.io/pod.name:my-csi-app-inline csi.storage.k8s.io/pod.namespace:default csi.storage.k8s.io/pod.uid:77268dbb-b6a8-4b73-a580-677d4eb93178 csi.storage.k8s.io/serviceAccount.name:default]"
time="2023-04-12T20:32:43Z" level=info msg="Oras - container: ghcr.io/singularityhub/github-ci:latest, target: /mnt/minikube"
time="2023-04-12T20:32:43Z" level=info msg="Artifact root does not exist, creating/pv_data/ghcr-io-singularityhub-github-ci-latest"
time="2023-04-12T20:32:43Z" level=info msg="Found ORAS container: ghcr.io/singularityhub/github-ci:latest"
time="2023-04-12T20:32:43Z" level=info msg="Creating oras filestore at: /pv_data/ghcr-io-singularityhub-github-ci-latest"
time="2023-04-12T20:32:43Z" level=info msg="Preparing to pull from remote repository: ghcr.io/singularityhub/github-ci"
time="2023-04-12T20:32:44Z" level=info msg="Oras artifact root: /pv_data/ghcr-io-singularityhub-github-ci-latest"
time="2023-04-12T20:32:44Z" level=info msg="Found artifact asset: container.sif"
time="2023-04-12T20:32:44Z" level=info msg="volume source directory:/pv_data/ghcr-io-singularityhub-github-ci-latest"
time="2023-04-12T20:32:44Z" level=info msg="volume target directory:/var/lib/kubelet/pods/77268dbb-b6a8-4b73-a580-677d4eb93178/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-04-12T20:32:44Z" level=info msg="volume options:[ro]"
time="2023-04-12T20:32:44Z" level=info msg="BindMount - source: /pv_data/ghcr-io-singularityhub-github-ci-latest, target: /var/lib/kubelet/pods/77268dbb-b6a8-4b73-a580-677d4eb93178/volumes/kubernetes.io~csi/oras-inline/mount, options: [ro]"
time="2023-04-12T20:32:44Z" level=info msg="mount -o bind /pv_data/ghcr-io-singularityhub-github-ci-latest /var/lib/kubelet/pods/77268dbb-b6a8-4b73-a580-677d4eb93178/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-04-12T20:32:44Z" level=info msg="Successfully mounted /pv_data/ghcr-io-singularityhub-github-ci-latest to /var/lib/kubelet/pods/77268dbb-b6a8-4b73-a580-677d4eb93178/volumes/kubernetes.io~csi/oras-inline/mount"
```

</details>

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
          container: "ghcr.io/singularityhub/github-ci:latest"
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
$ kubectl apply -f examples/kubernetes/pod/second-pod.yaml
$ kubectl exec -it my-second-csi-app-inline -- bash
```
```console
root@my-second-csi-app-inline:/# ls /mnt/second-oras/
container.sif
```

Importantly, in the logs we see an indication that the container was not re-pulled (our original goal): `Artifact root already exists, no need to re-create!`

```console
time="2023-04-12T20:34:37Z" level=info msg="Artifact root already exists, no need to re-create!"
time="2023-04-12T20:34:37Z" level=info msg="Oras artifact root: /pv_data/ghcr-io-singularityhub-github-ci-latest"
time="2023-04-12T20:34:37Z" level=info msg="Found artifact asset: container.sif"
time="2023-04-12T20:34:37Z" level=info msg="volume source directory:/pv_data/ghcr-io-singularityhub-github-ci-latest"
time="2023-04-12T20:34:37Z" level=info msg="volume target directory:/var/lib/kubelet/pods/eda9a3b5-b6ce-41c5-84d2-ea7a1ea677bb/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-04-12T20:34:37Z" level=info msg="volume options:[ro]"
time="2023-04-12T20:34:37Z" level=info msg="BindMount - source: /pv_data/ghcr-io-singularityhub-github-ci-latest, target: /var/lib/kubelet/pods/eda9a3b5-b6ce-41c5-84d2-ea7a1ea677bb/volumes/kubernetes.io~csi/oras-inline/mount, options: [ro]"
time="2023-04-12T20:34:37Z" level=info msg="mount -o bind /pv_data/ghcr-io-singularityhub-github-ci-latest /var/lib/kubelet/pods/eda9a3b5-b6ce-41c5-84d2-ea7a1ea677bb/volumes/kubernetes.io~csi/oras-inline/mount"
time="2023-04-12T20:34:37Z" level=info msg="Successfully mounted /pv_data/ghcr-io-singularityhub-github-ci-latest to /var/lib/kubelet/pods/eda9a3b5-b6ce-41c5-84d2-ea7a1ea677bb/volumes/kubernetes.io~csi/oras-inline/mount"
```

Let's try testing that the artifact remains persistent on the node and delete both pods, and also
test deletion. Deletion should run an unmount command, but should not delete anything from our actual node!
To delete:

```bash
$ kubectl delete -f examples/kubernetes/pod/pod.yaml
$ kubectl delete -f examples/kubernetes/pod/second-pod.yaml
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
$ kubectl apply -f examples/kubernetes/pod/pod.yaml
```
```console
time="2023-04-12T20:52:16Z" level=info msg="Artifact root already exists, no need to re-create!"
```

And that's it! We very likely should have an attribute that specifies for this to be cleaned up and re-created. E.g., you
can imagine using the same reference that has changed in the registry that you want to pull again.

Finally, clean up the whole thing and pod again!

```bash
$ kubectl delete -f examples/kubernetes/pod/pod.yaml
$ make clean
```
