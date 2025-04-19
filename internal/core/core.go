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

package core

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"Unbewohnte/gobooru-downloader/internal/booru"
	"Unbewohnte/gobooru-downloader/internal/config"
	"Unbewohnte/gobooru-downloader/internal/logger"
	"Unbewohnte/gobooru-downloader/internal/workerpool"

	"golang.org/x/time/rate"
)

const VERSION string = "0.3"

type Downloader struct {
	client       *http.Client
	limiter      *rate.Limiter
	pool         *workerpool.Pool[Job, Result]
	config       *config.Config
	shutdown     chan struct{}
	wg           sync.WaitGroup
	downloadedGB float64
	signalChan   chan os.Signal

	downloadedCount int
	totalCount      int
	startTime       time.Time
	lastBytes       float64
	lastTime        time.Time
}

func NewDownloader(cfg *config.Config) *Downloader {
	dl := &Downloader{
		client:       cfg.HTTPClient,
		limiter:      rate.NewLimiter(rate.Every(time.Second), int(cfg.WorkerCount)),
		pool:         workerpool.NewPool[Job, Result](cfg.WorkerCount),
		config:       cfg,
		shutdown:     make(chan struct{}),
		signalChan:   make(chan os.Signal, 1),
		downloadedGB: 0.0,
	}

	// Set up signal handling
	signal.Notify(dl.signalChan, os.Interrupt, syscall.SIGTERM)
	go dl.handleSignals()

	return dl
}

func (d *Downloader) Run() error {
	// Rest previous progress information
	d.downloadedCount = 0
	d.totalCount = 0
	d.startTime = time.Now()
	d.lastTime = time.Now()
	d.lastBytes = 0
	d.downloadedGB = 0.0

	// Start worker pool with our processing function
	d.pool.Start(d.workerFunc)

	// Handle results in background
	go d.handleResults()

	// Main download loop
	galleryURL := d.config.BooruURL
	currentPage := d.config.FromPage

	for {
		select {
		case <-d.shutdown:
			logger.Info("Shutting down...")
			return nil
		default:
			logger.Info("[Main] On page %d", currentPage)

			// Rate limit page requests
			if err := d.limiter.Wait(context.Background()); err != nil {
				logger.Error("[Main] Rate limiter error: %s", err)
				continue
			}

			// Get posts from current page
			posts, err := booru.GetPosts(*galleryURL, currentPage, d.config.Tags, d.client)
			if err != nil {
				logger.Error("[Main] Failed after retries: %s...", err)
				continue
			}

			// Submit posts to worker pool
			for _, post := range posts {
				select {
				case <-d.shutdown:
					return nil
				default:
					d.wg.Add(1)
					d.pool.Submit(NewJob(post))
				}
			}

			currentPage++
		}
	}
}

func (d *Downloader) Stop() error {
	close(d.shutdown)
	d.wg.Wait()
	d.pool.Shutdown()
	return nil
}

func (d *Downloader) handleSignals() {
	<-d.signalChan
	logger.Info("Caught interrupt, stopping...")
	d.Stop()
	os.Exit(0)
}

func (d *Downloader) workerFunc(j Job) Result {
	// Rate limit worker requests
	if err := d.limiter.Wait(context.Background()); err != nil {
		logger.Error("[Worker] Rate limiter error: %s", err)
		return NewResult(false, false, nil)
	}

	mediaName := path.Base(j.Post.MediaURL())

	// Apply filters
	if d.config.ImagesOnly && !j.Post.IsImage() {
		logger.Info("[Worker] Skipping %s, it's not an image", mediaName)
		return NewResult(false, true, j.Post.Metadata())
	}

	if d.config.VideosOnly && !j.Post.IsVideo() {
		logger.Info("[Worker] Skipping %s, it's not a video", mediaName)
		return NewResult(false, true, j.Post.Metadata())
	}

	if d.config.MaxFileSize != 0 {
		if j.Post.Size()/1024/1024 > uint64(d.config.MaxFileSize) {
			logger.Info("[Worker] Skipping %s because it's too large", mediaName)
			return NewResult(false, true, j.Post.Metadata())
		}
	}

	// Save media
	if err := j.Post.SaveMedia(d.config.OutputDir, d.client); err != nil {
		logger.Error("[Worker] Failed to save %s: %s", mediaName, err)
		return NewResult(false, false, j.Post.Metadata())
	}

	// Update downloaded GB after the image had been downloaded
	// d.downloadedGB += float64(j.Post.Size()) / 1024.0 / 1024.0

	// Save metadata
	if err := j.Post.SaveMetadata(d.config.OutputDir); err != nil {
		logger.Error("[Worker] Failed to save metadata for %s: %s", mediaName, err)
		return NewResult(false, false, j.Post.Metadata())
	}

	return NewResult(true, false, j.Post.Metadata())
}

func (d *Downloader) handleResults() {
	for result := range d.pool.GetResults() {
		d.totalCount++ // Increment total attempted count

		if result.Success {
			d.downloadedCount++ // Increment successful count
			logger.Info(
				"[Result] %s (%.02fMB)",
				result.Metadata.Hash,
				float64(result.Metadata.Size)/1024.0/1024.0,
			)
			d.downloadedGB += float64(result.Metadata.Size) / 1024.0 / 1024.0 / 1024.0
			d.lastBytes += float64(result.Metadata.Size)
		} else if !result.Skip && result.Metadata != nil {
			logger.Warning("[Result] Fail on %s", result.Metadata.URL)
		}
		d.wg.Done()
	}
}

type Progress struct {
	Downloaded   int
	Total        int
	SpeedKBps    float64
	DownloadedGB float64
}

// Update GetProgress to use the new fields and methods
func (d *Downloader) GetProgress() Progress {
	return Progress{
		Downloaded:   d.downloadedCount,
		Total:        d.totalCount,
		SpeedKBps:    d.calculateSpeed(),
		DownloadedGB: d.downloadedGB,
	}
}

// Add these new methods for progress calculation
func (d *Downloader) calculateSpeed() float64 {
	now := time.Now()
	elapsed := now.Sub(d.lastTime).Seconds()
	if elapsed == 0 {
		return 0
	}

	speedKBps := (d.lastBytes / 1024) / elapsed
	d.lastBytes = 0
	d.lastTime = now
	return speedKBps
}

func (d *Downloader) IsRunning() bool {
	select {
	case <-d.shutdown:
		return false
	default:
		return true
	}
}
