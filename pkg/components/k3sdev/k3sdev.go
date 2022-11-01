package k3sdev

import (
	"errors"
	"github.com/iyurev/pulumi-libs/pkg/components/k3sdev/gcp"
	"github.com/iyurev/pulumi-libs/pkg/components/k3sdev/types"
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
	ctx.Export("ssh_connection_string", k3sCluster.SSHConnStr)
	ctx.Export("cluster_api_addr", k3sCluster.ApiPubAddr)
	ctx.Export("public_ip", k3sCluster.PublicIP)
	return k3sCluster, nil
}
