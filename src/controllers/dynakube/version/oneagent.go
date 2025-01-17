package version

import (
	"context"

	dynatracev1beta1 "github.com/Dynatrace/dynatrace-operator/src/api/v1beta1"
	"github.com/Dynatrace/dynatrace-operator/src/dockerconfig"
	"github.com/Dynatrace/dynatrace-operator/src/dtclient"
	"github.com/Dynatrace/dynatrace-operator/src/version"
)

type oneAgentUpdater struct {
	dynakube   *dynatracev1beta1.DynaKube
	dtClient   dtclient.Client
	digestFunc ImageDigestFunc
}

func newOneAgentUpdater(
	dynakube *dynatracev1beta1.DynaKube,
	dtClient dtclient.Client,
	digestFunc ImageDigestFunc,
) *oneAgentUpdater {
	return &oneAgentUpdater{
		dynakube:   dynakube,
		dtClient:   dtClient,
		digestFunc: digestFunc,
	}
}

func (updater oneAgentUpdater) Name() string {
	return "oneagent"
}

func (updater oneAgentUpdater) IsEnabled() bool {
	return updater.dynakube.NeedsOneAgent()
}

func (updater *oneAgentUpdater) Target() *dynatracev1beta1.VersionStatus {
	return &updater.dynakube.Status.OneAgent.VersionStatus
}

func (updater oneAgentUpdater) CustomImage() string {
	return updater.dynakube.CustomOneAgentImage()
}

func (updater oneAgentUpdater) CustomVersion() string {
	return updater.dynakube.CustomOneAgentVersion()
}

func (updater oneAgentUpdater) IsAutoUpdateEnabled() bool {
	return updater.dynakube.ShouldAutoUpdateOneAgent()
}

func (updater oneAgentUpdater) IsPublicRegistryEnabled() bool {
	return updater.dynakube.FeaturePublicRegistry() && !updater.dynakube.ClassicFullStackMode()
}

func (updater oneAgentUpdater) LatestImageInfo() (*dtclient.LatestImageInfo, error) {
	return updater.dtClient.GetLatestOneAgentImage()
}

func (updater *oneAgentUpdater) UseDefaults(ctx context.Context, dockerCfg *dockerconfig.DockerConfig) error {
	var err error
	latestVersion := updater.CustomVersion()
	if latestVersion == "" {
		latestVersion, err = updater.dtClient.GetLatestAgentVersion(dtclient.OsUnix, dtclient.InstallerTypeDefault)
		if err != nil {
			return err
		}
	}

	previousVersion := updater.Target().Version
	if previousVersion != "" {
		if downgrade, err := version.IsDowngrade(previousVersion, latestVersion); err != nil {
			return err
		} else if downgrade {
			log.Info("downgrade detected, which is not allowed in this configuration", "updater", updater.Name(), "from", previousVersion, "to", latestVersion)
			return nil
		}
	}

	defaultImage := updater.dynakube.DefaultOneAgentImage()
	err = updateVersionStatus(ctx, updater.Target(), defaultImage, updater.digestFunc, dockerCfg)
	if err != nil {
		return err
	}

	updater.Target().Version = latestVersion

	return nil
}
