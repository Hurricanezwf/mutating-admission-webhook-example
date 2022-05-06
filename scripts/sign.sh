#!/bin/bash

set -x

#KUBECMD="kubectl121 --kubeconfig=$HOME/.kube/config.k121"
KUBECMD="kubectl"

rm server*

cat <<EOF | cfssl genkey - | cfssljson -bare server
{
  "hosts": [
    "webhook.default.svc.cluster.local",
    "webhook.default.pod.cluster.local",
    "webhook.default.svc",
    "192.0.2.24",
    "10.0.34.2"
  ],
  "CN": "system:node:webhook.default.pod.cluster.local",
  "key": {
    "algo": "ecdsa",
    "size": 256
  },
  "names": [
    {
      "O": "system:nodes"
    }
  ]
}
EOF


${KUBECMD} delete csr webhook.default

cat <<EOF | ${KUBECMD} apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: webhook.default
spec:
  request: $(cat server.csr | base64 | tr -d '\n')
  signerName: kubernetes.io/kubelet-serving
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF


${KUBECMD} certificate approve webhook.default


sleep 5


${KUBECMD} get csr webhook.default -o jsonpath='{.status.certificate}' | base64 --decode > server.crt


${KUBECMD} delete secret webhook-tls-secret
${KUBECMD} create secret tls webhook-tls-secret --cert=server.crt --key=server-key.pem


echo "caBundle is, please paste to the csr!"
${KUBECMD} config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}'
