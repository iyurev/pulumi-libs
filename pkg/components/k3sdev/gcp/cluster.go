package gcp

import (
	_ "embed"
	"errors"
	"fmt"
	cloudconf "github.com/iyurev/pulumi-libs/pkg/cloudinit"
	"github.com/iyurev/pulumi-libs/pkg/components/k3sdev/types"
	"github.com/iyurev/pulumi-libs/pkg/constants"
	"github.com/iyurev/pulumi-libs/pkg/constants/gcp"
	"github.com/iyurev/pulumi-libs/pkg/k3s"
	"github.com/pulumi/pulumi-cloudinit/sdk/go/cloudinit"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	v1 "github.com/pulumi/pulumi-google-native/sdk/go/google/compute/v1"
	dns "github.com/pulumi/pulumi-google-native/sdk/go/google/dns/v1"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"log"
)

var (
	cloudConfigMetaKey   = pulumi.StringPtr("user-data")
	ErrWrongInstanceSize = errors.New("wrong instance size")
)

const (
	MachineSeriesE2      = "e2"
	MachineTypeMedium    = "medium"
	MachineTypeStandard2 = "standard-2"
	MachineTypeStandard4 = "standard-4"
	MachineTypeStandard8 = "standard-8"

	DiskTypeSSD      = "pd-ssd"
	DiskTypeStandard = "pd-standard"
	DiskTypeBalanced = "pd-balanced"

	defaultOsUsername = "root"

	OsUbuntuMinimal2022LTS = "projects/ubuntu-os-cloud/global/images/ubuntu-minimal-2204-jammy-v20220902"
)

func machineTypeFromConfig(instSize string) (string, error) {
	switch instSize {
	case "Micro":
		return MachineTypeMedium, nil
	case "Small":
		return MachineTypeStandard2, nil
	case "Medium":
		return MachineTypeStandard4, nil
	case "Large":
		return MachineTypeStandard8, nil
	case "":
		return MachineTypeMedium, nil
	default:
		return "", ErrWrongInstanceSize
	}

}

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

