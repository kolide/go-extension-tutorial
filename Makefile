.PHONY: build env osqueryd osqueryi deps

all: build

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -vendor-only

build:  
	echo "$(shell pwd)/build/tutorial-extension.ext" > /tmp/extensions.load
	go build -i -o build/tutorial-extension.ext ./cmd/extension

osqueryd: build
	osqueryd \
		--extensions_autoload=/tmp/extensions.load \
		--logger_plugin=journal \
		--pidfile=/tmp/osquery.pid \
		--database_path=/tmp/osquery.db \
		--extensions_socket=/tmp/osquery.sock \
		--config_refresh=60 \
		--config_plugin=gist 
		# --distributed_plugin=twitter \
		# --distributed_interval=10 \
		# --disable_distributed=false \

osqueryi: build
	osqueryi --extension=./build/tutorial-extension.ext

clean:
	rm -rf /tmp/extensions.load
	rm -rf /tmp/osquery.*
