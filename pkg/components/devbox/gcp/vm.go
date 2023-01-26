package gcp

import (
	"github.com/iyurev/pulumi-libs/pkg/components/devbox/cloudconfig"
	"github.com/iyurev/pulumi-libs/pkg/components/devbox/types"
	"github.com/iyurev/pulumi-libs/pkg/constants"
	gcpconst "github.com/iyurev/pulumi-libs/pkg/constants/gcp"
	tools "github.com/iyurev/pulumi-libs/pkg/tools/gcp"
	v1 "github.com/pulumi/pulumi-google-native/sdk/go/google/compute/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func NewVM(ctx *pulumi.Context, devBox *types.DevBox) error {
	project := config.Require(ctx, "google-native:project")
	zone := config.Require(ctx, "google-native:zone")
	net, err := v1.LookupNetwork(ctx, &v1.LookupNetworkArgs{
		Network: gcpconst.DefaultNetwork,
	})
	cloudConfig, err := cloudconfig.NewCloudInitConfig(ctx)
	if err != nil {
		return err
	}
	vmArgs := &v1.InstanceArgs{
		Metadata: v1.MetadataArgs{
			Items: v1.MetadataItemsItemArray{
				v1.MetadataItemsItemArgs{
					Key:   pulumi.StringPtr(constants.CloudConfigMetaKey),
					Value: cloudConfig,
				},
			},
		},
		Tags: v1.TagsArgs{Items: devBox.Tags},
		Disks: v1.AttachedDiskArray{
			v1.AttachedDiskArgs{
				AutoDelete: pulumi.BoolPtr(true),
				Boot:       pulumi.BoolPtr(true),
				Type:       v1.AttachedDiskTypePersistent,
				InitializeParams: v1.AttachedDiskInitializeParamsArgs{
					SourceImage: pulumi.StringPtr(gcpconst.OsUbuntuMinimal2022LTS),
					DiskType:    tools.NewDiskTypeBuilder(zone, project).WithBalancedDisk().Done(),
				},
			},
		},
		MachineType: tools.NewMachineTypeBuilder(zone).TypeMedium().Done(),
		Zone:        pulumi.StringPtr(zone),
		NetworkInterfaces: v1.NetworkInterfaceArray{
			v1.NetworkInterfaceArgs{
				Network:       pulumi.StringPtr(net.SelfLink),
				AccessConfigs: v1.AccessConfigArray{v1.AccessConfigArgs{}},
			},
		},
	}

	fw, err := v1.NewFirewall(ctx, "firewall", &v1.FirewallArgs{Network: pulumi.StringPtr(net.SelfLink),
		TargetTags: devBox.Tags,
		Allowed: v1.FirewallAllowedItemArray{
			v1.FirewallAllowedItemArgs{
				IpProtocol: pulumi.StringPtr("tcp"),
				Ports: pulumi.StringArray{
					pulumi.String("80"),
					pulumi.String("443"),
					pulumi.String("6443"),
					pulumi.String("22"),
				},
			},
		},
	}, pulumi.Parent(devBox))
	vm, err := v1.NewInstance(ctx, devBox.Name, vmArgs, pulumi.DependsOn([]pulumi.Resource{fw}), pulumi.Parent(devBox))
	if err != nil {
		return err
	}
	_ = vm

	return nil
}
