package integration_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/cloudfoundry/libbuildpack/packager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var bpDir string
var buildpackVersion string
var packagedBuildpack cutlass.VersionedBuildpackPackage
var extensionBuildpackDir = "../../../fixtures/mysql-buildpack"
var extensionBuildpackName = "some_extension"

func init() {
	flag.StringVar(&buildpackVersion, "version", "", "version to use (builds if empty)")
	flag.BoolVar(&cutlass.Cached, "cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "384M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	if buildpackVersion == "" {
		err := PackageUniquelyVersionedExtensionBuildpack(extensionBuildpackName, extensionBuildpackDir, "", ApiHasStackAssociation())
		Expect(err).NotTo(HaveOccurred())

		packagedBuildpack, err := cutlass.PackageUniquelyVersionedBuildpack("", ApiHasStackAssociation())
		Expect(err).NotTo(HaveOccurred())

		data, err := json.Marshal(packagedBuildpack)
		Expect(err).NotTo(HaveOccurred())
		return data
	}

	return []byte{}
}, func(data []byte) {
	// Run on all nodes
	var err error
	if len(data) > 0 {
		err = json.Unmarshal(data, &packagedBuildpack)
		Expect(err).NotTo(HaveOccurred())
		buildpackVersion = packagedBuildpack.Version
	}

	bpDir, err = cutlass.FindRoot()
	Expect(err).NotTo(HaveOccurred())

	Expect(cutlass.CopyCfHome()).To(Succeed())
	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.RemovePackagedBuildpack(packagedBuildpack)).To(Succeed())
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func PushAppAndConfirm(app *cutlass.App) {
	err := app.Push()
	Expect(err).ToNot(HaveOccurred())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
	ExpectWithOffset(1, app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
}

func Restart(app *cutlass.App) {
	Expect(app.Restart()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}

func ApiHasTask() bool {
	apiVersionString, err := cutlass.ApiVersion()
	Expect(err).To(BeNil())
	apiVersion, err := semver.Make(apiVersionString)
	Expect(err).To(BeNil())
	apiHasTask, err := semver.ParseRange(">= 2.75.0")
	Expect(err).To(BeNil())
	return apiHasTask(apiVersion)
}

func ApiHasMultiBuildpack() bool {
	supported, err := cutlass.ApiGreaterThan("2.90.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func ApiHasStackAssociation() bool {
	supported, err := cutlass.ApiGreaterThan("2.113.0")
	Expect(err).NotTo(HaveOccurred())
	return supported
}

func SkipUnlessUncached() {
	if cutlass.Cached {
		Skip("Running cached tests")
	}
}

func SkipUnlessCached() {
	if !cutlass.Cached {
		Skip("Running uncached tests")
	}
}

func DestroyApp(app *cutlass.App) *cutlass.App {
	if app != nil {
		app.Destroy()
	}
	return nil
}

func AssertUsesProxyDuringStagingIfPresent(fixtureName string) {
	Context("with an uncached buildpack", func() {
		BeforeEach(SkipUnlessUncached)

		It("uses a proxy during staging if present", func() {
			proxy, err := cutlass.NewProxy()
			Expect(err).To(BeNil())
			defer proxy.Close()

			bpFile := filepath.Join(bpDir, buildpackVersion+"tmp")
			cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
			err = cmd.Run()
			Expect(err).To(BeNil())
			defer os.Remove(bpFile)

			traffic, built, _, err := cutlass.InternetTraffic(
				bpDir,
				filepath.Join("fixtures", fixtureName),
				bpFile,
				[]string{"HTTP_PROXY=" + proxy.URL, "HTTPS_PROXY=" + proxy.URL},
			)
			Expect(err).To(BeNil())
			Expect(built).To(BeTrue())

			destUrl, err := url.Parse(proxy.URL)
			Expect(err).To(BeNil())

			Expect(cutlass.UniqueDestination(
				traffic, fmt.Sprintf("%s.%s", destUrl.Hostname(), destUrl.Port()),
			)).To(BeNil())
		})
	})
}

func AssertNoInternetTraffic(fixtureName string) {
	It("has no traffic", func() {
		SkipUnlessCached()

		bpFile := filepath.Join(bpDir, buildpackVersion+"tmp")
		cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
		err := cmd.Run()
		Expect(err).To(BeNil())
		defer os.Remove(bpFile)

		traffic, built, _, err := cutlass.InternetTraffic(
			bpDir,
			filepath.Join("fixtures", fixtureName),
			bpFile,
			[]string{},
		)
		Expect(err).To(BeNil())
		Expect(built).To(BeTrue())
		Expect(traffic).To(BeEmpty())
	})
}

func PackageUniquelyVersionedExtensionBuildpack(name, fixtureDir, stack string, stackAssociationSupported bool) error {
	data, err := ioutil.ReadFile(filepath.Join(fixtureDir, "VERSION"))
	if err != nil {
		return fmt.Errorf("Failed to read VERSION file: %v", err)
	}
	buildpackVersion := strings.TrimSpace(string(data))
	buildpackVersion = fmt.Sprintf("%s.%s", buildpackVersion, time.Now().Format("20060102150405"))

	cached := cutlass.Cached
	file, err := packager.Package(fixtureDir, packager.CacheDir, buildpackVersion, stack, cached)
	if err != nil {
		return fmt.Errorf("Failed to package buildpack: %v", err)
	}

	if !stackAssociationSupported {
		stack = ""
	}

	err = cutlass.CreateOrUpdateBuildpack(name, file, stack)
	if err != nil {
		return fmt.Errorf("Failed to create or update buildpack: %v", err)
	}

	return nil
}
