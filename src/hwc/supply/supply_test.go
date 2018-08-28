package supply_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=supply.go --destination=mocks_test.go --package=supply_test

var _ = Describe("Supply", func() {
	It("example test", func() {
		Expect(false).To(Equal(false))
	})

	Describe("Run", func() {
		// it downloads and installs the hwc.exe dependency
		// it copies hwc.exe to .cloudfoundry/hwc.exe
	})
})
