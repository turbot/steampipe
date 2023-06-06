package steampipeconfig

//
//import (
//	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
//)
//
//type LoadModOption = func(mod *LoadModConfig)
//type LoadModConfig struct {
//	PostLoadHook       func(parseCtx *parse.ModParseContext) error
//	ReloadDependencies bool
//}
//
//func WithPostLoadHook(postLoadHook func(parseCtx *parse.ModParseContext) error) LoadModOption {
//	return func(config *LoadModConfig) {
//		config.PostLoadHook = postLoadHook
//	}
//}
//func WithReloadDependencies() LoadModOption {
//	return func(config *LoadModConfig) {
//		config.ReloadDependencies = true
//	}
//}
