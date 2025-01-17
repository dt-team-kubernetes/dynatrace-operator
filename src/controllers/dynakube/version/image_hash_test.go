package version

import (
	"context"
	"fmt"
	"testing"

	dynatracev1beta1 "github.com/Dynatrace/dynatrace-operator/src/api/v1beta1"
	"github.com/Dynatrace/dynatrace-operator/src/dockerconfig"
	"github.com/containers/image/v5/docker/reference"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
)

type fakeRegistry struct {
	imageHashes map[string]string
}

func newEmptyFakeRegistry() *fakeRegistry {
	return newFakeRegistry(make(map[string]string))
}

func newFakeRegistryForImages(images ...string) *fakeRegistry {
	registryMap := make(map[string]string, len(images))
	for i, imageInfo := range images {
		registryMap[imageInfo] = fmt.Sprintf("hash-%d", i)
	}
	return newFakeRegistry(registryMap)
}

func newFakeRegistry(src map[string]string) *fakeRegistry {
	reg := fakeRegistry{
		imageHashes: make(map[string]string),
	}
	for key, val := range src {
		reg.setHash(key, val)
	}
	return &reg
}

func (registry *fakeRegistry) setHash(imagePath, hash string) *fakeRegistry {
	registry.imageHashes[imagePath] = hash
	return registry
}

func (registry *fakeRegistry) ImageVersion(imagePath string) (digest.Digest, error) {
	if version, exists := registry.imageHashes[imagePath]; !exists {
		return "", fmt.Errorf(`cannot provide version for image: "%s"`, imagePath)
	} else {
		return digest.NewDigestFromBytes(digest.SHA256, []byte(imagePath+":"+version)), nil
	}
}

func (registry *fakeRegistry) ImageVersionExt(_ context.Context, imagePath string, _ *dockerconfig.DockerConfig) (digest.Digest, error) {
	return registry.ImageVersion(imagePath)
}

func assertPublicRegistryVersionStatusEquals(t *testing.T, registry *fakeRegistry, imageRef reference.NamedTagged, versionStatus dynatracev1beta1.VersionStatus) { //nolint:revive // argument-limit
	assertVersionStatusEquals(t, registry, imageRef, versionStatus)
	assert.Empty(t, versionStatus.Version)
}

func assertVersionStatusEquals(t *testing.T, registry *fakeRegistry, imageRef reference.NamedTagged, versionStatus dynatracev1beta1.VersionStatus) { //nolint:revive // argument-limit
	expectedDigest, err := registry.ImageVersion(imageRef.String())

	assert.NoError(t, err, "Image version is unexpectedly unknown for '%s'", imageRef.String())
	assert.Contains(t, versionStatus.ImageID, expectedDigest.String())
}
