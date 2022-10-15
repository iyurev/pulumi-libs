package gcp

import (
	_ "embed"
	"fmt"
	"github.com/iyurev/pulumi-libs/pkg/constants"
	"github.com/pulumi/pulumi-cloudinit/sdk/go/cloudinit"
	"github.com/pulumi/pulumi-google-native/sdk/go/google"
	v1 "github.com/pulumi/pulumi-google-native/sdk/go/google/compute/v1"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	cloudConfigMetaKey = pulumi.StringPtr("user-data")

	//go:embed default_user_data.yaml.tmpl
	defaultCloudInitTmpl string
)

const (
	MachineSeriesE2      = "e2"
	MachineTypeMedium    = "medium"
	MachineTypeStandard2 = "standard-2"
	MachineTypeStandard4 = "standard-4"

	DiskTypeSSD      = "pd-ssd"
	DiskTypeStandard = "pd-standard"
	DiskTypeBalanced = "pd-balanced"

	defaultOsUsername = "root"

	OsImageFCOS36          = "projects/fedora-coreos-cloud/global/images/fedora-coreos-36-20220716-3-1-gcp-x86-64"
	OsUbuntuMinimal2022LTS = "projects/ubuntu-os-cloud/global/images/ubuntu-minimal-2204-jammy-v20220902"
)

type VMConf struct {
	Name         string
	Image        pulumi.StringPtrInput
	OsUsername   string
	OsUserPubKey string
	InstanceType pulumi.StringPtrInput
	DiskType     pulumi.StringPtrInput
	Zone         string
	Project      string
	Series       string
}

type DefaultCloudInitConf struct {
	SshPubKey string
}

func NewDefaultCloudInitConf(sshPubKey string) *DefaultCloudInitConf {
	return &DefaultCloudInitConf{SshPubKey: sshPubKey}
}

func (dci *DefaultCloudInitConf) Rendering() (string, error) {
	return RenderingTmpl(defaultCloudInitTmpl, dci)

}

type VMOutput struct {
	PublicIP pulumi.StringOutput
	SSHKey   pulumi.StringOutput
}

type VmArgsOpt func(ctx *pulumi.Context, args *v1.InstanceArgs) error

func WithCloudInit() VmArgsOpt {
	return nil
}

func NewK3SCluster(ctx *pulumi.Context, vmName, pubKey, project, zone string) (*VMOutput, error) {
	vmOut, err := NewVM(ctx, NewVMConf(vmName, pubKey, project, zone))
	if err != nil {
		return nil, err
	}
	return vmOut, nil
}

