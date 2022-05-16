## Quick Summary
This article describes how to build an HTTP Reverse Proxy to help Spark on Kubernetes users to access
Spark UI from outside Kubernetes cluster. This Reverse Proxy needs to be deployed together with
[spark-on-k8s-operator](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator).

## Introduction

If you run Apache Spark on Kubernetes, you may spend quite some time figuring out how to access Spark UI,
since Spark UI is embedded inside Spark driver and not directly exposed outside the Kubernetes cluster.

In the old days with Apache Hadoop/YARN, YARN has a 
[Web Proxy](https://hadoop.apache.org/docs/current/hadoop-yarn/hadoop-yarn-site/WebApplicationProxy.html)
to route request to Spark driver and serve Spark UI web pages from YARN Resource Manager.

In the new era with Kubernetes, there is no YARN anymore. We need a similar web proxy to make it easy to
access Spark UI. [spark-on-k8s-operator](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator) is
a popular Kubernetes Operator to run Spark application on Kubernetes. It is very easy to write a web service
together with spark-on-k8s-operator, which acts like a reverse proxy to server Spark UI web pages.

Here we will present how to build such a simple Spark UI Reverse Proxy using Go program language (Golang). 
The fully functional code is in [spark-ui-reverse-proxy code repo](https://github.com/datapunchorg/spark-ui-reverse-proxy).
We will highlight some details in the following content.

## Web Framework

[Gin](https://github.com/gin-gonic/gin) is a widely used web framework written in Go. Following is a quick example to
start a server listening on `/sparkui/*` web endpoint:

```
	router := gin.Default()

	apiConfig := handlers.ApiConfig{
		SparkApplicationNamespace: "spark-01",
	}

	router.GET("/sparkui/*path",
		func(context *gin.Context) {
			handlers.ServeSparkUI(context, &apiConfig, "/sparkui")
		})

	router.Run(fmt.Sprintf(":%d", port))
```

When people browse url like "http://your-spark-ui-reverse-proxy-server:8080/sparkui/application-name", the proxy server
will parse the url, get the application name, connect to the Spark driver and fetch the Spark UI web pages.

## Code Example to Serve Spark UI

The major code to server Spark UI is like following:

```
func ServeSparkUI(c *gin.Context, config *ApiConfig) {
	path := c.Param("path")
	// remove / prefix if there is any
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	// get application name
	appName := ""
	index := strings.Index(path, "/")
	appName = path[0:index]
	path = path[index + 1:]
	// get url for the underlying Spark UI Kubernetes service, which is created by spark-on-k8s-operator
	sparkUIServiceUrl := getSparkUIServiceUrl(config.SparkUIServiceUrl, appName, config.SparkApplicationNamespace)
	proxy, err := newReverseProxy(sparkUIServiceUrl, path)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create reverse proxy for application %s: %s", appName, err.Error()))
		return
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}
```

The `getSparkUIServiceUrl` function will return the url for the underlying Spark UI Kubernetes service. For example,
if submit a Spark application `application-01` inside namespace `spark-01`,
[spark-on-k8s-operator](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator) will create a Spark UI service
`application-01-ui-svc` to expose port 4040. We could connect to
`http://application-01-ui-svc.spark-01.svc.cluster.local:4040` to fetch Spark UI web pages.

When people browse our Spark UI Reverse Proxy, the `newReverseProxy` function will do:

1. Intercept the http request
2. Forward it to `http://application-01-ui-svc.spark-01.svc.cluster.local:4040`
3. Fetch the web page
4. Return the content to the browser

The major code for `newReverseProxy` is like following:

```
func newReverseProxy(sparkUIServiceUrl string, targetPath string) (*httputil.ReverseProxy, error) {
	if targetPath != "" {
		if !strings.HasPrefix(targetPath, "/") {
			targetPath = "/" + targetPath
		}
		sparkUIServiceUrl = sparkUIServiceUrl + targetPath
	}
	url, err := url.Parse(sparkUIServiceUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target Spark UI url %s: %s", sparkUIServiceUrl, err.Error())
	}
	director := func(req *http.Request) {
		log.Printf("Reverse proxy: serving backend url %s for originally requested url %s", url, req.URL)
		req.URL = url
	}
	return &httputil.ReverseProxy{
		Director: director,
	}, nil
}
```

## How to Use Spark UI Reverse Proxy

1. Following instructions in [spark-on-k8s-operator](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator)
to deploy Spark Operator in your Kubernetes cluster. Remember to choose a namespace to for your Spark application to
run there, for example, use `spark-01` as the namespace name.

2. Deploy Spark UI Reverse Proxy

```
kubectl apply -f https://raw.githubusercontent.com/datapunchorg/spark-ui-reverse-proxy/main/manifest/deployment.yaml
```

3. Run a Spark application, for example:

```
kubectl apply -f https://raw.githubusercontent.com/datapunchorg/spark-ui-reverse-proxy/main/manifest/spark-example-01.yaml
```
Please remember to add Spark configuration like following:
```
  sparkConf:
    spark.ui.proxyBase: /sparkui/spark-example-01
    spark.ui.proxyRedirectUri: /
```

4. Expose Spark UI Reverse Proxy:

```
kubectl port-forward service/spark-ui-reverse-proxy 8080:8080
```

5. Now you could browse your application's Spark UI:

```
http://localhost:8080/sparkui/spark-example-01
```

6. When running in production, you could set up
[Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) and add an
[Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) to expose Spark UI Reverse Proxy
without needing `kubectl port-forward`.

## One Click to Automatically Deploy Spark Operator and Spark UI Reverse Proxy

If you feel it is too complicated to deploy all the components, you could use
[punch](https://github.com/datapunchorg/punch) project, which provides a fully automated tool to deploy a read-to-use
Spark service on top of EKS. It creates an EKS cluster, and deploy Spark Operator there. It also deploys an
API Gateway to accept Spark application submission and provide a Spark UI Reverse Proxy.

The command to run `punch` is like following, see its [User Guide](https://github.com/datapunchorg/punch/blob/main/UserGuide.md)
for more details:

```
punch install SparkOnEks --patch spec.spark.gateway.password=password1 --print-usage-example
```

## Summary

Hopefully this article helps you to simplify your effort to run Apache Spark on Kubernetes. It also demonstrates how
easy to write a web reverse proxy server using Golang. The [punch](https://github.com/datapunchorg/punch) project could
automatically deploy Spark and the proxy server for you as well.

Wish you a great journey with Spark and Kubernetes!
