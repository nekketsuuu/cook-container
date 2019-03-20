// +build linux
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

func Run() {
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWUTS,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func checkErr(err error, mes string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", errors.Wrap(err, mes))
		os.Exit(1)
	}
}

func InitContainer() error {
	// UTS namespace
	checkErr(syscall.Sethostname([]byte("container")), "Failed to set hostname")
	// cgroups
	checkErr(os.MkdirAll("/sys/fs/cgroup/cpu/my-container", 0700), "Failed to create cgroups' namespace 'my-container'")
	checkErr(ioutil.WriteFile("/sys/fs/cgroup/cpu/my-container/tasks", []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644), "Failed to register cgroups' tasks to my-container namespace")
	checkErr(ioutil.WriteFile("/sys/fs/cgroup/cpu/my-container/cpu.cfs_quota_us", []byte("1000\n"), 0644), "Failed to add cgroups limit cpu.cfs_quota_us to 1000")
	// Prepare for overlayfs
	// * `workDir` is needed after kernel 3.18 (It's "overlay", not "overlayfs")
	// * `workDir` is needed to be empty
	basePath := "/root/overlayfs"
	lowerId := "lower"
	upperId := "upper"
	workId := "work"
	mergedId := "merged"
	lowerDir := filepath.Join(basePath, lowerId)
	upperDir := filepath.Join(basePath, upperId)
	workDir := filepath.Join(basePath, workId)
	mergedDir := filepath.Join(basePath, mergedId)
	checkErr(os.RemoveAll(upperDir), "Failed to remove the upperdir of overlayfs")
	checkErr(os.RemoveAll(workDir), "Failed to remove the workdir of overlayfs")
	checkErr(os.RemoveAll(mergedDir), "Failed to remove the mergeddir of overlayfs")
	checkErr(os.MkdirAll(lowerDir, 0700), "Failed to create lowerdir of overlayfs")
	checkErr(os.MkdirAll(upperDir, 0700), "Failed to create upperdir of overlayfs")
	checkErr(os.MkdirAll(workDir, 0700), "Failed to create workdir of overlayfs")
	checkErr(os.MkdirAll(mergedDir, 0700), "Failed to create mergeddir of overlayfs")
	// Mount proc for PID namespace
	checkErr(os.MkdirAll(filepath.Join(lowerDir, "proc"), 0700), "Failed to create lowerdir/proc")
	flags := uintptr(syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV)
	checkErr(syscall.Mount("proc", filepath.Join(lowerDir, "proc"), "proc", flags, ""), "Failed to mount proc")
	// It's needed for pivot_root that a filesystem of parent is different from that of child.
	// To achieve that, we can use bind mount.
	checkErr(os.Chdir(basePath), "Failed to chdir to "+basePath)
	checkErr(syscall.Mount(lowerId, lowerDir, "", syscall.MS_BIND|syscall.MS_REC, ""), "Failed to bind-mount the lowerdir of overlayfs as rootfs")
	checkErr(os.MkdirAll(filepath.Join(lowerDir, "oldrootfs"), 0700), "Failed to create oldrootfs in the lowerdir of overlayfs")
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	checkErr(syscall.Mount("overlay", mergedDir, "overlay", 0, opts), "Failed to mount overlayfs")
	// pivot_root
	checkErr(os.Chdir(basePath), "Failed to chdir to "+basePath)
	checkErr(syscall.PivotRoot(mergedId, filepath.Join(mergedDir, "oldrootfs")), "Failed to pivot_root")
	// Disable /oldrootfs, which points to the filesystem before pivotting the root.
	checkErr(syscall.Unmount("/oldrootfs", syscall.MNT_DETACH), "Failed to unmount oldrootfs")
	checkErr(os.RemoveAll("/oldrootfs"), "Failed to remove oldrootfs")
	checkErr(os.Chdir("/"), "Failed to chdir to /")
	checkErr(syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ()), "Failed to exec shell; Does /root/overlayfs/lower/bin/sh exist?")
	return nil
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s run\n", os.Args[0])
	os.Exit(2)
}

func main() {
	if len(os.Args) <= 1 {
		Usage()
	}
	switch os.Args[1] {
	case "run":
		Run()
	case "init":
		if err := InitContainer(); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		Usage()
	}
}
