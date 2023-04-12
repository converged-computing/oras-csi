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
	rootPath       string // oras root path
	pluginDataPath string // plugin data path (inside rootPath)
	name           string // handler name
	hostMountPath  string // host mount path
}

// NewOrasHandler creates a new oras handles to mount a container URI, pulled once
func NewOrasHandler(rootPath, pluginDataPath, name string, num ...int) *orasHandler {
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

	return &orasHandler{
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
		log.Errorf("DeleteVolume -- Couldn't remove volume %s in directory %s. Error: %s",
			volumeId, path, err.Error())
		return err
	}

	return nil
}

// Ensure the artifact (and pull contents there)
func (mnt *orasHandler) ensureArtifact(artifactRoot string, container string) error {

	if _, err := os.Stat(artifactRoot); errors.Is(err, os.ErrNotExist) {

		log.Info("Artifact root does not exist, creating", artifactRoot)
		// Pull Oras Container, if doesn't exist
		if err := mnt.OrasPull(artifactRoot, container); err != nil {
			return err
		}
	} else {
		log.Info("Artifact root already exists, no need to re-create!")
	}
	return nil
}

// parseContainer URI into a container and a tag
func parseContainer(container string) (string, string) {
	container = strings.TrimPrefix(container, "oras://")
	tag := "latest"
	if strings.Contains(container, ":") {
		parts := strings.Split(container, ":")
		container = parts[0]
		tag = parts[1]
	}
	return container, tag
}

// Pull the oras container to the plugin data directory
func (mnt *orasHandler) OrasPull(artifactRoot string, container string) error {

	// Get rid of oras prefix, if provided
	container, tag := parseContainer(container)
	log.Infof("Found ORAS container: %s:%s", container, tag)

	// 0. Create a file store
	log.Info("Creating oras filestore at: ", artifactRoot)

	fs, err := file.New(artifactRoot)
	if err != nil {
		return err
	}
	defer fs.Close()

	// 1. Connect to a remote repository
	log.Info("Preparing to pull from remote repository: ", container)
	ctx := context.Background()
	repo, err := remote.NewRepository(container)
	if err != nil {
		return err
	}
	// If authentication needed, could be added here
	// 2. Copy from the remote repository to the file store
	_, err = oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}
	return nil
}

func (mnt *orasHandler) BindMount(source string, target string, options ...string) error {
	mounter := Mounter{}
	log.Infof("BindMount - source: %s, target: %s, options: %v", source, target, options)
	if isMounted, err := mounter.IsMounted(target); err != nil {
		return err
	} else if !isMounted {
		if err := mounter.Mount(source, target, "", append(options, "bind")...); err != nil {
			return err
		}
		log.Infof("Successfully mounted %s to %s", source, target)
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
	return path.Join(mnt.hostMountPath, "mount_volumes", volumeId)
}

// OrasPathToVolume ensures the artifact exists, and returns it
func (mnt *orasHandler) OrasPathToVolume(container string) (string, error) {

	log.Infof("Oras - container: %s, target: %s", container, mnt.hostMountPath)
	container, tag := parseContainer(container)

	// TODO need to be able to name this predictably!
	artifact := strings.ReplaceAll(strings.ReplaceAll(container, "/", "-"), ".", "-")

	// Ensure plugin data directory exists first
	pluginData := path.Join(mnt.rootPath, mnt.pluginDataPath)
	artifactRoot := path.Join(pluginData, artifact+"-"+tag)

	// TODO not sure if ORAS creates the artifact root for us
	// (and then we create the root, as we do here)
	if _, err := os.Stat(artifactRoot); os.IsNotExist(err) {
		os.MkdirAll(pluginData, os.ModePerm)
	}

	// Pull (or ensure artifact already exists)
	err := mnt.ensureArtifact(artifactRoot, container)
	if err != nil {
		return "", err
	}
	log.Infof("Oras artifact root: %s", artifactRoot)
	files, err := ioutil.ReadDir(artifactRoot)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		log.Info("Found artifact asset: ", f.Name())
	}
	return artifactRoot, nil
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
