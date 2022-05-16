/*
Copyright 2022 DataPunch Organization

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/server"
	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	namespace                      = flag.String("namespace", apiv1.NamespaceAll, "The Kubernetes namespace to manage. Will manage custom resource objects of the managed CRD types for the whole cluster if unset.")
	apiGatewayPort                 = flag.Int("api-gateway-port", 0, "API gateway REST server port.")
	apiGatewayUrlPrefix            = flag.String("api-gateway-url-prefix", server.DefaultUrlPrefix, "API gateway REST server url prefix.")
	apiGatewayUserName             = flag.String("api-gateway-user-name", "", "User name for API gateway REST server. If this is specified, people must provide the matching user name and password when sending request to the API gateway REST server.")
	apiGatewayUserPassword         = flag.String("api-gateway-user-password", "", "User password for API gateway REST server. If this is specified together with user name, people must provide the matching user name and password when sending request to the API gateway REST server.")
	apiGatewaySparkUIServiceURLFormat               = flag.String("api-gateway-spark-ui-service-url-format", "http://{{$appName}}-ui-svc.{{$appNamespace}}.svc.cluster.local:4040", "Spark UI Service URL format, used to provide accessing to Spark UI.")
	apiGatewayConfig               = flag.String("api-gateway-config", "", "Config file for API gateway REST server to provide extra settings.")
)

func main() {
	flag.Parse()

	glog.Infof("Starting server on port %s, url prefix: %s, application namespace: %s", *apiGatewayPort, *apiGatewayUrlPrefix, *namespace)
	config := server.Config{
		Port: *apiGatewayPort,
		UrlPrefix: *apiGatewayUrlPrefix,
		SparkApplicationNamespace: *namespace,
		SparkUIServiceUrlFormat: *apiGatewaySparkUIServiceURLFormat,
	}
	server.Run(config)

	glog.Info("Shutting down the server")
}
