if use `oidc` auth, you need create secret `kube-oidc-proxy-config` and `oidc-server-root-secret`, the data is same with `u4a-system`
```bash
kubectl get secret kube-oidc-proxy-config -n u4a-system -o json \
 | jq 'del(.metadata["namespace","creationTimestamp","resourceVersion","selfLink","uid"])' \
 | kubectl apply -n baas-system -f -

kubectl get secret oidc-server-root-secret -n u4a-system -o json \
 | jq 'del(.metadata["namespace","creationTimestamp","resourceVersion","selfLink","uid"])' \
 | kubectl apply -n baas-system -f -
```
