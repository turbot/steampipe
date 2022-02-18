
package: zip

zip: version
	cd build && zip -r -D ../../../report/reportassets/assets.zip *

version: build
	go run scripts/version.go

build: install
	yarn build

install: clean
	yarn install

clean:
	pwd
	[ ! -e ../../report/reportassets/assets.zip ] || rm ../../report/reportassets/assets.zip
	rm -rf build