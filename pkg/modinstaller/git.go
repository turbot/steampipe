package modinstaller

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

func getGitUrl(modName string) string {
	return fmt.Sprintf("https://%s", modName)
}

func getTags(repo string) ([]string, error) {
	// Create the remote with repository URL
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo},
	})

	// load remote references
	refs, err := rem.List(&git.ListOptions{})
	if err != nil {
		return nil, err
	}

	// filters the references list and only keeps tags
	var tags []string
	for _, ref := range refs {
		if ref.Name().IsTag() {
			tags = append(tags, ref.Name().Short())
		}
	}

	return tags, nil
}

func getTagVersionsFromGit(repo string, includePrerelease bool) (semver.Collection, error) {
	tags, err := getTags(repo)
	if err != nil {
		return nil, err
	}

	versions := make(semver.Collection, len(tags))
	// handle index manually as we may not add all tags - if we cannot parse them as a version
	idx := 0
	for _, raw := range tags {
		v, err := semver.NewVersion(raw)
		if err != nil {
			continue
		}

		if !includePrerelease && v.Metadata() != "" || v.Prerelease() != "" {
			continue
		}
		versions[idx] = v
		idx++
	}
	// shrink slice
	versions = versions[:idx]

	// sort the versions in REVERSE order
	sort.Sort(sort.Reverse(versions))
	return versions, nil
}
