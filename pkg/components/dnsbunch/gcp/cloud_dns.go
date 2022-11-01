package gcp

import (
	"fmt"
	dns "github.com/pulumi/pulumi-google-native/sdk/go/google/dns/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateRecords(ctx *pulumi.Context, parent pulumi.Resource, zoneName, zoneDNSName string, records map[string]pulumi.StringArray) error {
	for name, addrs := range records {
		newRec, err := dns.NewResourceRecordSet(ctx, fmt.Sprintf("ndns-add-a-record-%s", name), &dns.ResourceRecordSetArgs{
			Name:        pulumi.Sprintf("%s.%s.", name, zoneDNSName),
			ManagedZone: pulumi.String(zoneName),
			Type:        pulumi.StringPtr("A"),
			Rrdatas:     addrs,
		}, pulumi.Parent(parent))
		if err != nil {
			return err
		}
		_ = newRec
	}
	return nil
}
