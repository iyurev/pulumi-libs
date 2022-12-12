package k3sdev

import (
	"os"
	"path/filepath"
)

func GetManifestsList(dirPath string) ([]string, error) {
	var manifestsList = make([]string, 0)
	dirEntities, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return manifestsList, nil
		}
		return manifestsList, err
	}

	for _, entity := range dirEntities {
		if !entity.IsDir() {
			filePath := filepath.Join(dirPath, entity.Name())
			if filepath.Ext(filePath) == ".yaml" || filepath.Ext(filePath) == ".yml" {
				manifestsList = append(manifestsList, filePath)
			}
		}
	}
	return manifestsList, nil

}
