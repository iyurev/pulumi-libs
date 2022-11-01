package k3s

import "testing"

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	cfg.SetTLSSan("api.labs.local").SetWriteKubeConfigMode("0675")
	yml, err := cfg.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("K3S config: \n%s", yml)
}
