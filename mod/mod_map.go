package mod

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ModMap :: map of mod name to mod-version map
type ModMap map[string]*modconfig.Mod

func (m ModMap) String() string {
	var modStrings []string
	for _, mod := range m {
		modStrings = append(modStrings, mod.String())
	}
	return strings.Join(modStrings, "\n")
}

func (m ModMap) BuildNamedQueryMap() map[string]*modconfig.Query {
	//  build a list of long and short names for these queries
	var res = make(map[string]*modconfig.Query)
	var shortNameMap = make(map[string][]*modconfig.Query)

	for _, mod := range m {
		for _, q := range mod.Queries {
			longName := fmt.Sprintf("query.%s.%s", mod.Name, q.Name)
			shortName := fmt.Sprintf("query.%s", q.Name)
			res[longName] = q

			if matchingShortNameQueries, ok := shortNameMap[shortName]; ok {
				shortNameMap[shortName] = append(matchingShortNameQueries, q)
			} else {
			}
			shortNameMap[shortName] = []*modconfig.Query{q}
		}
	}
	var shortNames []string
	for shortName, queries := range shortNameMap {
		if len(queries) == 1 {
			res[shortName] = queries[0]
			shortNames = append(shortNames, shortName)
		}
	}
	return res
}
