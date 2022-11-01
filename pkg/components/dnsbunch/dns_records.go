package dnsbunch

import (
	"errors"
	"github.com/iyurev/pulumi-libs/pkg/components/dnsbunch/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	GCPCloud = iota
	AWSCloud
	AzureCloud
)

type CloudProvider int

type DNSRecordsBunch struct {
	pulumi.ResourceState
	Records  map[string]pulumi.StringArray
	ZoneName string
}

func NewDNSRecordsBunch(ctx *pulumi.Context, zoneName, zoneDNSName string, cloudProvider CloudProvider, records map[string]pulumi.StringArray) (*DNSRecordsBunch, error) {
	rb := &DNSRecordsBunch{Records: records, ZoneName: zoneName}
	switch cloudProvider {
	case GCPCloud:
		err := gcp.CreateRecords(ctx, rb, zoneName, zoneDNSName, records)
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("unsupported cloud provider")
	}
	return rb, nil
}