func NewK3SCluster(ctx *pulumi.Context, k3sCluster *types.K3sCluster, name string) error {
	project := config.Require(ctx, "google-native:project")
	zone := config.Require(ctx, "google-native:zone")
	pubKey := config.Get(ctx, "ssh:public-key")
	vmConf := NewVMConf(name, pubKey, project, zone)
	tags := pulumi.StringArray{
		pulumi.String("dev-k3s-cluster"),
	}
	dontDeleteDisk := config.GetBool(ctx, "k3s:dont-delete-disk")
	instanceSize := config.Get(ctx, "k3s:instance-size")
	_ = instanceSize
	// Create a new ssh key par
	key, err := tls.NewPrivateKey(ctx, "private-key", &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
		RsaBits:   pulumi.IntPtr(4096),
	}, pulumi.Parent(k3sCluster))
	if err != nil {
		return err
	}
	var createDNSRecords bool
	zoneName := config.Get(ctx, "k3s:zone-name")
	var zoneDNSName string
	if !IsEmptyStr(zoneName) {
		dnsZone, err := dns.LookupManagedZone(ctx, &dns.LookupManagedZoneArgs{
			ManagedZone: zoneName,
			Project:     &vmConf.Project,
		})
		if err != nil {
			return err
		}
		zoneDNSName = dnsZone.DnsName
		createDNSRecords = true

	}

	rootUser, err := cloudconf.NewUser(vmConf.OsUsername, []string{vmConf.OsUserPubKey})
	if err != nil {
		return err
	}
	cmds := []string{
		"systemctl  disable  cloud-init",
		"snap install yq",
		"curl -sfL https://get.k3s.io | sh -",
	}
	packages := []string{"git-core", "vim", "jq", "bat"}
	k3sServerConf, err := k3s.NewConfig().SetWriteKubeConfigMode("0775").SetTLSSan(fmt.Sprintf("api.%s", CutOutDot(zoneDNSName))).ToYAML()
	if err != nil {
		return err
	}
	cloudConf := key.PublicKeyOpenssh.ApplyT(func(pkey string) string {
		cloudconf.AddSSHPubKeyToUser(rootUser, pkey)
		cc := cloudconf.NewCloudConfig()
		cc.AddUser(*rootUser)
		cc.AddCMDs(cmds...)
		cc.AddPackages(packages...)
		cc.AddFile(cloudconf.NewFile(k3s.DefaultConfigPath, k3sServerConf, cloudconf.FileEncodingB64, "0750"))
		return cc.String()
	}).(pulumi.StringOutput)

	cloudInit, err := cloudinit.NewConfig(ctx, "cloud-init", &cloudinit.ConfigArgs{
		Gzip:         pulumi.Bool(false),
		Base64Encode: pulumi.Bool(false),
		Parts: cloudinit.ConfigPartArray{
			cloudinit.ConfigPartArgs{
				ContentType: constants.CloudInitContentTypeCloudConfig,
				Content:     cloudConf,
			},
		},
	}, pulumi.Parent(k3sCluster))
	net, err := v1.LookupNetwork(ctx, &v1.LookupNetworkArgs{
		Network: gcp.DefaultNetwork,
	})
	if err != nil {
		return err
	}
	mType, err := machineTypeFromConfig(instanceSize)
	if err != nil {
		return err
	}
	vmConf.setMachineType(mType)

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
				AutoDelete: pulumi.BoolPtr(dontDeleteDisk == false),
				Boot:       pulumi.BoolPtr(true),
				Type:       v1.AttachedDiskTypePersistent,
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
					pulumi.String("22"),
				},
			},
		},
	}, pulumi.Parent(k3sCluster))
	vm, err := v1.NewInstance(ctx, vmConf.Name, vmArgs, pulumi.DependsOn([]pulumi.Resource{fw}), pulumi.Parent(k3sCluster))
	if err != nil {
		return err
	}
	publicIP := vm.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIP()
	publicApiAddr := pulumi.Sprintf("https://%s:6443", publicIP)

	sshConnectionHost := publicIP

	if createDNSRecords {
		records := make(map[string]pulumi.StringArray)
		sshConnectionHost = pulumi.Sprintf("%s.%s", vm.Name, CutOutDot(zoneDNSName))
		_, err := dns.NewResourceRecordSet(ctx, "kube-master", &dns.ResourceRecordSetArgs{
			Name:        pulumi.Sprintf("%s.%s", vm.Name, zoneDNSName),
			ManagedZone: pulumi.String(zoneName),
			Type:        pulumi.StringPtr("A"),
			Rrdatas:     pulumi.StringArray{publicIP},
		}, pulumi.Parent(k3sCluster))
		if err != nil {
			return err
		}

		publicApiAddr = pulumi.Sprintf("https://%s:6443", sshConnectionHost)

		var advancedDNSRecords = make([]string, 0)
		if err := config.GetObject(ctx, "k3s:advanced-dns-records", &advancedDNSRecords); err != nil {
			return err
		}
		if len(advancedDNSRecords) > 0 {
			for _, advRec := range advancedDNSRecords {
				records[advRec] = pulumi.StringArray{publicIP}
			}
		}
		for dnsName, addrs := range records {
			_, err := dns.NewResourceRecordSet(ctx, fmt.Sprintf("a-rec-%s", dnsName), &dns.ResourceRecordSetArgs{
				Name:        pulumi.Sprintf("%s.%s", dnsName, zoneDNSName),
				ManagedZone: pulumi.String(zoneName),
				Type:        pulumi.StringPtr("A"),
				Rrdatas:     addrs,
			}, pulumi.Parent(k3sCluster))
			if err != nil {
				log.Fatal(err)
			}

		}
		ctx.Export("dns_zone_name", pulumi.String(zoneDNSName))

	}
	fetchKubeConfCmd := pulumi.Sprintf("until [ -f /etc/rancher/k3s/k3s.yaml ]; do sleep 5; done; sudo cat /etc/rancher/k3s/k3s.yaml | yq  '.clusters[0].cluster.server = \"%s\"'; sleep 10;", publicApiAddr)
	cmd, err := remote.NewCommand(ctx, "fetch-kubeconfig", &remote.CommandArgs{
		Connection: remote.ConnectionArgs{
			Host:       publicIP,
			User:       pulumi.StringPtr(vmConf.OsUsername),
			PrivateKey: key.PrivateKeyOpenssh,
		},
		Create: fetchKubeConfCmd,
	}, pulumi.DependsOn([]pulumi.Resource{vm}), pulumi.Parent(k3sCluster))
	if err != nil {
		return err
	}
	k3sCluster.KubeConfig = cmd.Stdout
	k3sCluster.SSHConnStr = pulumi.Sprintf("ssh %s@%s", vmConf.OsUsername, sshConnectionHost)
	k3sCluster.PublicIP = publicIP
	k3sCluster.ApiPubAddr = publicApiAddr

	kbconfFilePath := pulumi.Sprintf("%s/%s-kubeconfig", "/tmp", vmConf.Name)
	dumpKubeConfig, err := local.NewCommand(ctx, "dump-kubeconfig", &local.CommandArgs{
		Create: pulumi.Sprintf("echo -n \"%s\" > %s", cmd.Stdout, kbconfFilePath),
		Delete: pulumi.Sprintf("rm -f %s", kbconfFilePath),
	}, pulumi.DependsOn([]pulumi.Resource{vm}), pulumi.Parent(k3sCluster))
	if err != nil {
		return err
	}

	_, err = local.NewCommand(ctx, "export-kube-env", &local.CommandArgs{
		Create: pulumi.Sprintf("touch ./kube-env.sh && chmod +x kube-env.sh  && echo export KUBECONFIG=%s  > ./kube-env.sh", kbconfFilePath),
		Delete: pulumi.Sprintf("rm -f ./kube-env.sh"),
	})
	_, err = local.NewCommand(ctx, "ssh-helper", &local.CommandArgs{
		Create: pulumi.Sprintf("touch ./ssh-to-%s.sh && chmod +x ssh-to-%s.sh && echo  ssh %s@%s > ./ssh-to-%s.sh", vm.Name, vm.Name, vmConf.OsUsername, sshConnectionHost, vm.Name),
		Delete: pulumi.Sprintf("rm -f  ./ssh-to-%s.sh", vm.Name),
	})
	if err != nil {
		return err
	}

	_ = dumpKubeConfig

	return nil
}

func NewVMConf(name, pubKey, project, zone string) *VMConf {
	vmConf := &VMConf{
		Name:         name,
		Image:        pulumi.StringPtr(OsUbuntuMinimal2022LTS),
		OsUsername:   defaultOsUsername,
		OsUserPubKey: pubKey,
		Series:       MachineSeriesE2,
		Project:      project,
		Zone:         zone,
	}
	vmConf.setMachineType(MachineTypeMedium)
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
