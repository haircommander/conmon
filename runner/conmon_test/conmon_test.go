package conmon_test

import (
	"fmt"

	"github.com/containers/conmon/runner/conmon"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("conmon", func() {
	Describe("version", func() {
		It("Should return conmon version", func() {
			out, _ := getConmonOutputGivenOptions(
				conmon.WithVersion(),
				conmon.WithPath(conmonPath),
			)
			Expect(out).To(ContainSubstring("conmon version"))
			Expect(out).To(ContainSubstring("commit"))
		})
	})
	Describe("no container ID", func() {
		It("should fail", func() {
			_, err := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
			)
			Expect(err).To(ContainSubstring("conmon: Container ID not provided. Use --cid"))
		})
	})
	Describe("no container UUID", func() {
		It("should fail", func() {
			_, err := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
			)
			Expect(err).To(ContainSubstring("Container UUID not provided. Use --cuuid"))
		})
	})
	Describe("runtime path", func() {
		It("no path should fail", func() {
			_, err := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
			)
			Expect(err).To(ContainSubstring("Runtime path not provided. Use --runtime"))
		})
		It("invalid path should fail", func() {
			_, err := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(invalidPath),
			)
			Expect(err).To(ContainSubstring(fmt.Sprintf("Runtime path %s is not valid", invalidPath)))
		})
	})
})
