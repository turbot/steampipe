
package: zip

zip: version
	zip -r -D build/artifacts.zip build/*

version: build
	go run scripts/version.go

build: install
	yarn build

install: clean
	yarn install

clean:
	rm -rf build