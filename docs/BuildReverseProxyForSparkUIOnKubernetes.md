## Introduction

If you run Apache Spark on Kubernetes, you may spend quite some time figuring out how to access Spark UI,
since Spark UI is embedded inside Spark driver and not directly exposed outside the Kubernetes cluster.

In the old days with Apache Hadoop/YARN, YARN has a 
[Web Proxy](https://hadoop.apache.org/docs/current/hadoop-yarn/hadoop-yarn-site/WebApplicationProxy.html)
to route request to Spark driver and serve Spark UI web pages from YARN Resource Manager.

In the new era with Kubernetes, there is no YARN anymore. We need a similar web proxy to make it easy to
access Spark UI.

Here we will present how to build a simple Web Proxy using Go program language (Golang) and use it to serve Spark UI. 
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

