# Example

In this example we deploy NGINX Kubernetes Gateway, a simple web application, and then configure NGINX Gateway to route traffic to that application using `HTTPRoute` resources.

## Running the Example

## 1. Deploy NGINX Kubernetes Gateway

1. Follow the [installation instructions](https://github.com/nginxinc/nginx-kubernetes-gateway/blob/main/README.md#run-nginx-gateway) to deploy NGINX Gateway.

1. Save the public IP address of NGINX Kubernetes Gateway into a shell variable:
   
   ```
   GW_IP=XXX.YYY.ZZZ.III
   ```

1. Save the port of NGINX Kubernetes Gateway:
   
   ```
   GW_PORT=<port number>
   ```

## 2. Deploy the Cafe Application  

1. Create the coffee and the tea deployments and services:
   
   ```
   kubectl apply -f cafe.yaml
   ```

1. Check that the Pods are running in the `default` namespace:

   ```
   kubectl -n default get pods
   NAME                      READY   STATUS    RESTARTS   AGE
   coffee-6f4b79b975-2sb28   1/1     Running   0          12s
   tea-6fb46d899f-fm7zr      1/1     Running   0          12s
   ```

## 3. Configure Routing

1. Create the `Gateway`:

   ```
   kubectl apply -f gateway.yaml
   ```

1. Create the `HTTPRoute` resources:

   ```
   kubectl apply -f cafe-routes.yaml
   ```

## 4. Test the Application

To access the application, we will use `curl` to send requests to the `coffee` and `tea` services. 

To get coffee:

```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
Server address: 10.12.0.18:80
Server name: coffee-7586895968-r26zn
```

To get tea:

```
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea
Server address: 10.12.0.19:80
Server name: tea-7cd44fcb4d-xfw2x
```
