package main

import (
	"github.com/iyurev/pulumi-libs/pkg/components/k3sdev"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clusterName := "dev-k3s-cluster"
		newCluster, err := k3sdev.NewK3sCluster(ctx, clusterName, k3sdev.GCPCloud)
		if err != nil {
			return err
		}
		_ = newCluster

		// k8sProv, err := kubernetes.NewProvider(ctx, "k3s-provider", &kubernetes.ProviderArgs{
		// 	Kubeconfig: newCluster.KubeConfig,
		// })
		// customTLS, err := decrypt.File("secrets/custom-tls.enc.yaml", "yaml")
		// if err != nil {
		// 	return err
		// }
		// traefikCR, err := yaml.NewConfigGroup(ctx, "traefik-conf-cr", &yaml.ConfigGroupArgs{
		// 	YAML: []string{string(customTLS)},
		// }, pulumi.Provider(k8sProv), pulumi.RetainOnDelete(true))
		// _ = traefikCR
		return nil
	})

}
