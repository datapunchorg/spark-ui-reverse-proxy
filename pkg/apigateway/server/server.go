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

package server

import (
	"fmt"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apigateway/handlers"
	"github.com/gin-gonic/gin"
)

func getV1ApiRootPath(urlPrefix string) string {
	return fmt.Sprintf("%s/v1", urlPrefix)
}

func Run(config Config) {
	port := config.Port

	urlPrefix := config.UrlPrefix
	if urlPrefix == "" {
		urlPrefix = DefaultUrlPrefix
	}

	router := gin.Default()

	apiConfig := handlers.ApiConfig{
		SparkApplicationNamespace: config.SparkApplicationNamespace,
		SparkUIServiceUrlFormat:   config.SparkUIServiceUrlFormat,
		SparkUIBaseProxyPrefix:    fmt.Sprintf("%s/sparkui", getV1ApiRootPath(config.UrlPrefix)),
	}

	router.GET("/health", handlers.HealthCheck)

	apiRootPath := getV1ApiRootPath(config.UrlPrefix)

	group := router.Group(apiRootPath, func(context *gin.Context) {})

	group.GET("/sparkui/*path",
		func(context *gin.Context) {
			handlers.ServeSparkUI(context, &apiConfig, apiRootPath + "/sparkui")
		})

	group.GET("/health", handlers.HealthCheck)

	router.Run(fmt.Sprintf(":%d", port))
}
