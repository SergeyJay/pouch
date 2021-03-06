package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/environment"
	"github.com/alibaba/pouch/test/util"

	"github.com/go-check/check"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// PouchRunSuite is the test suite for run CLI.
type PouchRunSuite struct{}

func init() {
	check.Suite(&PouchRunSuite{})
}

// SetUpSuite does common setup in the beginning of each test suite.
func (suite *PouchRunSuite) SetUpSuite(c *check.C) {
	SkipIfFalse(c, environment.IsLinux)

	environment.PruneAllContainers(apiClient)

	PullImage(c, busyboxImage)
}

// TearDownTest does cleanup work in the end of each test.
func (suite *PouchRunSuite) TearDownTest(c *check.C) {
}

// TestRun is to verify the correctness of run container with specified name.
func (suite *PouchRunSuite) TestRun(c *check.C) {
	name := "test-run"

	res := command.PouchRun("run", "-d", "--name", name,
		busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	res = command.PouchRun("ps").Assert(c, icmd.Success)
	if out := res.Combined(); !strings.Contains(out, name) {
		c.Fatalf("unexpected output %s: should contains container %s\n",
			out, name)
	}
}

// TestRunPrintHi is to verify run container with executing a command.
func (suite *PouchRunSuite) TestRunPrintHi(c *check.C) {
	name := "test-run-print-hi"

	res := command.PouchRun("run", "--name", name, busyboxImage,
		"echo", "hi")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, "hi") {
		c.Fatalf("unexpected output %s expected hi\n", out)
	}
}

// TestRunPrintHiByImageID is to verify run container
// with executing a command by image ID.
func (suite *PouchRunSuite) TestRunPrintHiByImageID(c *check.C) {
	name := "test-run-print-hi-by-image-id"

	res := command.PouchRun("images")
	res.Assert(c, icmd.Success)

	imageID := imagesListToKV(res.Combined())[busyboxImage][0]

	res = command.PouchRun("run", "--name", name, imageID, "echo", "hi")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	if out := res.Combined(); !strings.Contains(out, "hi") {
		c.Fatalf("unexpected output %s expected hi\n", out)
	}
}

// TestRunInWrongWay tries to run create in wrong way.
func (suite *PouchRunSuite) TestRunInWrongWay(c *check.C) {
	for _, tc := range []struct {
		name string
		args string
	}{
		{name: "unknown flag", args: "-a"},

		// TODO: should add the following cases if ready
		// {name: "missing image name", args: ""},
	} {
		res := command.PouchRun("run", tc.args)
		c.Assert(res.Stderr(), check.NotNil, check.Commentf(tc.name))
	}
}

// Comment this flaky test.
// TestRunRestartPolicyAlways is to verify restart policy always works.
//func (suite *PouchRunSuite) TestRunRestartPolicyAlways(c *check.C) {
//	name := "TestRunRestartPolicyAlways"
//
//	command.PouchRun("run", "--name", name, "-d", "--restart=always",
// busyboxImage, "sh", "-c", "sleep 10000").Assert(c, icmd.Success)
//	command.PouchRun("stop", name).Assert(c, icmd.Success)
//	time.Sleep(5000 * time.Millisecond)
//
//	res := command.PouchRun("ps")
//	res.Assert(c, icmd.Success)
//
//	if out := res.Combined(); !strings.Contains(out, name) {
//		c.Fatalf("expect container %s to be up: %s\n", name, out)
//	}
//	DelContainerForceMultyTime(c,name)
//}

// TestRunRestartPolicyNone is to verify restart policy none works.
func (suite *PouchRunSuite) TestRunRestartPolicyNone(c *check.C) {
	name := "TestRunRestartPolicyNone"

	res := command.PouchRun("run", "--name", name, "-d",
		"--restart=no", busyboxImage,
		"sh", "-c", "sleep 1")
	defer DelContainerForceMultyTime(c, name)
	res.Assert(c, icmd.Success)

	time.Sleep(2000 * time.Millisecond)

	res = command.PouchRun("ps")
	res.Assert(c, icmd.Success)

	if out := res.Combined(); strings.Contains(out, name) {
		c.Fatalf("expect container %s to be exited: %s\n", name, out)
	}
}

// TestRunWithIPCMode is to verify --specific IPC mode when running a container.
// TODO: test container ipc namespace mode.
func (suite *PouchRunSuite) TestRunWithIPCMode(c *check.C) {
	name := "test-run-with-ipc-mode"

	res := command.PouchRun("run", "-d", "--name", name,
		"--ipc", "host", busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)
}

// TestRunWithUTSMode is to verify --specific UTS mode when running a container.
func (suite *PouchRunSuite) TestRunWithUTSMode(c *check.C) {
	name := "test-run-with-uts-mode"

	res := command.PouchRun("run", "-d", "--name", name,
		"--uts", "host", busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)
}

// TestRunWithSysctls is to verify run container with sysctls.
func (suite *PouchRunSuite) TestRunWithSysctls(c *check.C) {
	sysctl := "net.ipv4.ip_forward=1"
	name := "run-sysctl"

	res := command.PouchRun("run", "-d", "--name", name,
		"--sysctl", sysctl, busyboxImage, "top")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	output := command.PouchRun("exec", name,
		"cat", "/proc/sys/net/ipv4/ip_forward").Stdout()
	if !strings.Contains(output, "1") {
		c.Fatalf("failed to run a container with sysctls: %s", output)
	}
}

