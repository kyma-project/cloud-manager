package lib

import (
	"context"
	"fmt"
	"io/fs"
	"slices"
	"sync"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CloudProfileLoader interface {
	Load(ctx context.Context) (CloudProfileRegistry, error)
}

type CloudProfileRegistry interface {
	Get(name string) CloudProfileInfo
}

type CloudProfileInfo interface {
	Name() string
	GetGardenLinuxVersion() string
	GetKubernetesVersion() string
}

// cachingCloudProfileLoader

func NewCachingCloudProfileLoader(inner CloudProfileLoader) CloudProfileLoader {
	return &cachingCloudProfileLoader{
		inner: inner,
	}
}

type cachingCloudProfileLoader struct {
	m        sync.Mutex
	inner    CloudProfileLoader
	loaded   bool
	registry CloudProfileRegistry
	err      error
}

func (l *cachingCloudProfileLoader) Load(ctx context.Context) (CloudProfileRegistry, error) {
	l.m.Lock()
	defer l.m.Unlock()
	if !l.loaded {
		l.registry, l.err = l.inner.Load(ctx)
		l.loaded = true
	}
	return l.registry, l.err
}

// fileCloudProfileLoader

func NewFileCloudProfileLoader(fileSystem fs.ReadFileFS, fn string, config *e2econfig.ConfigType) CloudProfileLoader {
	return &fileCloudProfileLoader{
		fileSystem: fileSystem,
		fn:         fn,
		config:     config,
	}
}

type fileCloudProfileLoader struct {
	fileSystem fs.ReadFileFS
	fn         string
	config     *e2econfig.ConfigType
}

func (l *fileCloudProfileLoader) Load(ctx context.Context) (CloudProfileRegistry, error) {
	b, err := l.fileSystem.ReadFile(l.fn)
	if err != nil {
		return nil, err
	}
	cpList := &gardenerapicore.CloudProfileList{}
	decoder := scheme.Codecs.UniversalDeserializer()
	_, _, err = decoder.Decode(b, nil, cpList)
	if err != nil {
		return nil, err
	}

	var result []CloudProfileInfo
	for _, cp := range cpList.Items {
		if len(l.config.CloudProfiles) > 0 {
			specified := false
			for _, name := range l.config.CloudProfiles {
				if cp.Name == name {
					specified = true
					break
				}
			}
			if !specified {
				continue
			}
		}

		result = append(result, &defaultCloudProfileInfo{cp: cp})
	}

	return &defaultCloudProfileRegistry{
		profiles: result,
	}, nil
}

// gardenCloudProfileLoader

func NewGardenCloudProfileLoader(gardenClient client.Client, config *e2econfig.ConfigType) CloudProfileLoader {
	return &gardenCloudProfileLoader{
		gardenClient: gardenClient,
		config:       config,
	}
}

type gardenCloudProfileLoader struct {
	gardenClient client.Client
	config       *e2econfig.ConfigType
}

var _ CloudProfileLoader = &gardenCloudProfileLoader{}

func (l *gardenCloudProfileLoader) Load(ctx context.Context) (CloudProfileRegistry, error) {
	var result []CloudProfileInfo
	cpList := &gardenerapicore.CloudProfileList{}
	if err := l.gardenClient.List(ctx, cpList); err != nil {
		return nil, fmt.Errorf("failed to list CloudProfiles: %w", err)
	}
	for _, cp := range cpList.Items {
		specified := false
		for _, name := range l.config.CloudProfiles {
			if cp.Name == name {
				specified = true
				break
			}
		}
		if !specified {
			continue
		}

		result = append(result, &defaultCloudProfileInfo{cp: cp})
	}

	return &defaultCloudProfileRegistry{
		profiles: result,
	}, nil
}

// CloudProfileRegistry impl

type defaultCloudProfileRegistry struct {
	profiles []CloudProfileInfo
}

func (r *defaultCloudProfileRegistry) Get(name string) CloudProfileInfo {
	for _, p := range r.profiles {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// CloudProfileInfo impl

var _ CloudProfileInfo = &defaultCloudProfileInfo{}

type defaultCloudProfileInfo struct {
	cp                        gardenerapicore.CloudProfile
	maxKubernetesLinuxVersion string
	maxGardenLinuxVersion     string
}

func (p *defaultCloudProfileInfo) Name() string {
	return p.cp.Name
}

func (p *defaultCloudProfileInfo) findMaxVersion(supportedVersions []*version.Version) string {
	// sort descending
	slices.SortFunc(supportedVersions, func(a, b *version.Version) int {
		if a.LessThan(b) {
			return 1
		}
		if a.GreaterThan(b) {
			return -1
		}
		return 0
	})

	var v *version.Version
	if len(supportedVersions) == 0 {
		v, _ = version.ParseSemantic("0.0.0")
	} else {
		v = supportedVersions[0]
	}

	return v.String()
}

func (p *defaultCloudProfileInfo) GetGardenLinuxVersion() string {
	if p.maxGardenLinuxVersion != "" {
		return p.maxGardenLinuxVersion
	}

	var supportedVersions []*version.Version
	for _, mi := range p.cp.Spec.MachineImages {
		if mi.Name != "gardenlinux" {
			continue
		}
		for _, x := range mi.Versions {
			if ptr.Deref(x.Classification, gardenerapicore.ClassificationExpired) != gardenerapicore.ClassificationSupported {
				continue
			}
			v, err := version.ParseSemantic(x.Version)
			if err != nil || v == nil {
				continue
			}
			supportedVersions = append(supportedVersions, v)
		}
	}

	p.maxGardenLinuxVersion = p.findMaxVersion(supportedVersions)

	return p.maxGardenLinuxVersion
}

func (p *defaultCloudProfileInfo) GetKubernetesVersion() string {
	if p.maxKubernetesLinuxVersion != "" {
		return p.maxKubernetesLinuxVersion
	}

	var supportedVersions []*version.Version
	for _, x := range p.cp.Spec.Kubernetes.Versions {
		if ptr.Deref(x.Classification, gardenerapicore.ClassificationExpired) != gardenerapicore.ClassificationSupported {
			continue
		}
		v, err := version.ParseSemantic(x.Version)
		if err != nil || v == nil {
			continue
		}
		supportedVersions = append(supportedVersions, v)
	}

	p.maxKubernetesLinuxVersion = p.findMaxVersion(supportedVersions)

	return p.maxKubernetesLinuxVersion
}
