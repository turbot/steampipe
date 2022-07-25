
steampipe:
	go1.19rc1  build -o  /usr/local/bin/steampipe

dashboard_assets:
	$(MAKE) -C ui/dashboard

all:
	$(MAKE) -C pluginmanager
	$(MAKE) -C ui/dashboard
	go1.19rc1  build -o  /usr/local/bin/steampipe
