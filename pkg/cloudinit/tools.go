package cloudinit

import "encoding/base64"

func EncodeToBase64(plain string) string {
	return base64.StdEncoding.EncodeToString([]byte(plain))
}
