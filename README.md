# CSI ORAS Driver

This repository is a test to create a CSI driver into one that uses [ORAS](https://oras.land) to
generate a cluster-level cache of artifacts. The use case I have in mind is for Singularity containers, which can be run via workflows.

## Background

A CSI or ["Container storage interface"](https://github.com/container-storage-interface/spec) is a [standard plugin](https://github.com/container-storage-interface/spec/blob/master/spec.md) that we
can instantiate to make it easy for others to use it in their Kubernetes clusters. At least, this is my current level of understanding. 
What I want to get working is an emphemeral, inline plugin that sits alongside a node and can keep a local cache of artifacts
(e.g., containers or data) to be used between different pod runs (e.g., an indexed job).

For documentation, see our early [docs](docs) folder.

## Planning

These are ideas / features that would be nice to have:

 - standards for defining assets and actions to take (e.g., copy this binary here, this directory here, get this annotation and do X, name this file Y)
 - an ability to specify and enforce some permissions namespace for the artifacts (e.g., for multi-tenant cluster)
 - an ability to specify conditions for cleaning up an artifact when the kubernetes pod / object is deleted
 - Set a much higher [upper limit](https://github.com/oras-project/oras-go/blob/e8225cb1e125bd4c13d6b586ae6d862050c3fae2/registry/remote/repository.go#L98-L102) for the artifact size. 
 - Allow predictable naming for an artifact
 - Proper testing of the CSI

This is a pretty good start for a quick prototype!

## Thanks

I looked at a lot of examples to figure out the basic architecture I wanted to use. Here are the ones that I liked the design for:

- [csi-inline-volume](https://kubernetes.io/blog/2022/08/29/csi-inline-volumes-ga/)
- [moosefs-csi](https://github.com/moosefs/moosefs-csi) is what I used to learn and template the design here, also under an Apache 2.0 license. This is Copyright of Saglabs SA.
