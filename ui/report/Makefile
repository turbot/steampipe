
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
	rm -rf build