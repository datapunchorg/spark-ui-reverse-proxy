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

When people send browse url like "http://your-spark-ui-reverse-proxy-server:8080/sparkui/application-name", the proxy server
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
