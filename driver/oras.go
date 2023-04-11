package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	newVolumeMode  = 0755
	logsDirName    = "logs"
	volumesDirName = "volumes"
	mntDir         = "/mnt"
)

// todo(ad): in future possibly add more options (mount options?)
type orasHandler struct {
	container      string // oras artifact (container) to provide
	tag            string // oras artifact tag
	artifactRoot   string // path to artifact pull directory
	rootPath       string // oras root path
	pluginDataPath string // plugin data path (inside rootPath)
	name           string // handler name
	hostMountPath  string // host mount path
}

// NewOrasHandler creates a new oras handles to mount a container URI, pulled once
func NewOrasHandler(container string, rootPath, pluginDataPath, name string, num ...int) *orasHandler {
	var numSufix = ""
	if len(num) == 2 {
		if num[0] == 0 && num[1] == 1 {
			numSufix = ""
		} else {
			numSufix = fmt.Sprintf("_%02d", num[0])
		}
	} else if len(num) != 0 {
		log.Errorf("NewOrasHandler - Unexpected number of arguments: %d; expected 0 or 2", len(num))
	}

	// Get rid of oras prefix, if provided
	container = strings.TrimPrefix(container, "oras://")
	tag := "latest"
	if strings.Contains(container, ":") {
		parts := strings.Split(container, ":")
		container = parts[0]
		tag = parts[1]
	}

	log.Infof("Found ORAS container: %s:%s", container, tag)

	// Prepare a directory just for the artifact
	// TODO need to be able to name this predictably!
	artifact := strings.ReplaceAll(strings.ReplaceAll(container, "/", "-"), ".", "-")
	artifactRoot := path.Join(rootPath, pluginDataPath, artifact+"-"+tag)

	return &orasHandler{
		container:      container,
		artifactRoot:   artifactRoot,
		tag:            tag,
		rootPath:       rootPath,
		pluginDataPath: pluginDataPath,
		name:           name,
		hostMountPath:  path.Join(mntDir, fmt.Sprintf("%s%s", name, numSufix)),
	}
}

// SetOrasLogging sets up logging for the oras plugin
func (mnt *orasHandler) SetOrasLogging() {
	log.Infof("Setting up ORAS Logging. ORAS path: %s", path.Join(mnt.rootPath, mnt.pluginDataPath, logsDirName))
	orasLogFile := &lumberjack.Logger{
		Filename:   path.Join(mnt.HostPathToLogs(), fmt.Sprintf("%s.log", mnt.name)),
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     0,
		Compress:   true,
	}
	mw := io.MultiWriter(os.Stderr, orasLogFile)
	log.SetOutput(mw)
	log.Info("ORAS Logging set up!")
}

