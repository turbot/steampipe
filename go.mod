module github.com/turbot/steampipe

go 1.16

require (
	github.com/Machiel/slugify v1.0.1
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.22 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/ahmetb/go-linq v3.0.0+incompatible
	github.com/alecthomas/chroma v0.9.2
	github.com/bgentry/speakeasy v0.1.0
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/briandowns/spinner v1.11.1
	github.com/c-bata/go-prompt v0.2.6
	github.com/containerd/containerd v1.4.11
	github.com/containerd/continuity v0.2.0 // indirect
	github.com/deislabs/oras v0.8.1
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/fatih/color v1.9.0
	github.com/fsnotify/fsnotify v1.5.1
	github.com/gertd/go-pluralize v0.1.7
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.7.2
	github.com/go-git/go-git/v5 v5.4.2
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-hclog v0.15.0
	github.com/hashicorp/go-plugin v1.4.1
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/terraform v0.15.1
	github.com/jackc/pgx/v4 v4.14.0
	github.com/jedib0t/go-pretty/v6 v6.0.6
	github.com/karrick/gows v0.3.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-isatty v0.0.12
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.2.0
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/sethvargo/go-retry v0.1.0
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18
	github.com/shirou/gopsutil v3.21.10+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stevenle/topsort v0.0.0-20130922064739-8130c1d7596b
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/turbot/go-kit v0.3.0
	github.com/turbot/steampipe-plugin-sdk v1.8.0
	github.com/ulikunitz/xz v0.5.8
	github.com/zclconf/go-cty v1.8.2
	github.com/zclconf/go-cty-yaml v1.0.2
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20211011170408-caeb26a5c8c0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/text v0.3.7
	google.golang.org/genproto v0.0.0-20211013025323-ce878158c4d4 // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace github.com/c-bata/go-prompt => github.com/turbot/go-prompt v0.2.6-steampipe.0.20211124090719-0709bc8d8ce2
