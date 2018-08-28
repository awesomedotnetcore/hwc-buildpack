package finalize

import (
	"io/ioutil"
	"os"
	"strings"
)

type HwcImpl struct {
}

func (h *HwcImpl) CheckWebConfig(buildDir string) error {
	_, err := os.Stat(buildDir)
	if err != nil {
		return errInvalidBuildDir
	}

	files, err := ioutil.ReadDir(buildDir)
	if err != nil {
		return errInvalidBuildDir
	}

	var webConfigExists bool
	for _, file := range files {
		if strings.ToLower(file.Name()) == "web.config" {
			webConfigExists = true
			break
		}
	}

	if !webConfigExists {
		return errMissingWebConfig
	}
	return nil
}
