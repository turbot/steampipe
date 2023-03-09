package ociinstaller

import (
	"fmt"
	"strings"
)

const (
	DefaultImageTag            = "latest"
	DefaultImageRepoActualURL  = "us-docker.pkg.dev/steampipe"
	DefaultImageRepoDisplayURL = "hub.steampipe.io"

	DefaultImageOrg  = "turbot"
	DefaultImageType = "plugins"
)

// SteampipeImageRef :: a ref to an OCI image
type SteampipeImageRef struct {
	requestedRef string
}

// NewSteampipeImageRef :: creates and returns a New SteampipeImageRef
func NewSteampipeImageRef(ref string) *SteampipeImageRef {
	return &SteampipeImageRef{
		requestedRef: ref,
	}
}

// ActualImageRef :: returns the actual, physical full image ref
// (us-docker.pkg.dev/steampipe/plugins/turbot/aws:1.0.0)
func (r *SteampipeImageRef) ActualImageRef() string {
	ref := r.requestedRef

	if !isDigestRef(ref) {
		ref = strings.ReplaceAll(ref, "@", ":")
	}

	fullRef := getFullImageRef(ref)

	if strings.HasPrefix(fullRef, DefaultImageRepoDisplayURL) {
		fullRef = strings.ReplaceAll(fullRef, DefaultImageRepoDisplayURL, DefaultImageRepoActualURL)
	}

	return fullRef
}

// DisplayImageRef :: returns the "friendly" user-facing full image ref
// (hub.steampipe.io/plugins/turbot/aws@1.0.0)
func (r *SteampipeImageRef) DisplayImageRef() string {
	fullRef := r.ActualImageRef()
	if isDigestRef(fullRef) {
		fullRef = strings.ReplaceAll(fullRef, ":", "-")
	}
	fullRef = strings.ReplaceAll(fullRef, ":", "@")

	if strings.HasPrefix(fullRef, DefaultImageRepoActualURL) {
		fullRef = strings.ReplaceAll(fullRef, DefaultImageRepoActualURL, DefaultImageRepoDisplayURL)
	}

	return fullRef
}

func isDigestRef(ref string) bool {
	return strings.Contains(ref, "@sha256:")
}

// GetOrgNameAndStream :: splits the full image reference into
// (org, name, stream)
func (r *SteampipeImageRef) GetOrgNameAndStream() (string, string, string) {
	// plugin.Name looks like `hub.steampipe.io/plugins/turbot/aws@latest`
	split := strings.Split(r.DisplayImageRef(), "/")

	org := split[len(split)-2]
	pluginNameAndStream := strings.Split(split[len(split)-1], "@")

	return org, pluginNameAndStream[0], pluginNameAndStream[1]
}

// possible formats include
//		us-docker.pkg.dev/steampipe/plugin/turbot/aws:1.0.0
//		us-docker.pkg.dev/steampipe/plugin/turbot/aws@sha256:766389c9dd892132c7e7b9124f446b9599a80863d466cd1d333a167dedf2c2b1
//		turbot/aws:1.0.0
//		turbot/aws
//      dockerhub.org/myimage
//      dockerhub.org/myimage:mytag
//		aws:1.0.0
//		aws
//		us-docker.pkg.dev/steampipe/plugin/turbot/aws@1.0.0
//		us-docker.pkg.dev/steampipe/plugin/turbot/aws@sha256:766389c9dd892132c7e7b9124f446b9599a80863d466cd1d333a167dedf2c2b1
//		turbot/aws@1.0.0
//      dockerhub.org/myimage@mytag
//		aws@1.0.0
//		hub.steampipe.io/plugin/turbot/aws@1.0.0

func getFullImageRef(imagePath string) string {

	tag := DefaultImageTag

	// Get the tag, default to `latest`
	items := strings.Split(imagePath, ":")
	if len(items) > 1 {
		tag = items[1]
	}

	// Image path
	parts := strings.Split(items[0], "/")
	switch len(parts) {
	case 1: //ex:  aws
		return fmt.Sprintf("%s/%s/%s/%s:%s", DefaultImageRepoActualURL, DefaultImageType, DefaultImageOrg, parts[len(parts)-1], tag)
	case 2: //ex:   turbot/aws OR dockerhub.com/my-image
		org := parts[len(parts)-2]
		if strings.Contains(org, ".") {
			return fmt.Sprintf("%s:%s", items[0], tag)
		}
		return fmt.Sprintf("%s/%s/%s/%s:%s", DefaultImageRepoActualURL, DefaultImageType, org, parts[len(parts)-1], tag)
	default: //ex: us-docker.pkg.dev/steampipe/plugin/turbot/aws
		return fmt.Sprintf("%s:%s", items[0], tag)
	}
}
