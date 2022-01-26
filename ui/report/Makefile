
package: zip

zip: version
	zip -r -D build/artifacts.zip build/*

version: build
	echo "versioning"

build: install
	yarn build

install: clean
	yarn install

clean:
	rm -rf build