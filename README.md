# opentelemetry-operator

An operator to manage an instance of the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector), 
built on [OpenShift's library-go](http://github.com/openshift/library-go).

Its basic function is to manage:
* a deployment of the OpenTelemetry Collector
* a service exposing the Collector

### Deploying

`oc create -f manifests/.`

[See a detailed example here](docs/example.md)

### Configuring

Provide a custom resource such as:
```yaml
apiVersion: operator.openshift.io/v1alpha1
kind: OpenTelemetry
metadata:
  name: collector
  namespace: opentelemetry-operator
spec:
  image: "quay.io/opentelemetry/opentelemetry-collector:v0.2.0"
  service:
    - name: jaeger-grpc
      port: 14250
      targetPort: 14250
  config: |
    receivers:
      jaeger:
        protocols:
          grpc:
            endpoint: <JAEGER ENDPOINT>
    processors:
      queued_retry:
    
    exporters:
      logging:
      
      jaeger_grpc:
        endpoint: <JAEGER ENDPOINT>
    
    service:
      pipelines:
        traces:
          receivers: [jaeger]
          processors: [queued_retry]
          exporters: [logging, jaeger_grpc]
```

Inspired by the existing [OpenTelemetry Operator](https://github.com/open-telemetry/opentelemetry-operator), with OpenShift libraries