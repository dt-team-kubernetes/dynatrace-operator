FROM scratch

# Core 0.10.3 labels.
LABEL operators.operatorframework.io.bundle.channel.default.v1=alpha
LABEL operators.operatorframework.io.0.10.3.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.0.10.3.manifests.v1=manifests/
LABEL operators.operatorframework.io.0.10.3.metadata.v1=metadata/
LABEL operators.operatorframework.io.0.10.3.package.v1=dynatrace-operator
LABEL operators.operatorframework.io.0.10.3.channels.v1=alpha
LABEL operators.operatorframework.io.metrics.builder=operator-sdk-v1.16.0+git
LABEL operators.operatorframework.io.metrics.mediatype.v1=metrics+v1
LABEL operators.operatorframework.io.metrics.project_layout=go.kubebuilder.io/v3


# Copy files to locations specified by labels.
COPY 0.10.3/manifests /manifests/
COPY 0.10.3/metadata /metadata/
LABEL com.redhat.openshift.versions="v4.8-v4.11"
LABEL com.redhat.delivery.operator.bundle=true
LABEL com.redhat.delivery.backport=true