func NewVM(ctx *pulumi.Context, vmConf *VMConf) (*VMOutput, error) {
	var vmOut VMOutput
	provider, err := google.NewProvider(ctx, "gcp", &google.ProviderArgs{})
	if err != nil {
		return nil, err
	}
	pulumi.Printf("Current provider project: %s\n", provider.Project)

	tags := pulumi.StringArray{
		pulumi.String("dev-vm"),
	}
	var cloudConf pulumi.StringOutput

	defaultCloudInit := NewDefaultCloudInitConf(vmConf.OsUserPubKey)
	if IsEmptyStr(defaultCloudInit.SshPubKey) {
		//Create new ssh key par
		key, err := tls.NewPrivateKey(ctx, "private-key", &tls.PrivateKeyArgs{
			Algorithm: pulumi.String("RSA"),
			RsaBits:   pulumi.IntPtr(4096),
		})
		if err != nil {
			return nil, err
		}
		vmOut.SSHKey = key.PrivateKeyOpenssh
		cloudConfWithKey := key.PublicKeyOpenssh.ApplyT(func(pkey string) string {
			defaultCloudInit.SshPubKey = pkey
			conf, _ := defaultCloudInit.Rendering()
			return conf
		}).(pulumi.StringOutput)
		cloudConf = cloudConfWithKey

	} else {

		conf, err := defaultCloudInit.Rendering()
		if err != nil {
			return nil, err
		}
		cloudConf = pulumi.Sprintf("%s", conf)
	}

	cloudInit, err := cloudinit.NewConfig(ctx, "cloud-init", &cloudinit.ConfigArgs{
		Parts: cloudinit.ConfigPartArray{
			cloudinit.ConfigPartArgs{
				ContentType: constants.CloudInitContentTypeCloudConfig,
				Content:     cloudConf,
			},
		},
	})
	net, err := v1.LookupNetwork(ctx, &v1.LookupNetworkArgs{
		Network: "default",
	})
	if err != nil {
		return nil, err
	}
	vmArgs := &v1.InstanceArgs{
		Metadata: v1.MetadataArgs{
			Items: v1.MetadataItemsItemArray{
				v1.MetadataItemsItemArgs{
					Key:   cloudConfigMetaKey,
					Value: cloudInit.Rendered,
				},
			},
		},
		Tags: v1.TagsArgs{Items: tags},
		Disks: v1.AttachedDiskArray{
			v1.AttachedDiskArgs{
				Boot: pulumi.BoolPtr(true),
				Type: v1.AttachedDiskTypePersistent,
				InitializeParams: v1.AttachedDiskInitializeParamsArgs{
					SourceImage: vmConf.Image,
					DiskType:    vmConf.DiskType,
				},
			},
		},
		MachineType: vmConf.InstanceType,
		Zone:        pulumi.StringPtr(vmConf.Zone),
		NetworkInterfaces: v1.NetworkInterfaceArray{
			v1.NetworkInterfaceArgs{
				Network:       pulumi.StringPtr(net.SelfLink),
				AccessConfigs: v1.AccessConfigArray{v1.AccessConfigArgs{}},
			},
		},
	}

	fw, err := v1.NewFirewall(ctx, "firewall", &v1.FirewallArgs{Network: pulumi.StringPtr(net.SelfLink),
		TargetTags: tags,
		Allowed: v1.FirewallAllowedItemArray{
			v1.FirewallAllowedItemArgs{
				IpProtocol: pulumi.StringPtr("tcp"),
				Ports: pulumi.StringArray{
					pulumi.String("80"),
					pulumi.String("443"),
					pulumi.String("6443"),
				},
			},
		},
	}, pulumi.Provider(provider))
	vm, err := v1.NewInstance(ctx, vmConf.Name, vmArgs, pulumi.DependsOn([]pulumi.Resource{fw}), pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}
	_ = vm
	vmOut.PublicIP = vm.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIP()
	return &vmOut, nil
}

func NewVMConf(name, pubKey, project, zone string) *VMConf {
	vmConf := &VMConf{
		Name:         name,
		Image:        pulumi.StringPtr(OsImageFCOS36),
		OsUsername:   defaultOsUsername,
		OsUserPubKey: pubKey,
		Series:       MachineSeriesE2,
		Project:      project,
		Zone:         zone,
	}
	vmConf.setMachineType(MachineSeriesE2)
	vmConf.setDiskType(DiskTypeBalanced)
	return vmConf
}

func (vmc *VMConf) setMachineType(machineType string) {
	vmc.InstanceType = pulumi.StringPtr(fmt.Sprintf("zones/%s/machineTypes/%s-%s", vmc.Zone, vmc.Series, machineType))
}

func (vmc *VMConf) setDiskType(diskType string) pulumi.StringPtrInput {
	return pulumi.StringPtr(fmt.Sprintf("projects/%s/zones/%s/diskTypes/%s", vmc.Project, vmc.Zone, diskType))
}

func (vmc *VMConf) TypeMedium() {
	vmc.setMachineType(MachineTypeMedium)
}

func (vmc *VMConf) TypeStandard2() {
	vmc.setMachineType(MachineTypeStandard2)
}

func (vmc *VMConf) TypeStandard4() {
	vmc.setMachineType(MachineTypeStandard4)
}

func (vmc *VMConf) WithSSDDisk() {
	vmc.setDiskType(DiskTypeSSD)
}

func (vmc *VMConf) WithBalancedDisk() {
	vmc.setDiskType(DiskTypeBalanced)
}

func (vmc *VMConf) WithStandardDisk() {
	vmc.setDiskType(DiskTypeStandard)

}
