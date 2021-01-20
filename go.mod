module github.com/turbot/steampipe

go 1.14

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/ahmetb/go-linq v3.0.0+incompatible
	github.com/briandowns/spinner v1.11.1
	github.com/c-bata/go-prompt v0.2.5
	github.com/containerd/containerd v1.4.1
	github.com/deislabs/oras v0.8.1
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.7.0
	github.com/gertd/go-pluralize v0.1.7
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/google/uuid v1.1.2
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/go-plugin v1.3.0
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/karrick/gows v0.3.0
	github.com/kr/text v0.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lib/pq v1.8.0
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.2.0
	github.com/radovskyb/watcher v1.0.7
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18
	github.com/shirou/gopsutil v3.20.11+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/turbot/go-kit v0.0.0-20210119154454-db924443f736
	github.com/turbot/steampipe-plugin-sdk v0.0.0-20210119154656-52569ae6f10d
	github.com/ulikunitz/xz v0.5.8
	github.com/zclconf/go-cty v1.7.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

replace github.com/c-bata/go-prompt => github.com/binaek89/go-prompt v0.2.7-multiline-clearscreen
