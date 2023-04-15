package oras

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type MounterInterface interface {
	// Mount an artifact as a volume
	Mount(sourcePath string, destPath, mountType string, opts ...string) error

	// Unmount an artifact as a volume
	UMount(destPath string) error

	// Verify mount
	IsMounted(destPath string, testRun bool) (bool, error)
}

type Mounter struct {
	MounterInterface
}

var _ MounterInterface = &Mounter{}

type findmntResponse struct {
	FileSystems []fileSystem `json:"filesystems"`
}

type fileSystem struct {
	Target      string `json:"target"`
	Propagation string `json:"propagation"`
	FsType      string `json:"fstype"`
	Options     string `json:"options"`
}

const (
	mountCmd   = "mount"
	umountCmd  = "umount"
	findmntCmd = "findmnt"
	newDirMode = 0750
)

func (m *Mounter) Mount(sourcePath, destPath, mountType string, opts ...string) error {
	mountArgs := []string{}
	if sourcePath == "" {
		return errors.New("Mounter::Mount -- sourcePath must be provided")
	}

	if destPath == "" {
		return errors.New("Mounter::Mount -- Destination path must be provided")
	}

	// $ sudo mount -o ro,bind myfile destdir/myfile
	// TODO opts could be variable
	mountArgs = append(mountArgs, "-o", "bind")
	mountArgs = append(mountArgs, sourcePath)
	mountArgs = append(mountArgs, destPath)

	// create target, os.Mkdirall is noop if it exists
	err := os.MkdirAll(destPath, newDirMode)
	if err != nil {
		return err
	}
	log.Info(mountCmd + " " + strings.Join(mountArgs, " "))
	out, err := exec.Command(mountCmd, mountArgs...).CombinedOutput()

	// $ touch destdir/myfile
	// $ sudo mount -o ro,bind myfile destdir/myfile

	if err != nil {
		return fmt.Errorf("Mounter::Mount -- mounting failed: %v cmd: '%s %s' output: %q",
			err, mountCmd, strings.Join(mountArgs, " "), string(out))
	}
	return nil
}

func (m *Mounter) UMount(destPath string) error {
	umountArgs := []string{}

	if destPath == "" {
		return errors.New("Mounter::UMount -- Destination path must be provided")
	}
	umountArgs = append(umountArgs, destPath)

	out, err := exec.Command(umountCmd, umountArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Mounter::UMount -- mounting failed: %v cmd: '%s %s' output: %q",
			err, umountCmd, strings.Join(umountArgs, " "), string(out))
	}

	return nil
}

func (m *Mounter) IsMounted(destPath string, testRun bool) (bool, error) {
	if destPath == "" {
		return false, errors.New("Mounter::IsMounted -- target must be provided")
	}

	_, err := exec.LookPath(findmntCmd)
	if err != nil {
		if err == exec.ErrNotFound {
			return false, fmt.Errorf("Mounter::IsMounted -- %q executable not found in $PATH", findmntCmd)
		}
		return false, err
	}

	findmntArgs := []string{"-o", "TARGET,PROPAGATION,FSTYPE,OPTIONS", "-M", destPath, "-J"}
	out, err := exec.Command(findmntCmd, findmntArgs...).CombinedOutput()
	if err != nil {
		// findmnt exits with non zero exit status if it couldn't find anything
		if strings.TrimSpace(string(out)) == "" {
			return false, nil
		}
		return false, fmt.Errorf("Mounter::IsMounted -- checking mounted failed: %v cmd: %q output: %q",
			err, findmntCmd, string(out))
	}

	if string(out) == "" {
		log.Warningf("Mounter::IsMounted -- %s returns no output while returning status 0 - unexpected behaviour but not an actual error", findmntCmd)
		return false, nil
	}

	var resp *findmntResponse
	err = json.Unmarshal(out, &resp)
	if err != nil {
		return false, fmt.Errorf("Mounter::IsMounted -- couldn't unmarshal data: %q: %s", string(out), err)
	}

	for _, fs := range resp.FileSystems {
		// check if the mount is propagated correctly. It should be set to shared, unless we run sanity tests
		if fs.Propagation != "shared" && !testRun {
			return true, fmt.Errorf("Mounter::IsMounted -- mount propagation for target %q is not enabled (%s instead of shared)", destPath, fs.Propagation)
		}
		// the mountpoint should match as well
		if fs.Target == destPath {
			return true, nil
		}
	}
	return false, nil
}
