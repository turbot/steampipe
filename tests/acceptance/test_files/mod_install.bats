load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "list with no mods installed" {
  run steampipe mod list
  assert_output 'No mods installed.'
}

@test "install latest" {
  run steampipe mod install github.com/turbot/steampipe-mod-aws-compliance
  assert_output --partial 'Installed 1 mod:

local
└── github.com/turbot/steampipe-mod-aws-compliance'
  # need the check the version from mod.sp file as well
}

@test "install latest and then run install" {
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance
  run steampipe mod install
  assert_output 'All mods are up to date'
}

@test "install mod and list" {
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance@0.10
  run steampipe mod list
  assert_output '
local
└── github.com/turbot/steampipe-mod-aws-compliance@v0.10.0'
}

@test "install old version when latest already installed" {
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance
  run steampipe mod install github.com/turbot/steampipe-mod-aws-compliance@0.1
  assert_output '
Downgraded 1 mod:

local
└── github.com/turbot/steampipe-mod-aws-compliance@v0.1.0'
}

@test "install mod version, remove .steampipe folder and then run install" {
  # install particular mod version, remove .steampipe folder and run mod install
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance@0.1
  rm -rf .steampipe
  run steampipe mod install

  # should install the same cached version
  # better message
  assert_output '
Installed 1 mod:

local
└── github.com/turbot/steampipe-mod-aws-compliance@v0.1.0'
}

@test "install mod version, remove .cache file and then run install" {
  # install particular mod version, remove .mod.cache.json file and run mod install
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance@0.1
  rm -rf .mod.cache.json
  run steampipe mod install

  # should install the same cached version
  # better message
  assert_output '
Installed 1 mod:

local
└── github.com/turbot/steampipe-mod-aws-compliance@v0.1.0'
}

@test "install mod version should fail, since dependant mod has a requirement of different steampipe CLI version" {
  run steampipe mod install github.com/pskrbasu/steampipe-mod-m4
  assert_output --partial 'does not satisfy mod.m4 which requires version 10.99.99'
}

@test "install a mod with protocol in url" {
  run steampipe mod install https://github.com/turbot/steampipe-mod-hackernews-insights@0.3.0
  # should install with the protocol in the url prefix
  assert_output '
Installed 1 mod:

local
└── github.com/turbot/steampipe-mod-hackernews-insights@v0.3.0'
}

# Installed 4 mods:

# local
# └── github.com/pskrbasu/steampipe-mod-top-level@v3.0.0
#     ├── github.com/pskrbasu/steampipe-mod-dependency-1@v4.0.0
#     └── github.com/pskrbasu/steampipe-mod-dependency-2@v3.0.0
#         └── github.com/pskrbasu/steampipe-mod-dependency-1@v3.0.0
@test "complex mod dependency resolution - test tree structure" {
  run steampipe mod install github.com/pskrbasu/steampipe-mod-top-level
  # test the tree structure output
  assert_output '
Installed 4 mods:

local
└── github.com/pskrbasu/steampipe-mod-top-level@v3.0.0
    ├── github.com/pskrbasu/steampipe-mod-dependency-1@v4.0.0
    └── github.com/pskrbasu/steampipe-mod-dependency-2@v3.0.0
        └── github.com/pskrbasu/steampipe-mod-dependency-1@v3.0.0'
}

@test "complex mod dependency resolution - test benchmark and controls resolution 1" {
  steampipe mod install github.com/pskrbasu/steampipe-mod-top-level

  run steampipe check top_level.benchmark.bm_version_dependency_mod_1 --output csv
  # check the output - benchmark should run the control and query from dependency mod 1 which will
  # have the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | 4      | 4        | alarm  |
# +--------+----------+--------+
  assert_output 'group_id,title,description,control_id,control_title,control_description,reason,resource,status,severity
top_level.benchmark.bm_version_dependency_mod_1,Benchmark version dependency mod 1,,dependency_1.control.version,,,4,4,alarm,'
}

@test "complex mod dependency resolution - test benchmark and controls resolution 2" {
  steampipe mod install github.com/pskrbasu/steampipe-mod-top-level

  run steampipe check top_level.benchmark.bm_version_dependency_mod_2 --output csv
  # check the output - benchmark should run the control and query from dependency mod 2 which will
  # have the output:
# +--------+----------+--------+
# | reason | resource | status |
# +--------+----------+--------+
# | 3      | 3        | ok     |
# +--------+----------+--------+
  assert_output 'group_id,title,description,control_id,control_title,control_description,reason,resource,status,severity
top_level.benchmark.bm_version_dependency_mod_2,Benchmark version dependency mod 2,,dependency_2.control.version,,,3,3,ok,'
}

function teardown() {
  steampipe plugin uninstall aws
  rm -rf .steampipe/
  rm -rf .mod.cache.json
  rm -rf mod.sp
}

function setup() {
  cd $FILE_PATH/test_data/mod_install
  steampipe plugin install aws
}