// TestRunWithAppArmor is to verify run container with security option AppArmor.
func (suite *PouchRunSuite) TestRunWithAppArmor(c *check.C) {
	appArmor := "apparmor=unconfined"
	name := "run-apparmor"

	res := command.PouchRun("run", "-d", "--name", name,
		"--security-opt", appArmor, busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	// TODO: do the test more strictly with effective AppArmor profile.
}

// TestRunWithSeccomp is to verify run container with security option seccomp.
func (suite *PouchRunSuite) TestRunWithSeccomp(c *check.C) {
	seccomp := "seccomp=unconfined"
	name := "run-seccomp"

	res := command.PouchRun("run", "-d", "--name", name,
		"--security-opt", seccomp, busyboxImage)
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)

	// TODO: do the test more strictly with effective seccomp profile.
}

// TestRunWithCapability is to verify run container with capability.
func (suite *PouchRunSuite) TestRunWithCapability(c *check.C) {
	capability := "NET_ADMIN"
	name := "run-capability"

	res := command.PouchRun("run", "--name", name, "--cap-add", capability,
		busyboxImage, "brctl", "addbr", "foobar")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)
}

// TestRunWithoutCapability tests running container with --cap-drop
func (suite *PouchRunSuite) TestRunWithoutCapability(c *check.C) {
	capability := "chown"
	name := "run-capability"
	expt := icmd.Expected{
		Err: "Operation not permitted",
	}
	command.PouchRun("run", "--name", name, "--cap-drop", capability,
		busyboxImage, "chown", "755", "/tmp").Compare(expt)
	defer DelContainerForceMultyTime(c, name)
}

// TestRunWithPrivilege is to verify run container with privilege.
func (suite *PouchRunSuite) TestRunWithPrivilege(c *check.C) {
	name := "run-privilege"

	res := command.PouchRun("run", "--name", name, "--privileged",
		busyboxImage, "brctl", "addbr", "foobar")
	defer DelContainerForceMultyTime(c, name)

	res.Assert(c, icmd.Success)
}

// checkFileContains checks the content of fname contains expt
func checkFileContains(c *check.C, fname string, expt string) {
	cmdResult := icmd.RunCommand("cat", fname)
	cmdResult.Assert(c, icmd.Success)
	c.Assert(strings.Contains(string(cmdResult.Stdout()), expt),
		check.Equals, true)
}

//
func (suite *PouchRunSuite) TestRunAlikernelScheLatSwitch(c *check.C) {
	// TODO: as runc has not implemented it, add test later
	SkipIfFalse(c, environment.IsAliKernel)
}

//
func (suite *PouchRunSuite) TestRunAlikernelMemoryWmarkRatio(c *check.C) {
	// TODO: as runc has not implemented it, add test later
	SkipIfFalse(c, environment.IsAliKernel)
}

//
func (suite *PouchRunSuite) TestRunAlikernelMemoryExtra(c *check.C) {
	// TODO: as runc has not implemented it, add test later
	SkipIfFalse(c, environment.IsAliKernel)
}

//
func (suite *PouchRunSuite) TestRunAlikernelMemoryForceEmptyCtl(c *check.C) {
	// TODO: as runc has not implemented it, add test later
	SkipIfFalse(c, environment.IsAliKernel)
}

// TestRunWithAnnotation is to verify the valid running container
// with annotation, and verify SpecAnnotation filed has been in inspect output.
func (suite *PouchRunSuite) TestRunWithAnnotation(c *check.C) {
	cname := "TestRunWithAnnotation"
	res := command.PouchRun("run", "-d", "--annotation", "a=b",
		"--annotation", "foo=bar",
		"--name", cname, busyboxImage).Assert(c, icmd.Success)
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}

	// kv in map not in order.
	var annotationSlice []string
	for k, v := range result[0].Config.SpecAnnotation {
		annotationSlice = append(annotationSlice, fmt.Sprintf("%s=%s", k, v))
	}
	annotationStr := strings.Join(annotationSlice, " ")

	c.Assert(util.PartialEqual(annotationStr, "a=b"), check.IsNil)
	c.Assert(util.PartialEqual(annotationStr, "foo=bar"), check.IsNil)
}

// TestRunWithExitCode is to verify the valid running container with exit code != 0.
func (suite *PouchRunSuite) TestRunWithExitCode(c *check.C) {
	cname := "TestRunWithExitCode"
	ret := command.PouchRun("run", "--name", cname, busyboxImage,
		"sh", "-c", "exit 101")
	defer DelContainerForceMultyTime(c, cname)

	// test process exit code $? == 101
	ret.Assert(c, icmd.Expected{ExitCode: 101})

	// test container ExitCode == 101
	output := command.PouchRun("inspect", cname).Stdout()
	result := []types.ContainerJSON{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		c.Errorf("failed to decode inspect output: %v", err)
	}
	c.Assert(result[0].State.ExitCode, check.Equals, int64(101))
}

// TestRunWithRM is to verify the valid running container with rm flag
func (suite *PouchRunSuite) TestRunWithRM(c *check.C) {
	cname := "TestRunWithRM"
	res := command.PouchRun("run", "--rm", "--name", cname, busyboxImage,
		"echo", "hello")
	defer DelContainerForceMultyTime(c, cname)
	res.Assert(c, icmd.Success)

	output := command.PouchRun("inspect", cname).Stderr()
	c.Assert(util.PartialEqual(output, cname+": not found"), check.IsNil)
}
