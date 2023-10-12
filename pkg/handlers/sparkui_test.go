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
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getSparkUIServiceUrl(t *testing.T) {
	assert.Equal(t, "", getSparkUIServiceUrl("", "app1", "ns1"))
	assert.Equal(t,
		"http://%s-ui-svc.%s.svc.cluster.local:4040",
		getSparkUIServiceUrl(
			"http://%s-ui-svc.%s.svc.cluster.local:4040", "app1", "ns1"))
	assert.Equal(t,
		"http://app1-ui-svc.ns1.svc.cluster.local:4040",
		getSparkUIServiceUrl(
			"http://{{$appName}}-ui-svc.{{$appNamespace}}.svc.cluster.local:4040", "app1", "ns1"))
}

func TestModifyRequest(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/sparkui/a3ac46c8487ecb95/static/webui.js?id=87c23377-4a64-47d3-82d7-5da9b39801a5", nil)
	assert.NoError(t, err, "unexpected error")
	u, err := url.Parse("http://a3ac46c8487ecb95-ui-svc.cluster.local:4040/static/webui.js")
	assert.NoError(t, err, "unexpected error")
	modifyRequest(r, u)
	t.Logf("url=%s", r.URL.String())
}

func TestModifyResponse(t *testing.T) {
	headers := http.Header{}
	headers.Add("Location", "/sparkui/StreamingQuery/statistics/?id=7ab24792-82e1-433b-a158-dc5792878f57")
	resp := &http.Response{
		Status: http.StatusText(http.StatusFound),
		StatusCode: http.StatusFound,
		Header: headers,
	}
	u, err := url.Parse("http://a3ac46c8487ecb95-ui-svc.cluster.local:4040/StreamingQuery/statistics/")
	assert.NoError(t, err, "unexpected error")

	err = modifyResponseRedirect(resp, "/sparkui/a3ac46c8487ecb95", u)
	assert.NoError(t, err, "unexpected error")
	t.Logf("\n\"/sparkui/a3ac46c8487ecb95\" -> url=%s", resp.Header["Location"][0])

	err = modifyResponseRedirect(resp, "", u)
	assert.NoError(t, err, "unexpected error")
	t.Logf("\n\"\" -> url=%s", resp.Header["Location"][0])

	err = modifyResponseRedirect(resp, "/", u)
	assert.NoError(t, err, "unexpected error")
	t.Logf("\n\"/\" -> url=%s", resp.Header["Location"][0])
}
