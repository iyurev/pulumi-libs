package cloudconfig

import (
	cloudconf "github.com/iyurev/pulumi-libs/pkg/cloudinit"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	defaultUserName = "developer"
)

func NewCloudInitConfig(ctx *pulumi.Context) (pulumi.String, error) {
	sshPubKey := config.Require(ctx, "devbox:ssh-pub-key")
	username := config.Get(ctx, "devbox:username")
	if username == "" {
		username = defaultUserName
	}
	user, err := cloudconf.NewUser(username, []string{sshPubKey})
	if err != nil {
		return "", err
	}
	cfg := cloudconf.NewCloudConfig()
	cfg.AddUser(*user)
	return pulumi.String(cfg.String()), nil

}
