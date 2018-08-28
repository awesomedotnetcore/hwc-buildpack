package finalize_test

import (
	"hwc/finalize"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hwc", func() {
	var (
		err      error
		buildDir string
		hwc      finalize.HwcImpl
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "hwc-buildpack.build.")

		hwc = finalize.HwcImpl{}
	})

	Describe("CheckWebConfig", func() {
		Context("build dir does not exist", func() {
			BeforeEach(func() {
				buildDir = "not/a/directory"
			})

			It("returns an error", func() {
				err = hwc.CheckWebConfig(buildDir)
				Expect(err.Error()).To(Equal("Invalid build directory provided"))
			})
		})

		Context("build dir does not contain web.config", func() {
			It("returns an error", func() {
				err = hwc.CheckWebConfig(buildDir)
				Expect(err.Error()).To(Equal("Missing Web.config"))
			})
		})

		Context("build dir contains web.config", func() {
			BeforeEach(func() {
				err = ioutil.WriteFile(filepath.Join(buildDir, "Web.ConfiG"), []byte("xx"), 0644)
				Expect(err).To(BeNil())
			})

			It("does not return an error", func() {
				err = hwc.CheckWebConfig(buildDir)
				Expect(err).To(BeNil())
			})
		})
	})
	//	Describe("InstallHWC", func() {
	//		It("installs HWC to <build_dir>/.cloudfoundry", func() {
	//			dep := libbuildpack.Dependency{Name: "hwc", Version: "99.99"}
	//
	//			mockManifest.EXPECT().DefaultVersion("hwc").Return(dep, nil)
	//			//mockInstaller.EXPECT().InstallDependency(dep, filepath.Join(buildDir, ".cloudfoundry"))
	//
	//			err = finalizer.InstallHWC()
	//			Expect(err).To(BeNil())
	//
	//			Expect(buffer.String()).To(ContainSubstring("-----> Installing HWC"))
	//			Expect(buffer.String()).To(ContainSubstring("HWC version 99.99"))
	//		})
	//	})
})
