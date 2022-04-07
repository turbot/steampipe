module github.com/turbot/steampipe

go 1.18

require (
	github.com/Machiel/slugify v1.0.1
	github.com/Masterminds/semver v1.5.0
	github.com/ahmetb/go-linq v3.0.0+incompatible
	github.com/alecthomas/chroma v0.10.0
	github.com/bgentry/speakeasy v0.1.0
	github.com/briandowns/spinner v1.18.1
	github.com/c-bata/go-prompt v0.2.6
	github.com/containerd/containerd v1.4.12
	github.com/deislabs/oras v0.8.1
	github.com/fatih/color v1.13.0
	github.com/fsnotify/fsnotify v1.5.1
	github.com/gertd/go-pluralize v0.2.0
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.7.7
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-hclog v1.2.0
	github.com/hashicorp/go-plugin v1.4.3
	github.com/hashicorp/go-version v1.4.0
	github.com/hashicorp/hcl/v2 v2.11.1
	github.com/hashicorp/terraform v0.15.1
	github.com/jackc/pgx/v4 v4.15.0
	github.com/jedib0t/go-pretty/v6 v6.3.0
	github.com/karrick/gows v0.3.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-isatty v0.0.14
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/olekukonko/tablewriter v0.0.5
	github.com/opencontainers/image-spec v1.0.2
	github.com/otiai10/copy v1.7.0
	github.com/sethvargo/go-retry v0.1.0
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	github.com/stevenle/topsort v0.0.0-20130922064739-8130c1d7596b
	github.com/turbot/go-kit v0.3.0
	github.com/xlab/treeprint v1.1.0
	github.com/zclconf/go-cty v1.10.0
	github.com/zclconf/go-cty-yaml v1.0.2
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/text v0.3.7
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.22 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/containerd/continuity v0.2.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/turbot/steampipe-plugin-sdk/v3 v3.0.1
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
)

require (
	github.com/gosuri/uilive v0.0.4 // indirect
	github.com/gosuri/uiprogress v0.0.1
)

replace github.com/c-bata/go-prompt => github.com/turbot/go-prompt v0.2.6-steampipe.0.20211124090719-0709bc8d8ce2
