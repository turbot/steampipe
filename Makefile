
all:
	$(MAKE) -C plugin_manager
	go build -o  /usr/local/bin/steampipe

