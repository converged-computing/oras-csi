# CSI ORAS Driver

This repository is a test to create a CSI driver into one that uses [ORAS](https://oras.land) to
generate a cluster-level cache of artifacts. Read about the [use cases](docs/use-cases.md) or jump in!

```console
	██████╗ ██████╗  █████╗ ███████╗       ██████╗███████╗██╗
	██╔═══██╗██╔══██╗██╔══██╗██╔════╝      ██╔════╝██╔════╝██║
	██║   ██║██████╔╝███████║███████╗█████╗██║     ███████╗██║
	██║   ██║██╔══██╗██╔══██║╚════██║╚════╝██║     ╚════██║██║
	╚██████╔╝██║  ██║██║  ██║███████║      ╚██████╗███████║██║
	 ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝       ╚═════╝╚══════╝╚═╝
```

## Background

A CSI or ["Container storage interface"](https://github.com/container-storage-interface/spec) is a [standard plugin](https://github.com/container-storage-interface/spec/blob/master/spec.md) that we
can instantiate to make it easy for others to use it in their Kubernetes clusters. At least, this is my current level of understanding. 
What I want to get working is an emphemeral, inline plugin that sits alongside a node and can keep a local cache of artifacts
(e.g., containers or data) to be used between different pod runs (e.g., an indexed job).

For documentation, see our early [docs](docs) folder.

## TODO

 - [ ] add more kubernetes app labels?
 - [ ] make pretty, branded docs!
 - [ ] test with kind, write up tutorial (https issue too)
 - [ ] add concept of cleanup (on level of CSIDriver and pod?)
 - [ ] what about more customization to pull (e.g., pull if newer?)
 - [ ] everything must be tested, with tested examples
 - [ ] custom naming / locations for mount? Or should be handled by app?
 - [ ] better levels / control for logging

## Planning

These are ideas / features that would be nice to have:

 - standards for defining assets and actions to take (e.g., copy this binary here, this directory here, get this annotation and do X, name this file Y)
 - in addition to namespace, some other permissions / security features?
 - ability to add pull secrets to artifacts
 - an ability to specify conditions for cleaning up an artifact when the kubernetes pod / object is deleted
 - Allow predictable naming for an artifact

This is a pretty good start for a quick prototype!

## Thanks

I looked at a lot of examples to figure out the basic architecture I wanted to use. Here are the ones that I liked the design for:

- [csi-inline-volume](https://kubernetes.io/blog/2022/08/29/csi-inline-volumes-ga/)
- [moosefs-csi](https://github.com/moosefs/moosefs-csi) is what I used to learn and template the design here, also under an Apache 2.0 license. This is Copyright of Saglabs SA.


## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
