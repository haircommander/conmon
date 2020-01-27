package conmon_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/containers/conmon/runner/conmon"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	conmonPath  = "/usr/bin/conmon"
	ctrID       = "abcdefghijklm"
	validPath   = "/tmp"
	invalidPath = "/not/a/path"
)

var _ = Describe("conmon cli", func() {
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
	Describe("ctr logs", func() {
		var tmpDir string
		var tmpLogPath string
		BeforeEach(func() {
			d, err := ioutil.TempDir(os.TempDir(), "conmon-")
			Expect(err).To(BeNil())
			tmpDir = d
			tmpLogPath = filepath.Join(tmpDir, "log")
		})
		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(BeNil())
		})
		It("no log driver should fail", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
			)
			Expect(stderr).To(ContainSubstring("Log driver not provided. Use --log-path"))
		})
		It("log driver as path should pass", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("", tmpLogPath),
			)
			Expect(stderr).To(BeEmpty())

			_, err := os.Stat(tmpLogPath)
			Expect(err).To(BeNil())
		})
		It("log driver as journald should pass", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("journald", ""),
			)
			Expect(stderr).To(BeEmpty())
		})
		It("log driver as journald with short cid should fail", func() {
			// conmon requires a cid of len > 12
			shortCtrID := "abcdefghijkl"
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(shortCtrID),
				conmon.WithContainerUUID(shortCtrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("journald", ""),
			)
			Expect(stderr).To(ContainSubstring("Container ID must be longer than 12 characters"))
		})
		It("log driver as k8s-file with path should pass", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
			)
			Expect(stderr).To(BeEmpty())

			_, err := os.Stat(tmpLogPath)
			Expect(err).To(BeNil())
		})
		It("log driver as k8s-file with invalid path should fail", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", invalidPath),
			)
			Expect(stderr).To(ContainSubstring("Failed to open log file"))
		})
		It("log driver as invalid driver should fail", func() {
			invalidLogDriver := "invalid"
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("invalid", tmpLogPath),
			)
			Expect(stderr).To(ContainSubstring("No such log driver " + invalidLogDriver))
		})
		It("multiple log drivers should pass", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
				conmon.WithLogDriver("journald", ""),
			)
			Expect(stderr).To(BeEmpty())

			_, err := os.Stat(tmpLogPath)
			Expect(err).To(BeNil())
		})
		It("multiple log drivers with one invalid should fail", func() {
			invalidLogDriver := "invalid"
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
				conmon.WithLogDriver("invalid", tmpLogPath),
			)
			Expect(stderr).To(ContainSubstring("No such log driver " + invalidLogDriver))
		})
	})
	Describe("exec", func() {
		var tmpDir string
		var tmpLogPath string
		BeforeEach(func() {
			d, err := ioutil.TempDir(os.TempDir(), "conmon-")
			Expect(err).To(BeNil())
			tmpDir = d
			tmpLogPath = filepath.Join(tmpDir, "log")
		})
		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(BeNil())
		})
		It("restore and exec together should fail", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
				conmon.WithExec(),
				conmon.WithRestorePath(tmpLogPath),
			)
			Expect(stderr).To(ContainSubstring("Cannot use 'exec' and 'restore' at the same time"))
		})
		It("exec attach without exec should fail", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
				conmon.WithExecAttach(),
			)
			Expect(stderr).To(ContainSubstring("Attach can only be specified with exec"))
		})
		It("exec attach without api v1 should fail", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
				conmon.WithExec(),
				conmon.WithExecAttach(),
			)
			Expect(stderr).To(ContainSubstring("Attach can only be specified for a non-legacy exec session"))
		})
		It("attach without api v1 should fail", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
				conmon.WithExec(),
				conmon.WithExecAttach(),
			)
			Expect(stderr).To(ContainSubstring("Attach can only be specified for a non-legacy exec session"))
		})
		It("exec without CUUID and with api v1 should fail", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
				conmon.WithExec(),
				conmon.WithExecAttach(),
				conmon.WithAPIV1(),
			)
			Expect(stderr).To(ContainSubstring("Container UUID not provided. Use --cuuid"))
		})
		It("exec without process spec should fail", func() {
			_, stderr := getConmonOutputGivenOptions(
				conmon.WithPath(conmonPath),
				conmon.WithContainerID(ctrID),
				conmon.WithContainerUUID(ctrID),
				conmon.WithRuntimePath(validPath),
				conmon.WithLogDriver("k8s-file", tmpLogPath),
				conmon.WithExec(),
			)
			Expect(stderr).To(ContainSubstring("Exec process spec path not provided. Use --exec-process-spec"))
		})
	})
})

func getConmonOutputGivenOptions(options ...conmon.ConmonOption) (string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	options = append(options, conmon.WithStdout(&stdout), conmon.WithStderr(&stderr))

	ci, err := conmon.CreateAndExecConmon(options...)
	Expect(err).To(BeNil())

	ci.Wait()

	return stdout.String(), stderr.String()
}
