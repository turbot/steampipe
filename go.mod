module github.com/turbot/steampipe

go 1.16

require (
	github.com/Machiel/slugify v1.0.1
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/ahmetb/go-linq v3.0.0+incompatible
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/briandowns/spinner v1.11.1
	github.com/c-bata/go-prompt v0.2.5
	github.com/containerd/containerd v1.4.1
	github.com/deislabs/oras v0.8.1
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.7.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/gertd/go-pluralize v0.1.7
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/google/uuid v1.2.0
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-hclog v0.15.0
	github.com/hashicorp/go-plugin v1.4.1
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/terraform v0.15.1
	github.com/jedib0t/go-pretty/v6 v6.0.6
	github.com/karrick/gows v0.3.0
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.8.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.2.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18
	github.com/shirou/gopsutil v3.20.11+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.7.1
	github.com/stevenle/topsort v0.0.0-20130922064739-8130c1d7596b
	github.com/turbot/go-kit v0.1.4-rc.1
	github.com/turbot/steampipe-plugin-sdk v0.2.8
	github.com/ulikunitz/xz v0.5.8
	github.com/zclconf/go-cty v1.8.2
	github.com/zclconf/go-cty-yaml v1.0.2
	golang.org/x/text v0.3.5
	google.golang.org/grpc v1.33.1
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

replace github.com/c-bata/go-prompt => github.com/turbot/go-prompt v0.2.6-steampipe