func (mnt *orasHandler) VolumeExist(volumeId string) (bool, error) {
	path := mnt.HostPathToVolume(volumeId)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (mnt *orasHandler) MountVolumeExist(volumeId string) (bool, error) {
	path := mnt.HostPathToMountVolume(volumeId)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (mnt *orasHandler) CreateMountVolume(volumeId string) error {
	path := mnt.HostPathToMountVolume(volumeId)
	if err := os.MkdirAll(path, newVolumeMode); err != nil {
		return err
	}
	return nil
}

func (mnt *orasHandler) CreateVolume(volumeId string, size int64) error {
	path := mnt.HostPathToVolume(volumeId)
	err := os.MkdirAll(path, newVolumeMode)
	return err
}

func (mnt *orasHandler) DeleteVolume(volumeId string) error {
	path := mnt.HostPathToVolume(volumeId)
	if err := os.RemoveAll(path); err != nil {
		// todo(ad): fix msg
		log.Errorf("-------------------ControllerService::DeleteVolume -- Couldn't remove volume %s in directory %s. Error: %s",
			volumeId, path, err.Error())
		return err
	}

	return nil
}

// Ensure the artifact (and pull contents there)
func (mnt *orasHandler) ensureArtifact() error {

	if _, err := os.Stat(mnt.artifactRoot); errors.Is(err, os.ErrNotExist) {

		// Pull Oras Container, if doesn't exist
		if err := mnt.OrasPull(); err != nil {
			return err
		}
	}
	return nil
}

// Pull the oras container to the plugin data directory
func (mnt *orasHandler) OrasPull() error {

	// 0. Create a file store
	log.Info("Creating oras filestore at: ", mnt.artifactRoot)
	fs, err := file.New(mnt.artifactRoot)
	if err != nil {
		return err
	}
	defer fs.Close()

	// 1. Connect to a remote repository
	log.Info("Preparing to pull from remote repository: ", mnt.container)
	ctx := context.Background()
	repo, err := remote.NewRepository(mnt.container)
	if err != nil {
		return err
	}
	// If authentication needed, could be added here
	// not recommended, make your images open source and public!

	// 2. Copy from the remote repository to the file store
	_, err = oras.Copy(ctx, repo, mnt.tag, fs, mnt.tag, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}
	return nil
}

// Mount mounts an oras container at specified point.
func (mnt *orasHandler) MountOras() error {

	mounter := Mounter{}
	mountSource := fmt.Sprintf("%s:%s", mnt.container, mnt.rootPath)
	mountOptions := make([]string, 0)

	// TODO here we can pull to plugin data directory with oras and then mount single file
	log.Infof("Oras - container: %s, target: %s, options: %v", mountSource, mnt.hostMountPath, mountOptions)

	// Pull (or ensure artifact already exists)
	err := mnt.ensureArtifact()
	if err != nil {
		return err
	}
	log.Infof("Oras artifact root: %s", mnt.artifactRoot)
	files, err := ioutil.ReadDir(mnt.artifactRoot)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		log.Info("Found artifact asset: ", f.Name())
	}

	if isMounted, err := mounter.IsMounted(mnt.hostMountPath); err != nil {
		return err
	} else if isMounted {
		log.Warnf("Mount found in %s. Unmounting...", mnt.hostMountPath)
		if err = mounter.UMount(mnt.hostMountPath); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(mnt.hostMountPath); err != nil {
		return err
	}
	if err := mounter.Mount(mnt.artifactRoot, mnt.hostMountPath, "", mountOptions...); err != nil {
		return err
	}
	log.Infof("Successfully mounted %s to %s", mnt.artifactRoot, mnt.hostMountPath)
	return nil
}

func (mnt *orasHandler) BindMount(src string, target string, options ...string) error {
	mounter := Mounter{}
	source := mnt.HostPathTo(src)
	log.Infof("BindMount - source: %s, target: %s, options: %v", source, target, options)
	if isMounted, err := mounter.IsMounted(target); err != nil {
		return err
	} else if !isMounted {
		if err := mounter.Mount(source, target, "", append(options, "bind")...); err != nil {
			return err
		}
	} else {
		log.Infof("BindMount - target %s is already mounted", target)
	}
	return nil
}

func (mnt *orasHandler) BindUMount(target string) error {
	mounter := Mounter{}
	log.Infof("BindUMount - target: %s", target)
	if mounted, err := mounter.IsMounted(target); err != nil {
		return err
	} else if mounted {
		if err := mounter.UMount(target); err != nil {
			return err
		}
	} else {
		log.Infof("BindUMount - target %s was already unmounted", target)
	}
	return nil
}

// HostPathToVolume returns absoluthe path to given volumeId on host mountpoint
func (mnt *orasHandler) HostPathToVolume(volumeId string) string {
	return path.Join(mnt.hostMountPath, mnt.pluginDataPath, volumesDirName, volumeId)
}

func (mnt *orasHandler) HostPathToMountVolume(volumeId string) string {
	return path.Join(mnt.artifactRoot, "mount_volumes", volumeId)
}

// MfsPathToVolume
func (mnt *orasHandler) OrasPathToVolume(volumeId string) string {
	return path.Join(mnt.hostMountPath, mnt.pluginDataPath, "mount_volumes", volumeId)
}

func (mnt *orasHandler) HostPathToLogs() string {
	return path.Join(mnt.hostMountPath, mnt.pluginDataPath, logsDirName)
}

func (mnt *orasHandler) HostPluginDataPath() string {
	return path.Join(mnt.hostMountPath, mnt.pluginDataPath)
}

func (mnt *orasHandler) HostPathTo(to string) string {
	return path.Join(mnt.hostMountPath, to)
}
