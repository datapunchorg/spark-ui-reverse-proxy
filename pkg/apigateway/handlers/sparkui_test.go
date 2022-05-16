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
	"github.com/stretchr/testify/assert"
	"testing"
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
