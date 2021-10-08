
all:
	go build -o  /usr/local/bin/steampipe
	$(MAKE) -C plugin_manager

