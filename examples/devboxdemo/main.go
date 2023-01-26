package main

import (
	"github.com/iyurev/pulumi-libs/pkg/components/devbox"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		newBox, err := devbox.NewK3sCluster(ctx, "demo-box", 0)
		if err != nil {
			return err
		}
		_ = newBox
		return nil
	})
}
