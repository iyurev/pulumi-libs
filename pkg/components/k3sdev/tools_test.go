package k3sdev

import (
	"testing"
)

func TestGetManifestsList(t *testing.T) {
	manifestsList, err := GetManifestsList(manifestsDirPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range manifestsList {
		t.Logf("k8s manifest path: %s\n", m)
	}
}
