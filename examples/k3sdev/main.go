package main

import (
	"github.com/iyurev/pulumi-libs/pkg/components/k3sdev"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		newCluster, err := k3sdev.NewK3sCluster(ctx, "sandbox-k3s-cluster", 0)
		if err != nil {
			return err
		}
		_ = newCluster
		return nil
	})
}
