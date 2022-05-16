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

package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

var sparkUIAppNameURLRegex = regexp.MustCompile("{{\\s*[$]appName\\s*}}")
var sparkUIAppNamespaceURLRegex = regexp.MustCompile("{{\\s*[$]appNamespace\\s*}}")

// sparkUIBackendUrlFormat example: http://{{$appName}}-ui-svc.{{$appNamespace}}.svc.cluster.local:4040

func getSparkUIServiceUrl(sparkUIServiceUrlFormat string, appName string, appNamespace string) string {
	return sparkUIAppNamespaceURLRegex.ReplaceAllString(sparkUIAppNameURLRegex.ReplaceAllString(sparkUIServiceUrlFormat, appName), appNamespace)
}

func ServeSparkUI(c *gin.Context, config *ApiConfig, uiRootPath string) {
	path := c.Param("path")
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	id := ""
	index := strings.Index(path, "/")
	if index <= 0 {
		id = path
		path = ""
	} else {
		id = path[0:index]
		path = path[index + 1:]
	}
	backendUrl := getSparkUIServiceUrl(config.SparkUIServiceUrlFormat, id, config.SparkApplicationNamespace)
	proxyBasePath := fmt.Sprintf("%s/%s", uiRootPath, id)
	proxy, err := newReverseProxy(backendUrl, path, proxyBasePath)
	if err != nil {
		msg := fmt.Sprintf("Failed to create reverse proxy for %s: %s", id, err.Error())
		writeErrorResponse(c, http.StatusInternalServerError, msg, nil)
		return
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func newReverseProxy(backendUrl string, targetPath string, proxyBasePath string) (*httputil.ReverseProxy, error) {
	glog.Infof("Creating revers proxy for Spark UI backend url %s", backendUrl)
	if targetPath != "" {
		if !strings.HasPrefix(targetPath, "/") {
			targetPath = "/" + targetPath
		}
		backendUrl = backendUrl + targetPath
	}
	url, err := url.Parse(backendUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target Spark UI url %s: %s", backendUrl, err.Error())
	}
	director := func(req *http.Request) {
		glog.Infof("Reverse proxy: serving backend url %s for originally requested url %s", url, req.URL)
		req.URL = url
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	modifyResponse := func(resp *http.Response) error {
		if resp.StatusCode == http.StatusFound {
			headerName := "Location"
			locationHeaderValues := resp.Header[headerName]
			if len(locationHeaderValues) > 0 {
				newValues := make([]string, 0, len(locationHeaderValues))
				for _, oldHeaderValue := range locationHeaderValues {
					parsedUrl, err := url.Parse(oldHeaderValue)
					if err != nil {
						glog.Infof("Reverse proxy: invalid response header value %s: %s (backend url %s): %s", headerName, oldHeaderValue, url, err.Error())
						newValues = append(newValues, oldHeaderValue)
					} else {
						parsedUrl.Scheme = ""
						parsedUrl.Host = ""
						newPath := parsedUrl.Path
						if !strings.HasPrefix(newPath, "/") {
							newPath = "/" + newPath
						}
						parsedUrl.Path = proxyBasePath + newPath
						newHeaderValue := parsedUrl.String()
						glog.Infof("Reverse proxy: modifying response header %s from %s to %s (backend url %s)", headerName, oldHeaderValue, newHeaderValue, url)
						newValues = append(newValues, newHeaderValue)
					}
				}
				resp.Header[headerName] = newValues
			}
		}
		/* else {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return  err
			}
			err = resp.Body.Close()
			if err != nil {
				return err
			}
			b = bytes.Replace(b, []byte("setUIRoot('')"), []byte("setUIRoot('')"), -1)
			body := ioutil.NopCloser(bytes.NewReader(b))
			resp.Body = body
			resp.ContentLength = int64(len(b))
			resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
			return nil
		} */
		return nil
	}
	return &httputil.ReverseProxy{
		Director: director,
		ModifyResponse: modifyResponse,
	}, nil
}