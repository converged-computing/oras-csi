package oras

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"oras.land/oras-go/v2/registry"
)

// Oras Settings come from volume attributes
type orasSettings struct {
	namespace    string
	reference    string // Fully qualified reference with registry and no tag
	mediaTypes   []string
	rawReference string
	registry     string

	// ORAS options
	optionsPlainHttp   bool
	optionsInSecure    bool
	optionsConcurrency int
	optionsPullAlways  bool
	tag                string
}

// parseContainer URI into a reference(registry+repository) and a tag
func (settings *orasSettings) parseReference(reference string) error {
	artifactRef := strings.TrimPrefix(reference, "oras://")
	tag := "latest"

	// if the reference doesn't have a / then add a docker.io/library prefix
	// this is to support the docker.io library that is implicit
	// in docker pull commands
	if !strings.Contains(artifactRef, "/") {
		artifactRef = "docker.io/library/" + artifactRef
	}

	ref, err := registry.ParseReference(artifactRef)
	if err != nil {
		return err
	}

	// Oras uses the ref to mean digest or tag
	if ref.Reference != "" {
		tag = ref.Reference
	}

	settings.rawReference = reference
	settings.registry = ref.Registry
	settings.reference = ref.Registry + "/" + ref.Repository
	settings.tag = tag
	return nil
}

// NewSettings parses volume attributes and returns the settings
func NewSettings(volumeContext map[string]string) (orasSettings, error) {

	settings := orasSettings{}

	// oras.artifact.reference (required)
	reference, found := volumeContext["oras.artifact.reference"]
	if !found {
		return settings, status.Error(codes.InvalidArgument, "oras.artifact.reference is required")
	}

	// Split the reference into first part, and tag
	err := settings.parseReference(reference)
	if err != nil {
		return settings, status.Error(codes.InvalidArgument, fmt.Sprintf("issue parsing oras.artifact.reference %s", err))
	}

	// oras.options.plainhttp
	value, found := volumeContext["oras.options.plainhttp"]
	if found {
		plainhttp, err := strconv.ParseBool(value)
		if err != nil {
			return settings, status.Error(codes.InvalidArgument, fmt.Sprintf("issue parsing oras.options.plainhttp %s", err))
		}
		settings.optionsPlainHttp = plainhttp
	}

	// oras.options.insecure
	value, found = volumeContext["oras.options.insecure"]
	if found {
		insecure, err := strconv.ParseBool(value)
		if err != nil {
			return settings, status.Error(codes.InvalidArgument, fmt.Sprintf("issue parsing oras.options.insecure %s", err))
		}
		settings.optionsInSecure = insecure
	}

	// concurrency for downloads (defaults to 1)
	value, found = volumeContext["oras.options.concurrency"]
	settings.optionsConcurrency = 1
	if found {
		concurrency, err := strconv.Atoi(value)
		if err != nil || concurrency <= 0 {
			return settings, status.Error(codes.InvalidArgument, fmt.Sprintf("issue parsing oras.options.concurrency %s", err))
		}
		settings.optionsConcurrency = concurrency
	}

	// namespace for pod, can be used to enforce artifact pull directory structure
	// This should always be set, to default when not set in YAML
	value, found = volumeContext["csi.storage.k8s.io/pod.namespace"]
	if found {
		settings.namespace = value
	}

	// oras.options.pullalways means we pull always, regardless of existence
	value, found = volumeContext["oras.options.pullalways"]
	if found {
		pullAlways, err := strconv.ParseBool(value)
		if err != nil {
			return settings, status.Error(codes.InvalidArgument, fmt.Sprintf("issue parsing oras.options.pullalways %s", err))
		}
		settings.optionsPullAlways = pullAlways
	}
	// oras.artifact.layers.mediatype is a comma separated list to filter
	value, found = volumeContext["oras.artifact.layers.mediatype"]
	settings.mediaTypes = []string{}
	if found {
		settings.mediaTypes = strings.Split(value, ",")
	}
	return settings, nil
}
