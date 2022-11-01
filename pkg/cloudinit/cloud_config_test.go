package cloudinit

import (
	"runtime"
	"testing"
)

func TestNewUser(t *testing.T) {
	cc := NewCloudConfig()
	cc.AddPackages("jq", "vim")
	keys := []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDouY1ZKRvUNGyLgqHXoBHJ7xFwM/UJnzSY5a4svL94zKJJlpCKeVFGOALVkjTj7f2xskhwHoVE29nVqk+OmPhj5yPHHejObj/iiJsSlnCcT8jHZ3oIH507E/ZIdMRBSuCUJpLz1cpzzCR0B5/Q2/yYVp+sQfegPedkrLmeiKDbZoAGMmdOYu5UTXxWmeEeG0Q6dQofGIIY/cgPnb7HDR+mhQb+vj7Xx7NVBL+9/Vum2vlaBDPRJEA0e7Qd0RnrdEg/kfAmmWxTE6tHLPpt/dmMxVJ2GCB2tlQ8v5GBviaKOJTDJsJ6YWDtwPw6PS0H543ndgNKeVv2Gigtq59zWhbUX4wYtkGFLYA9JWdUvafxhe7XUlbvZlRHUi8hk2aTFOBLzjJXzIb+qPOFyvs3eCt8NE8msxub5I1r2sz81/bt+/r06sf/X0PcpEDrbAUrDeNv0A3RzJtu2qKDKiSq8W4BdqYr3YoxjLxkG7b3LEu3ArOZDKe2gUwna7sPT/+gqivIEiL8gKie/6RkJ7mgYVr/OIZh4QkJUdSzz9F2369WAce+/LxfwPi5VDO5Pa5pbQQeHsov1lcYLmfLhPEpnvIUdMARZR4c+54riCIG9owhf7GZgM8GAJC0a5Nzb7zVy98TLlgEJD2cFZGo9TnfSy3yT78u401Fzh1jG6tBFf+PeQ== test@local"}
	user, err := NewUser("root", keys)
	if err != nil {
		t.Fatal(err)
	}
	cc.AddUser(*user)
	t.Logf("Current environment OS: %s\n", runtime.GOOS)
	t.Logf("Cloud config user-data: %s\n", cc.String())
}
