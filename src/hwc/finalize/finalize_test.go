package finalize_test

//go:generate mockgen -source=finalize.go --destination=mocks_test.go --package=finalize_test
import (
	"bytes"
	"errors"
	"hwc/finalize"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Finalize", func() {
	var (
		err          error
		buildDir     string
		finalizer    finalize.Finalizer
		logger       *libbuildpack.Logger
		buffer       *bytes.Buffer
		mockCtrl     *gomock.Controller
		mockManifest *MockManifest
		mockStager   *MockStager
		mockCommand  *MockCommand
		mockHwc      *MockHwc
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "hwc-buildpack.build.")
		buffer = new(bytes.Buffer)
		logger = libbuildpack.NewLogger(buffer)

		mockCtrl = gomock.NewController(GinkgoT())
		mockManifest = NewMockManifest(mockCtrl)
		mockStager = NewMockStager(mockCtrl)
		mockCommand = NewMockCommand(mockCtrl)
		mockHwc = NewMockHwc(mockCtrl)

		finalizer = finalize.Finalizer{
			BuildDir: buildDir,
			Manifest: mockManifest,
			Stager:   mockStager,
			Command:  mockCommand,
			Hwc:      mockHwc,
			Log:      logger,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())
	})

	Describe("Run", func() {
		Describe("happy path", func() {
			It("runs the hwc functions", func() {
				mockHwc.EXPECT().CheckWebConfig(buildDir).Return(nil)
				mockHwc.EXPECT().InstallAppHwc().Return(nil)

				err = finalizer.Run()
				Expect(err).To(BeNil())
			})
		})

		Describe("sad path", func() {
			It("runs the hwc functions", func() {
				mockHwc.EXPECT().CheckWebConfig(buildDir).Return(errors.New("BOOM"))
				mockHwc.EXPECT().InstallAppHwc().Return(nil)

				err = finalizer.Run()
				Expect(err).To(BeNil())
			})
		})
	})
})
