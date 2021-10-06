module github.com/turbot/steampipe

go 1.16

require (
	github.com/Machiel/slugify v1.0.1
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.22 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/ahmetb/go-linq v3.0.0+incompatible
	github.com/alecthomas/chroma v0.9.2
	github.com/bgentry/speakeasy v0.1.0
	github.com/bits-and-blooms/bitset v1.2.1 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/briandowns/spinner v1.11.1
	github.com/c-bata/go-prompt v0.2.5
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/containerd v1.4.11
	github.com/containerd/continuity v0.2.0 // indirect
	github.com/deislabs/oras v0.8.1
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/fatih/color v1.7.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gertd/go-pluralize v0.1.7
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.7.2
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-hclog v0.15.0
	github.com/hashicorp/go-plugin v1.4.1
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/terraform v0.15.1
	github.com/jedib0t/go-pretty/v6 v6.0.6
	github.com/karrick/gows v0.3.0
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.8.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-isatty v0.0.12
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/miekg/dns v1.1.25 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.2 // indirect
	github.com/opencontainers/selinux v1.8.5 // indirect
	github.com/otiai10/copy v1.2.0
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18
	github.com/shirou/gopsutil v3.20.11+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stevenle/topsort v0.0.0-20130922064739-8130c1d7596b
	github.com/turbot/go-kit v0.2.2-0.20210730122803-1ecb35c27e98
	github.com/turbot/steampipe-plugin-sdk v1.3.0
	github.com/ulikunitz/xz v0.5.8
	github.com/zclconf/go-cty v1.8.2
	github.com/zclconf/go-cty-yaml v1.0.2
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20211005215030-d2e5035098b3 // indirect
	golang.org/x/sys v0.0.0-20211004093028-2c5d950f24ef // indirect
	golang.org/x/text v0.3.7
	golang.org/x/tools v0.1.5 // indirect
	google.golang.org/genproto v0.0.0-20211005153810-c76a74d43a8e // indirect
	google.golang.org/grpc v1.41.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	gotest.tools/v3 v3.0.3 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace github.com/c-bata/go-prompt => github.com/turbot/go-prompt v0.2.6-steampipe.0.20210830083819-c872df2bdcc9

replace github.com/turbot/steampipe-plugin-sdk => github.com/turbot/steampipe-plugin-sdk v1.5.1-0.20210827170319-ff928325577c
