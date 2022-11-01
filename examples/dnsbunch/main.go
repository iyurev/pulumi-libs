package main

import (
	"github.com/iyurev/pulumi-libs/pkg/components/dnsbunch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		zoneName := "lab-k7s-dev"
		zoneDNSName := "lab.k7s.dev"
		records := map[string]pulumi.StringArray{"test": pulumi.StringArray{pulumi.String("127.0.0.1")}}
		rb, err := dnsbunch.NewDNSRecordsBunch(ctx, zoneName, zoneDNSName, dnsbunch.GCPCloud, records)
		if err != nil {
			return err
		}
		_ = rb
		return nil
	})
}
