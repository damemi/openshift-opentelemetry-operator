# Running an example setup

The OpenTelemetry Collector is a "middleman" between 1) components that export OpenTelemetry traces and 2)
backend programs which parse traces into a UI.

The benefit of its use is the ability to hot-swap backends without having to recompile any of the traced components
or reconfigure their exporting endpoints. They need to only refer to the service that is configured on the Collector,
and any changes can be applied to the Collector.

So to run an example you need:
1. A tracing backend
2. The OpenTelemetry Collector (this operator)
3. A component that can be configured to export traces

## Jaeger backend installation

The Jaeger UI is a good example backend, and there is already an Operator provided by Red Hat on OperatorHub. To install it:
1. Search "**Jaeger**" on the OperatorHub in your OpenShift Console
2. Install the "**Red Hat OpenShift Jaeger**" operator (double check the namespace you're installing to, recommend creating a new project "jaeger")
3. Once it's installed, click on it and then click "**Create Instance**" to create an instance of the Jaeger UI. Click through using the default values

If you follow the above steps, you should see the following output from `oc`:
```
$ oc get all -n jaeger
NAME                                              READY   STATUS    RESTARTS   AGE
pod/jaeger-all-in-one-inmemory-86dcf8c9b7-prnr2   2/2     Running   0          59m

NAME                                                    TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                                  AGE
service/jaeger-all-in-one-inmemory-agent                ClusterIP   None             <none>        5775/UDP,5778/TCP,6831/UDP,6832/UDP      59m
service/jaeger-all-in-one-inmemory-collector            ClusterIP   172.30.217.8     <none>        9411/TCP,14250/TCP,14267/TCP,14268/TCP   59m
service/jaeger-all-in-one-inmemory-collector-headless   ClusterIP   None             <none>        9411/TCP,14250/TCP,14267/TCP,14268/TCP   59m
service/jaeger-all-in-one-inmemory-query                ClusterIP   172.30.215.229   <none>        443/TCP                                  59m

NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/jaeger-all-in-one-inmemory   1/1     1            1           59m

NAME                                                    DESIRED   CURRENT   READY   AGE
replicaset.apps/jaeger-all-in-one-inmemory-86dcf8c9b7   1         1         1       59m

NAME                                                  HOST/PORT                                                                                      PATH   SERVICES                           PORT    TERMINATION   WILDCARD
route.route.openshift.io/jaeger-all-in-one-inmemory   jaeger-all-in-one-inmemory-jaeger.apps.ci-ln-wz0w3tk-d5d6b.origin-ci-int-aws.dev.rhcloud.com          jaeger-all-in-one-inmemory-query   <all>   reencrypt     None
```
The `route` is the url to the Jaeger UI which allows you to search for traces. So in this example, it is `https://jaeger-all-in-one-inmemory-jaeger.apps.ci-ln-wz0w3tk-d5d6b.origin-ci-int-aws.dev.rhcloud.com`.

## OpenTelemetry Collector Installation

To install the OpenTelemetry Collector, simply run `oc create -f manifests/.` from the root of this operator's repo. This will install
everything to the namespace `opentelemetry-operator`.

You will also need to provide a config to the operator to configure the Collector's receivers and exporters (see
the [OpenTelemetry docs](https://opentelemetry.io/docs/collector/configuration/) for more info on these).

An example CR would be like so:
```
apiVersion: operator.openshift.io/v1alpha1
kind: OpenTelemetry
metadata:
  name: collector
  namespace: opentelemetry-operator
spec:
  image: "docker.io/otel/opentelemetry-collector:0.3.0"
  service:
    - name: jaeger-grpc
      port: 14250
      targetPort: 14250
    - name: otlp
      port: 9090
      targetPort: 9090
    - name: jaeger-http
      port: 14268
      targetPort: 14268
  config: |
    receivers:
      jaeger:
        protocols:
          thrift_http:
            endpoint: ":14268"

    processors:
      queued_retry:

    exporters:
      logging:
        loglevel: info
      jaeger_thrift_http:
        url: "http://jaeger-all-in-one-inmemory-collector.jaeger:14268/api/traces"

    service:
      pipelines:
        traces:
          receivers: [jaeger]
          processors: [queued_retry]
          exporters: [logging,jaeger_thrift_http]
```

**NOTE:** The `url` param under the `jaeger_thrift_http` exporter refers to the collector service created by the
Jaeger operator in the above step, serving on its default http port.

If your config works correctly, you should see the below in the operator namespace:

```
$ oc get all -n opentelemetry-operator
NAME                                           READY   STATUS    RESTARTS   AGE
pod/opentelemetry-collector-86cbd579c8-h5m9s   1/1     Running   0          51m
pod/opentelemetry-operator-789d7879bc-jdshj    1/1     Running   0          62m

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/otel-collector   ClusterIP   172.30.31.143   <none>        14250/TCP   62m

NAME                                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/opentelemetry-collector   1/1     1            1           62m
deployment.apps/opentelemetry-operator    1/1     1            1           62m

NAME                                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/opentelemetry-collector-86cbd579c8   1         1         1       51m
replicaset.apps/opentelemetry-operator-789d7879bc    1         1         1       62m
```

## Sample component with OTLP tracing

All of this is useless unless something is actually exporting OTLP (OpenTelemetry Collector)-compatible traces.

For this example, I've re-built the [Kube Scheduler Operator](https://github.com/damemi/cluster-kube-scheduler-operator/tree/jaeger-demo)
with some simple spans (and a Jaeger exporter) in its startup, using a trace wrapper I've provided in [library-go](https://github.com/damemi/library-go/tree/tracing-helpers).

To run this yourself, scale down the CVO so you can edit the scheduler operator's image.

Then, you can either build your own image from my fork or use the example `quay.io/mdame/kso-jaeger-demo`.
Edit the scheduler operator deployment (`oc edit deployment.apps/openshift-kube-scheduler-operator -n openshift-kube-scheduler-operator`)
to include the following in the pod template spec:

```
...
spec:
  containers:
    env:
    - name: JAEGER_ENDPOINT
      value: http://otel-collector.opentelemetry-operator:14268/api/traces
    ...
    image: quay.io/mdame/kso-jaeger-demo
    ...
...
```

After editing this you should see a new kube-scheduler-operator pod deploy, and refreshing the Jaeger UI a new
"kube-scheduler-operator" Service available to search traces for:

![jaeger](jaeger.png)

Clicking on a trace shows a detailed breakdown of each span within it:

![trace](trace.png)

### Extra Credit

To show more traces within kube components, I've also built a custom scheduler image, and a copy of kube-scheduler-operator
updated to use that image. This will show traces for a service `kube-scheduler` showing scheduler startup, and the scheduling
process for new pods.

To try this out with the above setup, use the `quay.io/mdame/kso-custom-tracing` image in your kube-scheduler-operator deployment.
Note that this will also require specifying the custom scheduler image name (`quay.io/mdame/custom-kube-scheduler-tracing`) and that
the Jaeger url needs to refer to the service IP, not just its name, or be exposed via a route. (This is because the static kube-scheduler
pods do not use the same cluster DNS as the operator pods).

Operator Deployment:
```
spec:
  containers:
    env:
    - name: JAEGER_ENDPOINT
      value: http://172.30.12.75:14268/api/traces
    - name: SCHEDULER_IMAGE
      value: quay.io/mdame/custom-kube-scheduler-tracing
    ...
    image: quay.io/mdame/kso-custom-tracing
```

With this change, once your kube-schedulers redeploy you'll be able to see traces when you schedule new pods that look like this:
![scheduler](https://user-images.githubusercontent.com/1839101/80658308-cc233700-8a53-11ea-92ca-a1edd4bfc40e.png)