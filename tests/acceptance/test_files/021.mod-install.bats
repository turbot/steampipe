load "$LIB_BATS_ASSERT/load.bash"
load "$LIB_BATS_SUPPORT/load.bash"

@test "list with no mods installed" {
  run steampipe mod list
  assert_output 'No mods installed'
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
└── github.com/turbot/steampipe-mod-aws-compliance@v0.10'
}

@test "install old version when latest already installed" {
  steampipe mod install github.com/turbot/steampipe-mod-aws-compliance
  run steampipe mod install github.com/turbot/steampipe-mod-aws-compliance@0.1
  assert_output '
Downgraded 1 mod:

local
└── github.com/turbot/steampipe-mod-aws-compliance@v0.1'
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
└── github.com/turbot/steampipe-mod-aws-compliance@v0.1'
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
└── github.com/turbot/steampipe-mod-aws-compliance@v0.1'
}

function teardown() {
  rm -rf .steampipe/
  rm -rf .mod.cache.json
  rm -rf mod.sp
}

function setup() {
  cd $FILE_PATH/test_data/mod_install
}
