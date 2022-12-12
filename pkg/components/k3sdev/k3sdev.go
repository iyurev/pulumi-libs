package k3sdev

import (
	"errors"
	"github.com/iyurev/pulumi-libs/pkg/components/k3sdev/gcp"
	"github.com/iyurev/pulumi-libs/pkg/components/k3sdev/types"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"go.mozilla.org/sops/v3/decrypt"
	"path"
)

const (
	GCPCloud = iota
	AWSCloud
	AzureCloud
)

var (
	ErrUnsupportedCloud       = errors.New("unsupported cloud provider")
	manifestsDirPath          = "./manifests"
	encryptedManifestsDirPath = "./secrets"
)

func NewK3sCluster(ctx *pulumi.Context, name string, underlingCloud int, opts ...pulumi.ResourceOption) (*types.K3sCluster, error) {
	k3sCluster := &types.K3sCluster{}
	err := ctx.RegisterComponentResource("pkg:index:K3SCluster", name, k3sCluster, opts...)
	if err != nil {
		return nil, err
	}
	switch underlingCloud {
	case GCPCloud:
		err := gcp.NewK3SCluster(ctx, k3sCluster, "k3s-dev")
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrUnsupportedCloud
	}
	manifestsList, err := GetManifestsList(manifestsDirPath)
	if err != nil {
		return nil, err
	}
	encryptedManifestsList, err := GetManifestsList(encryptedManifestsDirPath)
	if err != nil {
		return nil, err
	}

	if len(manifestsList) > 0 || len(encryptedManifestsList) > 0 {
		k8sProv, err := kubernetes.NewProvider(ctx, "installManifests", &kubernetes.ProviderArgs{
			Kubeconfig: k3sCluster.KubeConfig,
		})
		if err != nil {
			return nil, err
		}
		for _, manifestPath := range manifestsList {

			_, err := yaml.NewConfigFile(ctx, path.Base(manifestPath), &yaml.ConfigFileArgs{
				File: manifestPath,
			}, pulumi.Provider(k8sProv), pulumi.RetainOnDelete(true))
			if err != nil {
				return nil, err
			}
		}

		for _, manifestPath := range encryptedManifestsList {
			decryptedManifest, err := decrypt.File(manifestPath, "yaml")
			if err != nil {
				return nil, err
			}
			_, err = yaml.NewConfigGroup(ctx, path.Base(manifestPath), &yaml.ConfigGroupArgs{
				YAML: []string{string(decryptedManifest)},
			}, pulumi.Provider(k8sProv), pulumi.RetainOnDelete(true))
			if err != nil {
				return nil, err
			}
		}
	}
	ctx.Export("ssh_connection_string", k3sCluster.SSHConnStr)
	ctx.Export("cluster_api_addr", k3sCluster.ApiPubAddr)
	ctx.Export("public_ip", k3sCluster.PublicIP)
	return k3sCluster, nil
}
