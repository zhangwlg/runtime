// Copyright (c) 2017 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package virtcontainers

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kata-containers/runtime/virtcontainers/persist/fs"
	"github.com/kata-containers/runtime/virtcontainers/store"
	"github.com/kata-containers/runtime/virtcontainers/utils"
	"github.com/sirupsen/logrus"
)

const testSandboxID = "7f49d00d-1995-4156-8c79-5f5ab24ce138"
const testContainerID = "containerID"
const testKernel = "kernel"
const testInitrd = "initrd"
const testImage = "image"
const testHypervisor = "hypervisor"
const testVirtiofsd = "virtiofsd"
const testHypervisorCtl = "hypervisorctl"
const testBundle = "bundle"

const testDisabledAsNonRoot = "Test disabled as requires root privileges"

// package variables set in TestMain
var testDir = ""
var sandboxDirConfig = ""
var sandboxFileConfig = ""
var sandboxDirState = ""
var sandboxDirLock = ""
var sandboxFileState = ""
var sandboxFileLock = ""
var testQemuKernelPath = ""
var testQemuInitrdPath = ""
var testQemuImagePath = ""
var testQemuPath = ""
var testClhKernelPath = ""
var testClhImagePath = ""
var testClhPath = ""
var testAcrnKernelPath = ""
var testAcrnImagePath = ""
var testAcrnPath = ""
var testAcrnCtlPath = ""
var testVirtiofsdPath = ""

var testHyperstartCtlSocket = ""
var testHyperstartTtySocket = ""

// cleanUp Removes any stale sandbox/container state that can affect
// the next test to run.
func cleanUp() {
	globalSandboxList.removeSandbox(testSandboxID)
	store.DeleteAll()
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, store.DirMode)

	setup()
}

func setup() {
	os.Mkdir(filepath.Join(testDir, testBundle), store.DirMode)

	for _, filename := range []string{testQemuKernelPath, testQemuInitrdPath, testQemuImagePath, testQemuPath} {
		_, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Could not recreate %s:%v", filename, err)
			os.Exit(1)
		}
	}
}

func setupAcrn() {
	os.Mkdir(filepath.Join(testDir, testBundle), store.DirMode)

	for _, filename := range []string{testAcrnKernelPath, testAcrnImagePath, testAcrnPath, testAcrnCtlPath} {
		_, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Could not recreate %s:%v", filename, err)
			os.Exit(1)
		}
	}
}

func setupClh() {
	os.Mkdir(filepath.Join(testDir, testBundle), store.DirMode)

	for _, filename := range []string{testClhKernelPath, testClhImagePath, testClhPath, testVirtiofsdPath} {
		_, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Could not recreate %s:%v", filename, err)
			os.Exit(1)
		}
	}
}

// TestMain is the common main function used by ALL the test functions
// for this package.
func TestMain(m *testing.M) {
	var err error

	flag.Parse()

	logger := logrus.NewEntry(logrus.New())
	logger.Logger.Level = logrus.ErrorLevel
	for _, arg := range flag.Args() {
		if arg == "debug-logs" {
			logger.Logger.Level = logrus.DebugLevel
		}
	}
	SetLogger(context.Background(), logger)

	testDir, err = ioutil.TempDir("", "vc-tmp-")
	if err != nil {
		panic(err)
	}

	fmt.Printf("INFO: Creating virtcontainers test directory %s\n", testDir)
	err = os.MkdirAll(testDir, store.DirMode)
	if err != nil {
		fmt.Println("Could not create test directories:", err)
		os.Exit(1)
	}

	utils.StartCmd = func(c *exec.Cmd) error {
		//startSandbox will check if the hypervisor is alive and
		// checks for the PID is running, lets fake it using our
		// own PID
		c.Process = &os.Process{Pid: os.Getpid()}
		return nil
	}

	testQemuKernelPath = filepath.Join(testDir, testKernel)
	testQemuInitrdPath = filepath.Join(testDir, testInitrd)
	testQemuImagePath = filepath.Join(testDir, testImage)
	testQemuPath = filepath.Join(testDir, testHypervisor)

	setup()

	testAcrnKernelPath = filepath.Join(testDir, testKernel)
	testAcrnImagePath = filepath.Join(testDir, testImage)
	testAcrnPath = filepath.Join(testDir, testHypervisor)
	testAcrnCtlPath = filepath.Join(testDir, testHypervisorCtl)

	setupAcrn()

	testVirtiofsdPath = filepath.Join(testDir, testBundle, testVirtiofsd)
	testClhKernelPath = filepath.Join(testDir, testBundle, testKernel)
	testClhImagePath = filepath.Join(testDir, testBundle, testImage)
	testClhPath = filepath.Join(testDir, testBundle, testHypervisor)

	setupClh()

	ConfigStoragePathSaved := store.ConfigStoragePath
	RunStoragePathSaved := store.RunStoragePath
	// allow the tests to run without affecting the host system.
	store.ConfigStoragePath = func() string { return filepath.Join(testDir, store.StoragePathSuffix, "config") }
	store.RunStoragePath = func() string { return filepath.Join(testDir, store.StoragePathSuffix, "run") }
	fs.TestSetRunStoragePath(filepath.Join(testDir, "vc", "sbs"))

	defer func() {
		store.ConfigStoragePath = ConfigStoragePathSaved
		store.RunStoragePath = RunStoragePathSaved
	}()

	// set now that configStoragePath has been overridden.
	sandboxDirConfig = filepath.Join(store.ConfigStoragePath(), testSandboxID)
	sandboxFileConfig = filepath.Join(store.ConfigStoragePath(), testSandboxID, store.ConfigurationFile)
	sandboxDirState = filepath.Join(store.RunStoragePath(), testSandboxID)
	sandboxDirLock = filepath.Join(store.RunStoragePath(), testSandboxID)
	sandboxFileState = filepath.Join(store.RunStoragePath(), testSandboxID, store.StateFile)
	sandboxFileLock = filepath.Join(store.RunStoragePath(), testSandboxID, store.LockFile)

	testHyperstartCtlSocket = filepath.Join(testDir, "test_hyper.sock")
	testHyperstartTtySocket = filepath.Join(testDir, "test_tty.sock")

	ret := m.Run()

	os.RemoveAll(testDir)

	os.Exit(ret)
}
