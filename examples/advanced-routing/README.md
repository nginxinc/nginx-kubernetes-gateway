# Advanced Routing

In this example we will deploy NGINX Kubernetes Gateway and configure advanced routing rules for a simple cafe application. 
We will use `HTTPRoute` resources to route traffic to the cafe application based on a combination of the request method, headers, and query parameters.

The cafe application consists of four services: `coffee-v1-svc`, `coffee-v2-svc`, `tea-svc`, and `tea-post-svc`. In the next section we will create the following routing rules for the cafe application:
- For the path `/coffee` route requests with the header `version` set to `v2` or with the query param `TEST` set to `v2` to `coffee-v2-svc`, and all other requests to `coffee-v1-svc`.
- For the path `/tea` route POST requests to `tea-post-svc`, and all other requests, such as `GET` requests, to `tea-svc`.  

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
   NAME                         READY   STATUS    RESTARTS   AGE
   coffee-v1-75869cf7ff-vlfpq   1/1     Running   0          17m
   coffee-v2-67499ff985-2k6cc   1/1     Running   0          17m
   tea-6fb46d899f-hjzwr         1/1     Running   0          17m
   tea-post-648dfcdd6c-2rlqb    1/1     Running   0          17m
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

We will use `curl` to send requests to the `/coffee` and `/tea` endpoints of the cafe application.

### 4.1 Access coffee

Send a request with the header `version:v2` and confirm that the response comes from `coffee-v2-svc`:

```bash
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee -H "version:v2"
Server address: 10.116.2.67:8080
Server name: coffee-v2-67499ff985-gw6vt
...
```

Send a request with the query parameter `TEST=v2` and confirm that the response comes from `coffee-v2-svc`:

```bash
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee?TEST=v2
Server address: 10.116.2.67:8080
Server name: coffee-v2-67499ff985-gw6vt
...
```

Send a request without the header or the query parameter and confirm the response comes from `coffee-v1-svc`:

```bash
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/coffee
Server address: 10.116.2.70:8080
Server name: coffee-v1-75869cf7ff-vlfpq
...
```

### 4.2 Access tea

Send a POST request and confirm that the response comes from `tea-post-svc`:

```bash
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea -X POST
Server address: 10.116.2.72:8080
Server name: tea-post-648dfcdd6c-2rlqb
...
```

Send a GET request and confirm that the response comes from `tea-svc`:

```bash
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea
Server address: 10.116.3.30:8080
Server name: tea-6fb46d899f-hjzwr
...
```

The `/tea` endpoint has routing rules configured for GET and POST requests. If you send a request with a different method, NGINX Kubernetes Gateway will return a 404.

Send a PUT request and confirm the 404 Not Found response:

```bash
curl --resolve cafe.example.com:$GW_PORT:$GW_IP http://cafe.example.com:$GW_PORT/tea -X PUT
<html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>nginx/1.21.3</center>
</body>
</html>
```
