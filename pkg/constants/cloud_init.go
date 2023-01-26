package constants

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

const (
	CloudInitContentTypeCloudConfig pulumi.String = "text/cloud-config"
	CloudInitContentTypeShellScript pulumi.String = "text/x-shellscript"
	CloudConfigMetaKey                            = "user-data"
)
