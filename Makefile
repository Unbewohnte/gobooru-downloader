both: dir cli gui

gui: dir
	go build -tags gui -o bin/gobooru-downloader-gui ./cmd/gui

cli: dir
	go build -tags cli -o bin/gobooru-downloader ./cmd/cli

dir:
	mkdir -p bin