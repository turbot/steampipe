
steampipe:
	go build -o  /usr/local/bin/steampipe

report_assets:
	$(MAKE) -C ui/report

all:
	$(MAKE) -C pluginmanager
	$(MAKE) -C ui/report
	go build -o  /usr/local/bin/steampipe
