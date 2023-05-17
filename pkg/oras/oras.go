package oras

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/billy-playground/oras-csi/pkg/utils"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/natefinch/lumberjack.v2"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
)

const (
	manifestMediaTypeV1 = "application/vnd.oci.image.manifest.v1+json"
	manifestMediaTypeV2 = "application/vnd.docker.distribution.manifest.v2+json"
	newVolumeMode       = 0755
	logsDirName         = "logs"
	volumesDirName      = "volumes"
	mntDir              = "/mnt"
)

var log logrus.Logger

func Init(logLevel int) error {
	log = *logrus.New()
	log.SetLevel(logrus.Level(logLevel))
	return nil
}

type OrasHandler struct {
	name              string     // handler name
	testRun           bool       // is this just a test run?
	rootPath          string     // oras root path
	pluginDataPath    string     // plugin data path (inside rootPath)
	hostMountPath     string     // host mount path
	enforceNamespaces bool       // do not allow artifacts to cross namespaces
	manifests         *sync.Map  // inmemory cached manifest
	ociStore          *oci.Store // oci.Store cache
}

// NewOrasHandler creates a new oras handles to mount a container URI, pulled once
func NewOrasHandler(rootPath, pluginDataPath string, enforceNamespaces bool, name string, num ...int) (*OrasHandler, error) {
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

	store, err := oci.New(filepath.Join(rootPath, pluginDataPath))
	if err != nil {
		return nil, err
	}
	store.AutoSaveIndex = false
	return &OrasHandler{
		rootPath:          rootPath,
		pluginDataPath:    pluginDataPath,
		name:              name,
		hostMountPath:     path.Join(mntDir, fmt.Sprintf("%s%s", name, numSufix)),
		enforceNamespaces: enforceNamespaces,
		manifests:         &sync.Map{},
		ociStore:          store,
	}, nil
}

