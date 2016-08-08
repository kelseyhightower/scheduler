# scheduler

Toy scheduler for use in Kubernetes demos.

## Usage

```
kubectl get nodes
```
```
NAME                                     STATUS    AGE
gke-testing-default-pool-5b24138e-iw9u   Ready     14d
gke-testing-default-pool-5b24138e-tudj   Ready     14d
gke-testing-default-pool-5b24138e-vobb   Ready     14d
```

```
kubectl annotate nodes --overwrite \
  gke-testing-default-pool-5b24138e-iw9u kubernetes.io/cost=0.05
```

```
kubectl annotate nodes --overwrite \
  gke-testing-default-pool-5b24138e-tudj kubernetes.io/cost=0.20
```

```
kubectl annotate nodes --overwrite \
  gke-testing-default-pool-5b24138e-vobb kubernetes.io/cost=1.60
```
