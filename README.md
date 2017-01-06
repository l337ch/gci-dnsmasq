# gci-dnsmasq

For GKE managed Kubernetes clusters it is extremely difficult to manage DNS
domains for a) Private IP (RFC1918) spaces connected via Cloud VPN, b) split
horizon resolution for on premise hosted services (privates side of VPN), 
versus GKE resident services.  

## Current Solution

The solution provided here is a small Go application that runs dnsmas as a 
Deployment in the GKE Kubernetes cluster and is inserted as a shim in the
node hosts resolv.conf to intercept specific domains that require special
handling.

## Building
From source, create the Go static binary:
```
$ mkdir -p "${GOPATH}/src/github.com/samsung-cnct"
$ cd "${GOPATH}/src/github.com/samsung-cnct"
$ git clone https://github.com/samsung-cnct/gci-dnsmasq.git
$ cd gci-dnsmasq
$ CGO_ENABLED=0 GOOS=linux godep go build -a -ldflags '-w' -o gci_dnsmasq
```
## Building the Docker Image
Build and push the docker image, replacing Quay with your target registry.
```
$ docker build --rm --tag quay.io/samsung_cnct/gci-dnsmasq .
$ docker push quay.io/samsung_cnct/gci-dnsmasq:latest
```

## Helm Chart
This project will also end up being packaged as a Helm Chart eventually.

## Deployment
```
$ kubectl create -f gci-dnsmasq.yaml
deployment "gci-dnsmasq" created

$ kubectl get po,deployment -l app=gci-dnsmasq
NAME                              READY     STATUS    RESTARTS   AGE
po/gci-dnsmasq-2691091649-mgg5b   1/1       Running   0          8m

NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/gci-dnsmasq   1         1         1            1           8m

$ kubectl expose -f gci-dnsmasq.yaml 
service "gci-dnsmasq" exposed

$ kubectl get po,deployment,ep,svc -l app=gci-dnsmasq
NAME                              READY     STATUS    RESTARTS   AGE
po/gci-dnsmasq-1456860450-g2x6x   1/1       Running   0          2m

NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/gci-dnsmasq   1         1         1            1           2m

NAME             ENDPOINTS       AGE
ep/gci-dnsmasq   10.64.7.96:53   16s

NAME              CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
svc/gci-dnsmasq   10.67.245.64   <none>        53/TCP    16s

$ kubectl logs gci-dnsmasq-1456860450-g2x6x
gci-dnsmasq: DNSMASQ_CMD_ARGS: -k -d --no-resolv
gci-dnsmasq: metadata server as resolver ip: 169.254.169.254
gci-dnsmasq: kube-dns service cluster ip (vip): 10.67.240.10
gci-dnsmasq: kube-dns nameserver present in /etc/resolv.conf: true
gci-dnsmasq: starting dnsmasq: cmd: /usr/sbin/dnsmasq argv: [/usr/sbin/dnsmasq --keep-in-foreground --no-daemon --server=/**redacted**/10.40.240.70]
```
