#!/usr/bin/env bats

# This is run relative to the root
EXAMPLES_DIR=examples/basic/pod
SLEEP_TIME=10

@test "basic pod test" {
  kubectl apply -f ${EXAMPLES_DIR}/pod.yaml
  sleep ${SLEEP_TIME}
  kubectl describe pods --all-namespaces
  kubectl wait --for=condition=ready pod -l job-name=oras-test --timeout=1200s

  # This test is looking inside our pod to the requested mount point
  kubectl exec -t my-csi-app-inline -- ls /mnt/oras | grep container.sif
  pod=$(kubectl get -n kube-system pods -o json | jq -r .items[].metadata.name | grep csi)
  echo pod is ${pod}

  # These tests are looking inside the csi driver pod!
  # Top level has the namespace of the pod
  kubectl exec -t -n kube-system ${pod} -c csi-oras-plugin -- ls /pv_data | grep default
  
  # Next level is the container URI (repository name + tag)
  kubectl exec -t -n kube-system ${pod} -c csi-oras-plugin -- ls /pv_data/default | grep ghcr-io-singularityhub-github-ci-latest

  # Next level is the container URI (repository name + tag)
  kubectl exec -t -n kube-system ${pod} -c csi-oras-plugin -- ls /pv_data/default | grep ghcr-io-singularityhub-github-ci-latest

  # And then the container.sif
  kubectl exec -t -n kube-system ${pod} -c csi-oras-plugin -- ls /pv_data/default/ghcr-io-singularityhub-github-ci-latest | grep container.sif
  kubectl delete -f ${EXAMPLES_DIR}/pod.yaml
}


