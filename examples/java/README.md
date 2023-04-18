# Componentized Java application example 

The java appliation is a simple java application that will use the oras-csi driver to mount a container image as a volume.
The application creates a simple web listening on port 8080 and return "Hello World" when you access the root path.

The sample uses a local registry `localhost:5001` to push the image to. You can change the registry in the `Makefile` and `pod.yaml` files.
For setup using kind follow [our kind setup example](../kind). WHen you've created the kind cluster and installed the driver, e.g.,:

```bash
```

# Build and push the base image

```
make build-base-image
make push-base-image
```

## Build and push the java application

```
make build-app
make push-app
```

## Deploy the driver

```
kubectl -f ./pod.yaml
```

You should be able to query if the application is running 

```
$ kubectl get pods/my-csi-java-app
NAME              READY   STATUS    RESTARTS   AGE
my-csi-java-app   1/1     Running   0          11m
```

```
$ kubectl exec -it my-csi-java-app -- /bin/sh -c 'wget -q -O - http://localhost:8080 2>&1'
Hello, World!%
```

