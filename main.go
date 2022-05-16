/*
Copyright https://github.com/datapunchorg

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
	"log"
)

var (
	namespace         = flag.String("namespace", apiv1.NamespaceAll, "The Kubernetes namespace where Spark applications are running.")
	port              = flag.Int("port", 8080, "Server port for this reverse proxy.")
	sparkUIServiceUrl = flag.String("spark-ui-service-url", "http://{{$appName}}-ui-svc.{{$appNamespace}}.svc.cluster.local:4040", "Spark UI Service URL, this should point to the Spark driver service which provides Spark UI inside that driver.")
	modifyRedirectUrl = flag.Bool("modify-redirect-url", true, "Whether to modify redirect url to make sure the redirect url uses correct path.")
)

func main() {
	flag.Parse()

	log.Printf("Starting server on port %s, application namespace: %s", *port, *namespace)

	config := server.Config{
		Port: *port,
		SparkApplicationNamespace: *namespace,
		SparkUIServiceUrl: *sparkUIServiceUrl,
		ModifyRedirectUrl: *modifyRedirectUrl,
	}
	server.Run(config)

	glog.Info("Shutting down the server")
}
