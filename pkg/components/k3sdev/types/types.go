package types

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type K3sCluster struct {
	pulumi.ResourceState
	KubeConfig pulumi.StringOutput
	SSHConnStr pulumi.StringOutput
	ApiPubAddr pulumi.StringOutput
	PublicIP   pulumi.StringOutput
}
