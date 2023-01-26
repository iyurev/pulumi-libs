package types

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type DevBox struct {
	Name string
	pulumi.ResourceState
	SSHConnStr pulumi.StringOutput
	PublicIP   pulumi.StringOutput
	Tags       pulumi.StringArrayInput
}
