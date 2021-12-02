package mod_installer

import (
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	goVersion "github.com/hashicorp/go-version"
)

func GetTags(repo string) ([]string, error) {
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

func GetTagVersionsFromGit(repo string) ([]*goVersion.Version, error) {
	tags, err := GetTags(repo)
	if err != nil {
		return nil, err
	}

	versions := make(goVersion.Collection, len(tags))
	// handle index manually as we may not add all tags - if we cannot parse them as a version
	idx := 0
	for _, raw := range tags {
		v, err := goVersion.NewVersion(raw)
		if err != nil {
			continue
		}
		versions[idx] = v
		idx++
	}

	// sort the versions in REVERSE order
	sort.Reverse(versions)
	return versions, nil
}
