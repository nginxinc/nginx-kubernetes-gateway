[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-kubernetes-gateway.svg?type=shield)](https://app.fossa.com/projects/custom%2B5618%2Fgithub.com%2Fnginxinc%2Fnginx-kubernetes-gateway?ref=badge_shield)

# NGINX Kubernetes Gateway

NGINX Kubernetes Gateway is an open-source project that provides an implementation of the [Gateway API](https://gateway-api.sigs.k8s.io/) using [NGINX](https://nginx.org/) as the data plane. The goal of this project is to implement the core Gateway APIs -- `Gateway`, `GatewayClass`, `HTTPRoute`, `TCPRoute`, `TLSRoute`, and `UDPRoute` -- to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. NGINX Kubernetes Gateway is currently under development and supports a subset of the Gateway API.

> Warning: This project is actively in development (pre-alpha feature state) and should not be deployed in a production environment.
> All APIs, SDKs, designs, and packages are subject to change.

# Run NGINX Kubernetes Gateway

## Prerequisites

Before you can build and run the NGINX Kubernetes Gateway, make sure you have the following software installed on your machine:
- [git](https://git-scm.com/)
- [GNU Make](https://www.gnu.org/software/software.html)
- [Docker](https://www.docker.com/) v18.09+
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Build the image

1. Clone the repo and change into the `nginx-kubernetes-gateway` directory:

   ```
   git clone https://github.com/nginxinc/nginx-kubernetes-gateway.git
   cd nginx-kubernetes-gateway
   ```

1. Build the image:
 
   ```
   make PREFIX=myregistry.example.com/nginx-kubernetes-gateway container
   ```

   Set the `PREFIX` variable to the name of the registry you'd like to push the image to. By default, the image will be named `nginx-kubernetes-gateway:0.0.1`.

1. Push the image to your container registry:

   ```
   docker push myregistry.example.com/nginx-kubernetes-gateway:0.0.1
   ```

   Make sure to substitute `myregistry.example.com/nginx-kubernetes-gateway` with your private registry.

## Deploy NGINX Kubernetes Gateway

You can deploy NGINX Kubernetes Gateway on an existing Kubernetes 1.16+ cluster. The following instructions walk through the steps for deploying on a [kind](https://kind.sigs.k8s.io/) cluster. 

1. Load the NGINX Kubernetes Gateway image onto your kind cluster:

   ```
   kind load docker-image nginx-kubernetes-gateway:0.0.1
   ```

   Make sure to substitute the image name with the name of the image you built.

1. Install the Gateway CRDs:

   ```
   kubectl apply -k "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.4.2" 
   ```

1. Create the nginx-gateway namespace:
   
    ```
    kubectl apply -f deploy/manifests/namespace.yaml
    ```

1. Create the njs-modules configmap:

    ```
    kubectl create configmap njs-modules --from-file=internal/nginx/modules/src/httpmatches.js -n nginx-gateway  
    ```

1. Create the GatewayClass resource:

    ```
    kubectl apply -f deploy/manifests/gatewayclass.yaml
    ```

1. Deploy the NGINX Kubernetes Gateway:

   Before deploying, make sure to update the Deployment spec in `nginx-gateway.yaml` to reference the image you built.

   ```
   kubectl apply -f deploy/manifests/nginx-gateway.yaml
   ``` 

1. Confirm the NGINX Kubernetes Gateway is running in `nginx-gateway` namespace:

   ```
   kubectl get pods -n nginx-gateway
   NAME                             READY   STATUS    RESTARTS   AGE
   nginx-gateway-5d4f4c7db7-xk2kq   2/2     Running   0          112s
   ```

## Expose NGINX Kubernetes Gateway

You can gain access to NGINX Kubernetes Gateway by creating a `NodePort` Service or a `LoadBalancer` Service.

### Create a NodePort Service

Create a service with type `NodePort`:

```
kubectl apply -f deploy/manifests/service/nodeport.yaml
```

A `NodePort` service will randomly allocate one port on every node of the cluster. To access NGINX Kubernetes Gateway, use an IP address of any node in the cluster along with the allocated port.

### Create a LoadBalancer Service

Create a service with type `LoadBalancer` using the appropriate manifest for your cloud provider. 

- For GCP or Azure:

   ```
   kubectl apply -f deploy/manifests/service/loadbalancer.yaml
   ```

   Lookup the public IP of the load balancer:
   
   ```
   kubectl get svc nginx-gateway -n nginx-gateway
   ``` 
   
   Use the public IP of the load balancer to access NGINX Kubernetes Gateway.
   
- For AWS:

   ```
   kubectl apply -f deploy/manifests/service/loadbalancer-aws-nlb.yaml
   ```

   In AWS, the NLB DNS name will be reported by Kubernetes in lieu of a public IP. To get the DNS name run:

   ```
   kubectl get svc nginx-gateway -n nginx-gateway
   ``` 
  
   In general, you should rely on the NLB DNS name, however for testing purposes you can resolve the DNS name to get the IP address of the load balancer:

   ```
   nslookup <dns-name>
   ```

# Test NGINX Kubernetes Gateway

To test the NGINX Kubernetes Gateway run:

```
make unit-test
```
