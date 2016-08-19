# scheduler

Toy scheduler for use in Kubernetes demos.

## Usage

Annotate each node using the annotator command:

```
kubectl proxy
```
```
Starting to serve on 127.0.0.1:8001
```

```
go run annotator/main.go
```
```
gke-k0-default-pool-728d327f-00lq 1.60
gke-k0-default-pool-728d327f-3vzg 0.20
gke-k0-default-pool-728d327f-nmz7 0.80
gke-k0-default-pool-728d327f-pxee 0.05
gke-k0-default-pool-728d327f-xm4i 0.05
gke-k0-default-pool-728d327f-zynj 0.20
```

### Create a deployment

```
kubectl create -f deployments/nginx.yaml
```
```
deployment "nginx" created
```

The nginx pod should be in a pending state:

```
kubectl get pods
```
```
NAME                     READY     STATUS    RESTARTS   AGE
nginx-1431970305-mwghf   0/1       Pending   0          27s
```

### Run the Scheduler

List the nodes and note the price of each node.

```
annotator -l
```
```
gke-k0-default-pool-728d327f-00lq 0.80
gke-k0-default-pool-728d327f-3vzg 0.40
gke-k0-default-pool-728d327f-nmz7 0.40
gke-k0-default-pool-728d327f-pxee 0.05
gke-k0-default-pool-728d327f-xm4i 1.60
gke-k0-default-pool-728d327f-zynj 0.40
```

Run the best price scheduler:

```
scheduler
```
```
2016/08/19 11:16:25 Starting custom scheduler...
2016/08/19 11:16:28 Successfully assigned nginx-1431970305-mwghf to gke-k0-default-pool-728d327f-pxee
2016/08/19 11:16:35 Shutdown signal received, exiting...
2016/08/19 11:16:35 Stopped reconciliation loop.
2016/08/19 11:16:35 Stopped scheduler.
```

> Notice the pending nginx pod is deployed to the node with the lowest cost.

## Run the Scheduler on Kubernetes

```
kubectl create -f deployments/scheduler.yaml
```
``` 
deployment "scheduler" created
```
