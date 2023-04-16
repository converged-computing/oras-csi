# About

ORAS, or OCI registry as storage, makes it easy to store artifacts in an OCI registry. The artifacts can range from binaries to text files or other associated assets. In layman’s terms, it’s a way to store data, organized by content type, under a common unique resource identifier (URI).

## Frequently Asked Questions

### What is a Container Storage Interface?

A container storage interface (CSI) is a Kubernetes plugin ([spec](https://github.com/container-storage-interface/spec/blob/master/spec.md)) that makes it possible to mount custom storage interfaces to a container orchestration system such as Kubernetes. While CSIs typically mount an external, persistent filesystem, object storage, or block storage and expect a request for a persistent volume claim and volumes, a particular kind of new "ephemeral" CSI ([see csi inline volumes](https://kubernetes.io/blog/2022/08/29/csi-inline-volumes-ga/)) allows for more creativity in the kind of bind. This is where we saw opportunity for ORAS CSI.

### What does ORAS OCI allow?

[ORAS](https://oras.land) or "OCI Registry as Storage" is a set of tools that makes it easy to interact with the contents of [OCI registries](https://oras.land/#what-are-oci-registries). In layman's terms, these registries can hold artifacts that range from binaries to associated assets that you might need for software. Thus, the most basic desire for an ORAS CSI would be to make it easy to retrieve and mount an OCI artifact. An easy example might be injecting the latest version of a compatible binary into a container.

## Use Cases

In this section, we expand upon the use cases for which we think this would be wanted.

### Indexed Job Assets

A Kubernetes indexed job creates many pods that might need shared assets. An easy example is an indexed job that needs to run a Singularity container. If we were to pull this container to all pods once for one job, that might be a reasonable thing to do, albeit if we have a lot of pods we might be stressing the registry. We can do better to pull it once to a shared location, as [demonstrated in this tutorial for the Flux Operator](https://flux-framework.org/flux-operator/tutorials/singularity.html). That works well if we run the job once. However, if we are to run the job 1000 times, we don’t want to unnecessarily stress the remote registry with J (number of jobs) x P (number of pods) pulls. This is where ORAS-OCI could help. Instead, we pull the container once using the CSI driver, likely during the run of the first job, and then on subsequent jobs simply mount the already existing artifact to use. Here is an example yaml manifest that would do exactly this - mounting an artifact with a known `container.sif` that we can run with Singularity:

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: my-singularity-app
spec:
  containers:
    - name: singularity-runner
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
```

You could easily [install oras](https://oras.land/cli/) and test pulling and shelling into this basic container:

```bash
$ oras pull
...
```
```bash
$ singularity run container.sif 
Hold me closer... tiny container :) :D
```

In the above, we use the [oras cli](https://github.com/oras-project/oras) to pull, and the driver takes a similar approach, but using the [oras go sdk](https://github.com/oras-project/oras-go). If you are interested in this use case (running a container inside of a Kubernetes pod, which is possible with Singularity and appropriate permissions) you can use [this repostiory template](https://github.com/singularityhub/github-ci/) to reproduce the above container.

### Custom Image Assembly

In that we can imagine most pods, deployments, or indexed jobs are some base container with some set of assets needed at runtime, the ORAS csi driver would allow us to intelligently assemble images like blocks at runtime. We wouldn’t always need to pre-build every image dynamically - compatible binaries and general assets could be added on demand. A more robust spec for doing this (eg provided via the OCI working group) would be warranted.
As a simple example, Nextflow has a filesystem named [Fusion](https://seqera.io/fusion/) that is mounted into running user containers. They use [wave](https://seqera.io/wave/) as a quasi build service to do a custom build beforehand that can be pulled with the binary added. With ORAS-CSI, we wouldn't need to do a separate build - we would simply package the binary as an ORAS artifact, and mount it to the pods when requested.

### Componentized Patching

Applications may be composed of a base image and assets that may be executable binaries, libraries or configuration. These typically get bundled as a single image.  If a base image requires updating, the application must be rebuilt. Options like [`crane rebase`](https://github.com/google/go-containerregistry/blob/main/cmd/crane/doc/crane_rebase.md) take a similar approach. By composing a pod with two components, an image and the application or configuration artifact, updating the base image without rebuilding the binaries becomes possible, provided that the base image maintains [ABI compatibility](https://en.wikipedia.org/wiki/Application_binary_interface)

Another advantage of not having to rebuild the full application is improved deployment speed and efficiency. When only specific components of an application are updated, the time and resources required for building and deploying the changes are significantly reduced. This allows for faster application updates and reduces the overall time-to-market for new features and bug fixes.

Additionally, this approach minimizes the risk of introducing new issues or errors during the rebuild process. By only updating the necessary components, developers can maintain better control over the application's stability and ensure that the rest of the application remains unaffected by the changes.

#### Sample java app

The following application is composed from a java base image, and the application artifact is added by the CSI.

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: my-csi-app-inline
spec:
  containers:
    - name: my-java-app
      image: docker.io/library/openjdk:8-jre-alpine
      command: [ "java", "-jar", "/app/HelloWorldServer.jar" ]
      volumeMounts:
      - mountPath: "/app"
        name: oras-inline
  volumes:
    - name: oras-inline
      csi:
        driver: csi.oras.land
        readOnly: true
        volumeAttributes:
          oras.artifact.reference: "registry/my-java-app-artifact:v1"
```
