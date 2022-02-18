
steampipe:
	go build -o  /usr/local/bin/steampipe

dashboard_assets:
	$(MAKE) -C ui/dashboard

all:
	$(MAKE) -C pluginmanager
	$(MAKE) -C ui/dashboard
	go build -o  /usr/local/bin/steampipe
