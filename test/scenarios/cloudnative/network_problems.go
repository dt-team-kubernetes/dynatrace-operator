//go:build e2e

package cloudnative

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/Dynatrace/dynatrace-operator/test/helpers/components/dynakube"
	"github.com/Dynatrace/dynatrace-operator/test/helpers/kubeobjects/namespace"
	"github.com/Dynatrace/dynatrace-operator/test/helpers/kubeobjects/pod"
	"github.com/Dynatrace/dynatrace-operator/test/helpers/sampleapps"
	"github.com/Dynatrace/dynatrace-operator/test/helpers/shell"
	"github.com/Dynatrace/dynatrace-operator/test/helpers/steps/assess"
	"github.com/Dynatrace/dynatrace-operator/test/helpers/steps/teardown"
	"github.com/Dynatrace/dynatrace-operator/test/helpers/tenant"
	"github.com/Dynatrace/dynatrace-operator/test/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const (
	agentMountPath    = "/opt/dynatrace/oneagent-paas"
	ldPreloadError    = "ERROR: ld.so: object '/opt/dynatrace/oneagent-paas/agent/lib64/liboneagentproc.so' from LD_PRELOAD cannot be preloaded"
	podRestartTimeout = 5 * time.Minute
)

var (
	csiNetworkPolicy = path.Join(project.TestDataDir(), "network/csi-denial.yaml")
)

func NetworkProblems(t *testing.T) features.Feature {
	builder := features.New("creating network problems")
	secretConfig := tenant.GetSingleTenantSecret(t)

	testDynakube := dynakube.NewBuilder().
		WithDefaultObjectMeta().
		ApiUrl(secretConfig.ApiUrl).
		CloudNative(defaultCloudNativeSpec()).
		WithAnnotations(map[string]string{
			"feature.dynatrace.com/max-csi-mount-attempts": "2",
		}).
		Build()

	namespaceBuilder := namespace.NewBuilder("network-problem-sample")
	sampleNamespace := namespaceBuilder.Build()
	sampleApp := sampleapps.NewSampleDeployment(t, testDynakube)
	sampleApp.WithNamespace(sampleNamespace)

	// Register operator install
	assess.InstallOperatorFromSource(builder, testDynakube)

	// Register network policy to block csi driver traffic
	assess.InstallManifest(builder, csiNetworkPolicy)

	// Register actual test
	assess.InstallDynakube(builder, &secretConfig, testDynakube)
	builder.Assess("install sample-apps", sampleApp.Install())
	builder.Assess("check for dummy volume", checkForDummyVolume(sampleApp))

	// Register network-policy, sample, dynakube and operator uninstall
	teardown.UninstallManifest(builder, csiNetworkPolicy)
	builder.Teardown(sampleApp.UninstallNamespace())
	teardown.UninstallDynatrace(builder, testDynakube)

	return builder.Feature()
}

func checkForDummyVolume(sampleApp sampleapps.SampleApp) features.Func {
	return func(ctx context.Context, t *testing.T, environmentConfig *envconf.Config) context.Context {
		resources := environmentConfig.Client().Resources()
		pods := sampleApp.GetPods(ctx, t, resources)

		for _, podItem := range pods.Items {
			require.NotNil(t, podItem)
			require.NotNil(t, podItem.Spec)
			require.NotEmpty(t, podItem.Spec.InitContainers)

			listCommand := shell.ListDirectory(agentMountPath)
			result, err := pod.Exec(ctx, resources, podItem, sampleApp.ContainerName(), listCommand...)

			require.NoError(t, err)
			assert.Contains(t, result.StdErr.String(), ldPreloadError)
		}
		return ctx
	}
}
