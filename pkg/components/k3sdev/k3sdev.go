package k3sdev

import (
	"errors"
	"github.com/iyurev/pulumi-libs/pkg/components/k3sdev/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	GCPCloud = iota
	AWSCloud
	AzureCloud
)

var (
	ErrUnsupportedCloud = errors.New("unsupported cloud provider")
)

type K3sCluster struct {
	pulumi.ResourceState
	publicIP pulumi.StringOutput
}

func NewK3sCluster(ctx *pulumi.Context, name string, underlingCloud int, opts ...pulumi.ResourceOption) (*K3sCluster, error) {
	k3sCluster := &K3sCluster{}
	err := ctx.RegisterComponentResource("pkg:index:K3SCluster", name, k3sCluster, opts...)
	if err != nil {
		return nil, err
	}
	switch underlingCloud {
	case GCPCloud:
		out, err := gcp.NewK3SCluster(ctx, "k3s-dev", "", "gothic-concept-349709", "us-central1-a")
		if err != nil {
			return nil, err
		}
		k3sCluster.publicIP = out.PublicIP
	default:
		return nil, ErrUnsupportedCloud

	}

	return k3sCluster, nil
}
