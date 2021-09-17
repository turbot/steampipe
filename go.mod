module github.com/turbot/steampipe

go 1.16

require (
	github.com/Machiel/slugify v1.0.1
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.14 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/ahmetb/go-linq v3.0.0+incompatible
	github.com/alecthomas/chroma v0.9.2
	github.com/bgentry/speakeasy v0.1.0
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/briandowns/spinner v1.16.0
	github.com/c-bata/go-prompt v0.2.5
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/containerd v1.4.8
	github.com/containerd/continuity v0.1.0 // indirect
	github.com/deislabs/oras v0.8.1
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/fatih/color v1.12.0
	github.com/fsnotify/fsnotify v1.5.1
	github.com/gertd/go-pluralize v0.1.7
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.7.4
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/go-playground/validator/v10 v10.9.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-hclog v0.16.2
	github.com/hashicorp/go-plugin v1.4.3
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/terraform v0.15.1
	github.com/hashicorp/yamux v0.0.0-20210826001029-26ff87cf9493 // indirect
	github.com/jedib0t/go-pretty/v6 v6.0.6
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/gows v0.3.0
	github.com/lib/pq v1.10.3
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-isatty v0.0.14
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.6.0
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18
	github.com/shirou/gopsutil v3.21.8+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stevenle/topsort v0.2.0
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/turbot/go-kit v0.2.2-0.20210730122803-1ecb35c27e98
	github.com/turbot/steampipe-plugin-sdk v1.5.1
	github.com/ugorji/go v1.2.6 // indirect
	github.com/ulikunitz/xz v0.5.10
	github.com/zclconf/go-cty v1.9.1
	github.com/zclconf/go-cty-yaml v1.0.2
	golang.org/x/crypto v0.0.0-20210915214749-c084706c2272 // indirect
	golang.org/x/net v0.0.0-20210916014120-12bc252f5db8 // indirect
	golang.org/x/sys v0.0.0-20210915083310-ed5796bab164 // indirect
	golang.org/x/text v0.3.7
	google.golang.org/genproto v0.0.0-20210916144049-3192f974c780 // indirect
	google.golang.org/grpc v1.40.0
	gopkg.in/ini.v1 v1.63.2 // indirect
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	gotest.tools/v3 v3.0.3 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace github.com/c-bata/go-prompt => github.com/turbot/go-prompt v0.2.6-steampipe.0.20210830083819-c872df2bdcc9

replace github.com/turbot/steampipe-plugin-sdk => github.com/turbot/steampipe-plugin-sdk v1.5.1-0.20210827170319-ff928325577c
