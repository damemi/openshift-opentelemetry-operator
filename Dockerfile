FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13 AS builder
WORKDIR /go/src/github.com/damemi/opentelemetry-operator
COPY . .
RUN CGO_ENABLED=0 go build -mod vendor -o opentelemetry-operator ./cmd/opentelemetry-operator

FROM registry.svc.ci.openshift.org/openshift/origin-v4.0:base
COPY --from=builder /go/src/github.com/damemi/opentelemetry-operator/opentelemetry-operator /usr/bin/
CMD ["/usr/bin/opentelemetry-operator"]