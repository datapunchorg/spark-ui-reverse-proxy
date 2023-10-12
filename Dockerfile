#
# This file was copied and modifed from
# https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/blob/master/Dockerfile
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARG SPARK_IMAGE=ghcr.io/datapunchorg/spark:spark-3.2-1642867779

FROM golang:1.21.3-alpine3.17 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source code
COPY main.go main.go
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o /usr/bin/spark-ui-reverse-proxy main.go

FROM alpine:3.17
USER root
COPY --from=builder /usr/bin/spark-ui-reverse-proxy /usr/bin/

ENTRYPOINT ["/usr/bin/spark-ui-reverse-proxy"]
