OUTPUT_DIR?=/usr/local/bin

build:
	$(eval TIMESTAMP := $(shell date +%Y%m%d%H%M%S))
	$(eval GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD | sed 's/[\/_]/-/g' | sed 's/[^a-zA-Z0-9.-]//g'))

	go build -o $(OUTPUT_DIR) -ldflags "-X main.version=0.0.0-dev-$(GIT_BRANCH).$(TIMESTAMP)" .

steampipe:
	go build -o ${OUTPUT_DIR}/steampipe

dashboard_assets:
	$(MAKE) -C ui/dashboard

all:
	$(MAKE) -C pkg/pluginmanager_service
	$(MAKE) -C ui/dashboard
	go build -o ${OUTPUT_DIR}/steampipe
