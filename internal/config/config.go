/*
   gobooru-downloader
   Copyright (C) 2025 Kasyanov Nikolay Alexeevich (Unbewohnte)

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package config

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"Unbewohnte/gobooru-downloader/internal/logger"
	"Unbewohnte/gobooru-downloader/internal/proxy"
)

type Config struct {
	Version         bool
	BooruURL        *url.URL
	ProxyString     string
	WorkerCount     uint
	OutputDir       string
	Silent          bool
	MaxRetries      uint
	ImagesOnly      bool
	VideosOnly      bool
	Tags            string
	FromPage        uint
	MaxFileSize     uint
	DownloadLimitGb float64
	HTTPClient      *http.Client
	NoMetadata      bool
}

func ParseFlags() *Config {
	var (
		version         = flag.Bool("version", false, "Print version information and exit")
		booruURL        = flag.String("url", "https://danbooru.donmai.us/", "URL to the booru page (blank for danbooru.donmai.us)")
		proxyString     = flag.String("proxy", "", "Set proxy connection string")
		workerCount     = flag.Uint("workers", 8, "Set worker count")
		outputDir       = flag.String("output", "output", "Set output directory name")
		silent          = flag.Bool("silent", false, "Output nothing to the console")
		maxRetries      = flag.Uint("max-retries", 3, "Set max http request retry count")
		imagesOnly      = flag.Bool("only-images", false, "Save only images")
		videosOnly      = flag.Bool("only-videos", false, "Save only videos")
		tags            = flag.String("tags", "", "Set tags")
		fromPage        = flag.Uint("from-page", 1, "Set initial page number")
		maxFileSize     = flag.Uint("max-filesize-mb", 0, "Set max file size in megabytes (0 for no cap)")
		downloadLimitGb = flag.Float64("download-limit-gb", 0.0, "Set download limit in gigabytes (0 for no cap)")
		noMetadata      = flag.Bool("no-metadata", false, "Do not save image metadata files")
	)

	flag.Parse()

	// Handle silent mode
	if *silent {
		logger.SetOutput(io.Discard)
	}

	// Apply cofiguration
	parsedURL, err := url.Parse(*booruURL)
	if err != nil {
		logger.Error("[Config] %s is not a valid URL: %s", *booruURL, err)
		os.Exit(1)
	}

	cfg := &Config{
		Version:         *version,
		BooruURL:        parsedURL,
		ProxyString:     *proxyString,
		WorkerCount:     *workerCount,
		OutputDir:       *outputDir,
		Silent:          *silent,
		MaxRetries:      *maxRetries,
		ImagesOnly:      *imagesOnly,
		VideosOnly:      *videosOnly,
		Tags:            *tags,
		FromPage:        *fromPage,
		MaxFileSize:     *maxFileSize,
		DownloadLimitGb: *downloadLimitGb,
		HTTPClient:      nil,
		NoMetadata:      *noMetadata,
	}

	cfg.Apply()
	return cfg
}

func (c *Config) PrintVersion(version string) {
	fmt.Printf(
		`gobooru-downloader v%s
Copyright (C) 2025  Kasyanov Nikolay Alexeevich (Unbewohnte)
This program comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it
under certain conditions (see COPYING)
`, version)
}

func ApplyConfig(cfg *Config) {
	// Handle silent mode
	if cfg.Silent {
		logger.SetOutput(io.Discard)
	}

	// Create output directory if needed
	if strings.TrimSpace(cfg.OutputDir) == "" {
		cfg.OutputDir = "output"
	}
	err := os.MkdirAll(cfg.OutputDir, os.ModePerm)
	if err != nil {
		logger.Error("[Config] Failed to create %s: %s", cfg.OutputDir, err)
		os.Exit(1)
	}

	// Set proxy retry count
	proxy.MAXRETRIES = cfg.MaxRetries

	// Create HTTP client
	var client *http.Client
	if strings.TrimSpace(cfg.ProxyString) != "" {
		client, err = proxy.NewProxyClient(cfg.ProxyString)
		if err != nil {
			logger.Error("[Config] Failed to create proxy client: %s", err)
			os.Exit(1)
		}
	} else {
		client = http.DefaultClient
	}

	cfg.HTTPClient = client
}

func (c *Config) Apply() {
	ApplyConfig(c)
}
