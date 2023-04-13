# Use Cases for ORAS CSI

ORAS, or OCI registry as storage, makes it easy to store artifacts in an OCI registry. The artifacts can range from binaries to text files or other associated assets. In layman’s terms, it’s a way to store data, organized by content type, under a common unique resource identifier (URI).

## Container Storage Interface

A container storage interface (CSI) is a Kubernetes plugin ([spec](https://github.com/container-storage-interface/spec/blob/master/spec.md)) that makes it possible to mount custom storage interfaces to a container orchestration system such as Kubernetes. While CSIs typically mount an external, persistent filesystem, object storage, or block storage and expect a request for a persistent volume claim and volumes, a particular kind of new "ephemeral" CSI ([see csi inline volumes](https://kubernetes.io/blog/2022/08/29/csi-inline-volumes-ga/)) allows for more creativity in the kind of bind. This is where we saw opportunity for ORAS CSI.

## ORAS CSI for mounting OCI artifacts

In that most use cases for storage are in the context of pods, the most basic desire for an ORAS CSI would be to mount an OCI artifact, as a single layer asset or entire directory, to a pod. Below we expand upon the use cases for which we think this would be wanted.

### Indexed Job Assets

A Kubernetes indexed job creates many pods that might need shared assets. An easy example is an indexed job that needs to run a Singularity container. If we were to pull this container to all pods once for one job, that might be a reasonable thing to do. But if we are to run the job 1000 times, we don’t want to unnecessarily stress the remote registry with N pulls (the number of jobs X pods per job). Instead, we pull the container once using the CSI driver, and then on subsequent jobs simply mount the already existing artifact to use. This could be further optimized, for example, if some central driver would pull it once and then distribute across nodes where pods are needed.

### Custom Image Assembly

In that we can imagine most pods, deployments, or indexed jobs are some base container with some set of assets needed at runtime, the ORAS csi driver would allow us to intelligently assemble images like blocks at ru time. We wouldn’t always need to pre-build every image dynamically - compatible binaries and general assets could be added on demand. A more robust spec for doing this (eg provided via the OCI working group) would be warranted.

### Componentized Patching

Applications may be composed of a base image and assets that may be executable binaries, libraries or configuration. These typically get bundled as a single image.  If a base image requires updating, the application must be rebuilt. Options like [`crane rebase`](https://github.com/google/go-containerregistry/blob/main/cmd/crane/doc/crane_rebase.md) take a similar approach. By composing a pod with two components, an image and the application or configuration artifact, updating the base image without rebuilding the binaries becomes possible, provided that the base image maintains [ABI compatibility](https://en.wikipedia.org/wiki/Application_binary_interface)

Another advantage of not having to rebuild the full application is improved deployment speed and efficiency. When only specific components of an application are updated, the time and resources required for building and deploying the changes are significantly reduced. This allows for faster application updates and reduces the overall time-to-market for new features and bug fixes.

Additionally, this approach minimizes the risk of introducing new issues or errors during the rebuild process. By only updating the necessary components, developers can maintain better control over the application's stability and ensure that the rest of the application remains unaffected by the changes.

#### Sample java app

The following application is composed from a java 

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