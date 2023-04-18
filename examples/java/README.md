# Componentized Java application example 

The java appliation is a simple java application that will use the oras-csi driver to mount a container image as a volume.
The application creates a simple web listening on port 8080 and return "Hello World" when you access the root path.

The sample uses a local registry `localhost:5001` to push the image to. You can change the registry in the `Makefile` and `pod.yaml` files.
For setup using kind follow [our kind setup example](../kind). When you've created the kind cluster and installed the driver, e.g.,:

```bash
# run from ../kind
$ ./kind-with-registry.sh 

# from the root of the repository
$ make dev-helm DOCKER_REGISTRY=localhost:5001
```

proceed to the next steps to build and push your Java app!

# Build and push the base image

Here is how to build the base image, and push to your local registry:

```bash
$ make build-base-image
$ make push-base-image
```

## Build and push the java application

Now let's do the same. This "app" image has the component we want to add onto the base image.

```bash
$ make build-app
$ make push-app
```

## Create the Pod

Creating the pod will make a request for storage from the driver, and we will be mounting `/app`
from the app image into the base image.

```bash
$ kubectl apply -f ./pod.yaml
```

You should be able to query if the application is running 

```bash
$ kubectl get pods/my-csi-java-app
NAME              READY   STATUS    RESTARTS   AGE
my-csi-java-app   1/1     Running   0          11m
```

And then to do a request to the app endpoint. If it's successfully running (java from the base
image running the app) you'll get the "Hello World!" response.

```bash
$ kubectl exec -it my-csi-java-app -- /bin/sh -c 'wget -q -O - http://localhost:8080 2>&1'
Hello, World!
```

And that's it! This demonstrates that we can successfully add the `app` layer onto the base
image and result in a functioning application. If we were to use this layer many times (e.g.,
across an indexed job or similar) the artifact would only need to be pulled once.

When you are done, cleanup!

```bash
$ kubectl delete -f pod.yaml
$ kind delete cluster --name oras
$ docker stop kind-registry
$ docker rm kind-registry
```