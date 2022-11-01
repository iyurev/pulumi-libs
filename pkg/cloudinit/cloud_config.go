package cloudinit

import (
	"fmt"
	"github.com/coreos/yaml"
	"github.com/elotl/cloud-init/config"
)

type CloudConfig struct {
	Users      []config.User `yaml:"users,omitempty"`
	Packages   []string      `yaml:"packages,omitempty"`
	RunCmd     []string      `yaml:"runcmd,omitempty"`
	WriteFiles []config.File `yaml:"write_files,omitempty"`
}

func (c *CloudConfig) AddCMDs(cmd ...string) {
	c.RunCmd = append(c.RunCmd, cmd...)
}

func (c *CloudConfig) AddPackages(packages ...string) {
	c.Packages = append(c.Packages, packages...)
}

func (c *CloudConfig) AddUser(user config.User) {
	c.Users = append(c.Users, user)
}

func (c *CloudConfig) String() string {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("#cloud-config\n%s", string(bytes))
}

func NewCloudConfig() *CloudConfig {
	return &CloudConfig{}
}

func NewUser(username string, sshKeys []string) (*config.User, error) {
	user := &config.User{
		Name:              username,
		SSHAuthorizedKeys: sshKeys,
	}
	return user, nil
}

func (c *CloudConfig) AddFile(files ...config.File) {
	c.WriteFiles = append(c.WriteFiles, files...)
}

const (
	FileEncodingB64         = "b64"
	FileEncodingGzip        = "gzip"
	FileEncodingB64PlusGzip = "gz+b64"
)

func NewFile(path, content, encoding, permission string) config.File {
	var fileContent = content
	if encoding == FileEncodingB64 {
		fileContent = EncodeToBase64(content)
	}
	return config.File{
		Content:            fileContent,
		Path:               path,
		Encoding:           encoding,
		RawFilePermissions: permission,
	}
}

func AddSSHPubKeyToUser(user *config.User, key string) {
	user.SSHAuthorizedKeys = append(user.SSHAuthorizedKeys, key)
}
