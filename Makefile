both: dir cli gui

gui: dir
	go build -o bin/gobooru-downloader-gui ./cmd/gui

cli: dir
	go build -o bin/gobooru-downloader ./cmd/cli

cross: clean dir
	mkdir -p bin/gobooru-downloader-linux && \
	GOOS=linux GOARCH=amd64 go build -o bin/gobooru-downloader-linux/gobooru-downloader ./cmd/cli ; \
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/gobooru-downloader-linux/gobooru-downloader-gui ./cmd/gui

	mkdir -p bin/gobooru-downloader-windows && \
	GOOS=windows GOARCH=amd64 go build -o bin/gobooru-downloader-windows/gobooru-downloader ./cmd/cli ; \
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o bin/gobooru-downloader-windows/gobooru-downloader-gui ./cmd/gui

dir:
	mkdir -p bin

clean:
	rm -rf bin