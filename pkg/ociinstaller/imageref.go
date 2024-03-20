package ociinstaller

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
)

const (
	DefaultImageTag            = "latest"
	DefaultImageRepoActualURL  = "ghcr.io/turbot/steampipe"
	DefaultImageRepoDisplayURL = "hub.steampipe.io"

	DefaultImageOrg  = "turbot"
	DefaultImageType = "plugins"
)

// SteampipeImageRef a struct encapsulating a ref to an OCI image
type SteampipeImageRef struct {
	requestedRef string
}

// NewSteampipeImageRef creates and returns a New SteampipeImageRef
func NewSteampipeImageRef(ref string) *SteampipeImageRef {
	ref = sanitizeRefStream(ref)
	return &SteampipeImageRef{
		requestedRef: ref,
	}
}

// ActualImageRef returns the actual, physical full image ref
// (ghcr.io/turbot/steampipe/plugins/turbot/aws:1.0.0)
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

// DisplayImageRef returns the "friendly" user-facing full image ref
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

// DisplayImageRefConstraintOverride returns a "friendly" user-facing version of the image ref
// but with the version replaced by provided constraint
// (hub.steampipe.io/plugins/turbot/aws@^1.0)
func (r *SteampipeImageRef) DisplayImageRefConstraintOverride(constraint string) string {
	dir := r.DisplayImageRef()
	s := strings.Split(dir, "@")
	return fmt.Sprintf("%s@%s", s[0], constraint)
}

func isDigestRef(ref string) bool {
	return strings.Contains(ref, "@sha256:")
}

// sanitizes the ref to exclude any 'v' prefix
// in the stream (if any)
func sanitizeRefStream(ref string) string {
	if !isDigestRef(ref) {
		splitByAt := strings.Split(ref, "@")
		if len(splitByAt) == 1 {
			// no stream mentioned
			return ref
		}
		// trim out the 'v' prefix
		splitByAt[1] = strings.TrimPrefix(splitByAt[1], "v")

		ref = strings.Join(splitByAt, "@")
	}
	return ref
}

func (r *SteampipeImageRef) IsFromSteampipeHub() bool {
	return strings.HasPrefix(r.DisplayImageRef(), constants.SteampipeHubOCIBase)
}

// GetOrgNameAndConstraint splits the full image reference into (org, name, constraint)
// Constraint will be either a SemVer version (1.2.3) or a SemVer constraint (^0.4)
func (r *SteampipeImageRef) GetOrgNameAndConstraint() (string, string, string) {
	// plugin.Name looks like `hub.steampipe.io/plugins/turbot/aws@latest`
	split := strings.Split(r.DisplayImageRef(), "/")
	pluginNameAndSuffix := strings.Split(split[len(split)-1], "@")
	if r.IsFromSteampipeHub() {
		return split[len(split)-2], pluginNameAndSuffix[0], pluginNameAndSuffix[1]
	}
	return strings.Join(split[0:len(split)-1], "/"), pluginNameAndSuffix[0], pluginNameAndSuffix[1]
}

// GetFriendlyName returns the minimum friendly name so that the original name can be rebuilt using preset defaults:
// hub.steampipe.io/plugins/turbot/aws@1.0.0    => aws@1.0.0
// hub.steampipe.io/plugins/turbot/aws@latest   => aws
// hub.steampipe.io/plugins/otherOrg/aws@latest => otherOrg/aws
// hub.steampipe.io/plugins/otherOrg/aws@1.0.0  => otherOrg/aws@1.0.0
// differentRegistry.com/otherOrg/aws@latest    => differentRegistry.com/otherOrg/aws@latest
// differentRegistry.com/otherOrg/aws@1.0.0     => differentRegistry.com/otherOrg/aws@1.0.0
func (r *SteampipeImageRef) GetFriendlyName() string {
	return getCondensedImageRef(r.DisplayImageRef())
}

func getCondensedImageRef(imageRef string) string {
	// if this is not from the default steampipe registry - DO NOT CONDENSE - return as is
	// (we are not aware of any conventions in the registry)
	if !strings.HasPrefix(imageRef, DefaultImageRepoDisplayURL) {
		return imageRef
	}

	// So this is an image reference from the Steampipe HUB registry
	// remove the registry URL
	ref := strings.TrimPrefix(imageRef, DefaultImageRepoDisplayURL)
	// remove the 'plugins' namespace where steampipe hub keeps the images
	ref = strings.TrimPrefix(ref, "/plugins/")
	// remove the default organization - "turbot"
	ref = strings.TrimPrefix(ref, DefaultImageOrg)
	// remove any leading '/'
	ref = strings.TrimPrefix(ref, "/")
	// remove the '@latest' tag (not others)
	ref = strings.TrimSuffix(ref, fmt.Sprintf("@%s", DefaultImageTag))

	return ref
}

// possible formats include
//		ghcr.io/turbot/steampipe/plugins/turbot/aws:1.0.0
//		ghcr.io/turbot/steampipe/plugins/turbot/aws@sha256:766389c9dd892132c7e7b9124f446b9599a80863d466cd1d333a167dedf2c2b1
//		turbot/aws:1.0.0
//		turbot/aws
//      dockerhub.org/myimage
//      dockerhub.org/myimage:mytag
//		aws:1.0.0
//		aws
//		ghcr.io/turbot/steampipe/plugins/turbot/aws@1.0.0
//		ghcr.io/turbot/steampipe/plugins/turbot/aws@sha256:766389c9dd892132c7e7b9124f446b9599a80863d466cd1d333a167dedf2c2b1
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
	default: //ex: ghcr.io/turbot/steampipe/plugins/turbot/aws
		return fmt.Sprintf("%s:%s", items[0], tag)
	}
}
