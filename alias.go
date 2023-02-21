package oncall

import (
	"encoding/json"
)

type OnCallURLAliases map[string]string

func (a OnCallURLAliases) FindByAlias(alias string) (string, bool) {
	u, ok := a[alias]
	return u, ok
}

func parseAliases(src string, aliases *OnCallURLAliases) error {
	if err := json.Unmarshal([]byte(src), aliases); err != nil {
		return err
	}
	return nil
}
