
steampipe:
	go build -o  /usr/local/bin/steampipe

all:
	$(MAKE) -C pluginmanager
	$(MAKE) -C ui/report
	go build -o  /usr/local/bin/steampipe