// SetOrasLogging sets up logging for the oras plugin
func (mnt *OrasHandler) SetOrasLogging() {
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

func (mnt *OrasHandler) CreateMountVolume(volumeId string) error {
	path := mnt.HostPathToMountVolume(volumeId)
	if err := os.MkdirAll(path, newVolumeMode); err != nil {
		return err
	}
	return nil
}

func (mnt *OrasHandler) CreateVolume(volumeId string, size int64) error {
	path := mnt.HostPathToVolume(volumeId)
	err := os.MkdirAll(path, newVolumeMode)
	return err
}

// Delete a volume (this isn't currently used, we need to design concept of cleanup)
func (mnt *OrasHandler) DeleteVolume(volumeId string) error {
	path := mnt.HostPathToVolume(volumeId)
	if err := os.RemoveAll(path); err != nil {
		log.Errorf("DeleteVolume -- Couldn't remove volume %s in directory %s. Error: %s",
			volumeId, path, err.Error())
		return err
	}

	return nil
}

// Pull the oras container to the plugin data directory
// Derived from https://github.com/sajayantony/csi-driver-host-path/blob/1bcc9d435d0c3ccd93fa1855e8669aad0f3bd1b5/pkg/oci/oci.go
// We are working on this plugin together
func (mnt *OrasHandler) OrasPull(artifactRoot string, settings orasSettings) error {
	// 0. caching the remote repository with OCI layout as proxy
	log.Infof("Remote repository %s will be proxied by %s", settings.rawReference, artifactRoot)
	repo, err := remote.NewRepository(settings.rawReference)
	if err != nil {
		return err
	}
	repo.PlainHTTP = settings.optionsPlainHttp

	proxy := New(repo, mnt.ociStore)
	blobs := New(repo.Blobs(), mnt.ociStore) //FIXME: just to accomodate the current implementation, can be replaced by oras.Copy if blob pull can be inlined

	// 1. resolve manifest descriptor
	var desc ocispec.Descriptor
	ctx := context.Background()
	if found, ok := mnt.manifests.Load(settings.rawReference); ok {
		log.Infof("Manifest cached for %s", settings.rawReference)
		desc = found.(ocispec.Descriptor)
	} else {
		// todo: uri-based request caching
		log.Infof("Resolving manifest descriptor for %s", settings.rawReference)
		desc, err = oras.Resolve(ctx, proxy, settings.rawReference, oras.DefaultResolveOptions)
		if err != nil {
			return err
		}
		// todo: cache entry expiration
		mnt.manifests.Store(settings.rawReference, desc)
	}

	// 2. fetch manifest
	log.Printf("Fetching manifest %s", utils.DescToString(desc))
	readCloser, err := proxy.Fetch(ctx, desc)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	// Read the pulled content
	content, err := content.ReadAll(readCloser, desc)
	if err != nil {
		return err
	}

	// We are expecting to find an ORAS manifest
	if desc.MediaType == manifestMediaTypeV2 {
		return fmt.Errorf("found docker manifest %s, was this pushed with ORAS?", desc.MediaType)
	}
	if desc.MediaType != manifestMediaTypeV1 {
		return fmt.Errorf("found unknown media type %s", desc.MediaType)
	}

	var manifest ocispec.Manifest
	err = json.Unmarshal(content, &manifest)
	if err != nil {
		return err
	}

	// Loop through layers to parse and selectively download blobs
	total := len(manifest.Layers)
	extractCount := 0
	for i, layer := range manifest.Layers {
		log.Infof("Pulling %s, %d of %d", layer.Digest, i+1, total)
		filename, found := layer.Annotations["org.opencontainers.image.title"]

		// This shouldn't happen, but you never know!
		if !found {
			log.Infof("layer with digest %s is missing org.opencontainers.image.title annotation", layer.Digest)
			continue
		}

		// Are we filtering to a custom content type?
		if len(settings.mediaTypes) > 0 && !utils.ListContains(settings.mediaTypes, layer.MediaType) {
			log.Infof("layer for %s has undesired media type %s", filename, layer.MediaType)
			continue
		}
		fullPath := path.Join(artifactRoot, filename)

		// Ensure directory exists
		err := os.MkdirAll(path.Dir(fullPath), os.ModePerm)
		if err != nil {
			return err
		}

		// TODO could have a "pull if different size" or similar here
		err = pullBlob(ctx, blobs, layer, fullPath)
		if err != nil {
			return err
		}
		extractCount += 1
	}
	if extractCount == 0 {
		log.Warningf("There were no layers extracted for reference %s:%s", settings.reference, settings.tag)
	}
	return nil
}

func (mnt *OrasHandler) BindMount(source string, target string, options ...string) error {
	mounter := Mounter{}
	log.Infof("BindMount - source: %s, target: %s, options: %v", source, target, options)
	if isMounted, err := mounter.IsMounted(target, mnt.testRun); err != nil {
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

func (mnt *OrasHandler) BindUMount(target string) error {
	mounter := Mounter{}
	log.Infof("BindUMount - target: %s", target)
	if mounted, err := mounter.IsMounted(target, mnt.testRun); err != nil {
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
func (mnt *OrasHandler) HostPathToVolume(volumeId string) string {
	return path.Join(mnt.hostMountPath, mnt.pluginDataPath, volumesDirName, volumeId)
}

func (mnt *OrasHandler) HostPathToMountVolume(volumeId string) string {
	return path.Join(mnt.hostMountPath, "mount_volumes", volumeId)
}

// OrasPathToVolume ensures the artifact exists, and returns it
func (mnt *OrasHandler) OrasPathToVolume(settings orasSettings) (string, error) {

	log.Infof("Oras - container: %s, target: %s", settings.reference, mnt.hostMountPath)

	// TODO need to be able to name this predictably!
	artifact := strings.ReplaceAll(strings.ReplaceAll(settings.reference, "/", "-"), ".", "-")

	// Ensure plugin data directory exists first
	pluginData := path.Join(mnt.rootPath, mnt.pluginDataPath)
	artifactDir := artifact + "-" + settings.tag
	artifactRoot := path.Join(pluginData, artifactDir)

	// If we enforce a namespace, must go under that
	log.Infof("Enforce namespaces: %t", mnt.enforceNamespaces)
	if mnt.enforceNamespaces {
		log.Infof("Enforcing artifact namespace to be under %s", settings.namespace)
		artifactRoot = path.Join(pluginData, settings.namespace, artifactDir)
	}

	// Ensure the artifact root exists
	if _, err := os.Stat(artifactRoot); os.IsNotExist(err) {
		os.MkdirAll(pluginData, os.ModePerm)
	}

	// Pull (or ensure artifact already exists)
	err := mnt.OrasPull(artifactRoot, settings)
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

func (mnt *OrasHandler) HostPathToLogs() string {
	return path.Join(mnt.hostMountPath, mnt.pluginDataPath, logsDirName)
}

func (mnt *OrasHandler) HostPluginDataPath() string {
	return path.Join(mnt.hostMountPath, mnt.pluginDataPath)
}

func (mnt *OrasHandler) HostPathTo(to string) string {
	return path.Join(mnt.hostMountPath, to)
}
