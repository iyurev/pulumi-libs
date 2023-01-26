package devbox

import (
	"errors"

	"github.com/iyurev/pulumi-libs/pkg/components/devbox/gcp"
	"github.com/iyurev/pulumi-libs/pkg/components/devbox/types"
	"github.com/iyurev/pulumi-libs/pkg/constants"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewK3sCluster(ctx *pulumi.Context, name string, cloudProvider int, opts ...pulumi.ResourceOption) (*types.DevBox, error) {
	devBox := &types.DevBox{Name: name}
	err := ctx.RegisterComponentResource("pkg:index:DevBox", name, devBox, opts...)
	if err != nil {
		return nil, err
	}
	switch cloudProvider {
	case constants.GCPCloud:
		if err := gcp.NewVM(ctx, devBox); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported cloud provider")
	}

	return devBox, nil

}
