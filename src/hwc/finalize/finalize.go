package finalize

import (
	"errors"
	"io"

	"github.com/cloudfoundry/libbuildpack"
)

type Stager interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/stager.go
	BuildDir() string
	DepDir() string
	DepsIdx() string
	DepsDir() string
}

type Manifest interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/manifest.go
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Installer interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/installer.go
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Command interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/command.go
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(dir string, program string, args ...string) (string, error)
}

type Hwc interface {
	CheckWebConfig(buildDir string) error
	InstallAppHwc() error
}

type Finalizer struct {
	BuildDir string
	Manifest Manifest
	Stager   Stager
	Command  Command
	Hwc      Hwc
	Log      *libbuildpack.Logger
}

var (
	errInvalidBuildDir  = errors.New("Invalid build directory provided")
	errMissingWebConfig = errors.New("Missing Web.config")
)

func (f *Finalizer) Run() error {
	f.Log.BeginStep("Configuring hwc")

	if err := f.Hwc.CheckWebConfig(f.BuildDir); err != nil {
		return err
	}

	return nil
}
