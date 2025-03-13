package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"Unbewohnte/gobooru-downloader/booru"
	"Unbewohnte/gobooru-downloader/logger"
	"Unbewohnte/gobooru-downloader/proxy"
	"Unbewohnte/gobooru-downloader/util"
	"Unbewohnte/gobooru-downloader/workerpool"

	"golang.org/x/time/rate"
)

const VERSION string = "v0.1"

var (
	version     *bool   = flag.Bool("version", false, "Print version information and exit")
	booruURL    *string = flag.String("url", "https://danbooru.donmai.us/", "URL to the booru page (blank for danbooru.donmai.us)")
	proxyString *string = flag.String("proxy", "", "Set proxy connection string")
	workerCount *uint   = flag.Uint("workers", 12, "Set worker count")
	outputDir   *string = flag.String("output", "output", "Set output directory name")
	silent      *bool   = flag.Bool("silent", false, "Output nothing to the console")
	maxRetries  *uint   = flag.Uint("max-retries", 3, "Set max http request retry count")
	imagesOnly  *bool   = flag.Bool("images-only", false, "Save only images")
	videosOnly  *bool   = flag.Bool("videos-only", false, "Save only videos")
	tags        *string = flag.String("tags", "", "Set tags")
	fromPage    *uint   = flag.Uint("from-page", 1, "Set initial page number")
)

type Job struct {
	Post booru.Post
}

func NewJob(post booru.Post) Job {
	return Job{
		Post: post,
	}
}

type Result struct {
	Success  bool
	Skip     bool
	Metadata *booru.Metadata
}

func NewResult(success bool, skip bool, metadata *booru.Metadata) Result {
	return Result{
		Success:  success,
		Skip:     skip,
		Metadata: metadata,
	}
}

var (
	pool       *workerpool.Pool[Job, Result]
	workerFunc func(Job) Result
	httpClient *http.Client
	limiter    *rate.Limiter
	wg         sync.WaitGroup
	shutdown   chan struct{}
)

func init() {
	flag.Parse()

	// Process version
	if *version {
		fmt.Printf("GOBOORU-DOWNLOADER %v\n(C) 2025 Kasyanov Nikolay Alexeevich (Unbewohnte)\n", VERSION)
		os.Exit(0)
	}

	if *silent {
		logger.SetOutput(io.Discard)
	}

	// Process proxy
	if strings.TrimSpace(*proxyString) != "" {
		client, err := proxy.NewProxyClient(*proxyString)
		if err != nil {
			logger.Error("[Init] Failed to make a new proxy client: %s", err)
			os.Exit(1)
		}
		httpClient = client
	} else {
		httpClient = http.DefaultClient
	}

	// Set retry count
	util.MAXRETRIES = *maxRetries

	// Check if booruURL is a valid URL
	_, err := url.Parse(*booruURL)
	if err != nil {
		logger.Error("[Init] %s is not a valid URL: %s", *booruURL, err)
		os.Exit(1)
	}

	// Create output directory
	if strings.TrimSpace(*outputDir) == "" {
		*outputDir = "output"
	}

	err = os.MkdirAll(*outputDir, os.ModePerm)
	if err != nil {
		logger.Error("[Init] Failed to create %s: %s", *outputDir, err)
		os.Exit(1)
	}

	// Create a worker pool
	pool = workerpool.NewPool[Job, Result](*workerCount)

	// Rate limiter: Allow 1 request per second with a burst of 8
	limiter = rate.NewLimiter(rate.Every(time.Second), 8)

	// Shutdown channel
	shutdown = make(chan struct{})

	workerFunc = func(j Job) Result {
		// Wait for the rate limiter
		if err := limiter.Wait(context.Background()); err != nil {
			logger.Error("[Worker] Rate limiter error: %s", err)
			return NewResult(false, false, nil)
		}

		// Process booru post

		mediaName := path.Base(j.Post.MediaURL())

		if *imagesOnly && !j.Post.IsImage() {
			// Skip
			logger.Info("[Worker] Skipping %s", mediaName)
			return NewResult(false, true, j.Post.Metadata())
		}

		if *videosOnly && !j.Post.IsVideo() {
			// Skip
			logger.Info("[Worker] Skipping %s", mediaName)
			return NewResult(false, true, j.Post.Metadata())
		}

		// Save media
		err = j.Post.SaveMedia(
			*outputDir,
			httpClient,
		)
		if err != nil {
			logger.Error("[Worker] Failed to save %s: %s", mediaName, err)
			return NewResult(false, false, j.Post.Metadata())
		}

		// Save metadata
		err = j.Post.SaveMetadata(*outputDir)
		if err != nil {
			logger.Error("[Worker] Failed to save metadata for %+v: %s", mediaName, err)
			return NewResult(false, false, j.Post.Metadata())
		}

		return NewResult(true, false, j.Post.Metadata())
	}

	// Handle interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Info("Caught interrupt, stopping...")
		// Signal shutdown
		close(shutdown)

		// Wait for all jobs to complete
		wg.Wait()

		// Shutdown the worker pool
		pool.Shutdown()
		logger.Info("Worker pool is closed!")
		os.Exit(0)
	}()
}

func main() {
	// Launch worker pool
	pool.Start(workerFunc)

	// Print results
	go func() {
		for result := range pool.GetResults() {
			if result.Success {
				logger.Info("[Result] Done with %s", result.Metadata.Hash)
			} else if !result.Skip {
				logger.Warning("[Result] Fail")
			}
			wg.Done() // Mark job as done
		}
	}()

	// Get the engine runnin'
	galleryURL, _ := url.Parse(*booruURL)
	if *fromPage == 0 {
		*fromPage = 1
	}
	var currentPage uint = *fromPage

	for {
		select {
		case <-shutdown:
			// Stop accepting new jobs
			logger.Info("Shutting down...")
			return
		default:
			// Retrieve a new gallery page and find posts

			logger.Info("[Main] On page %d", currentPage)

			// Wait for the rate limiter
			if err := limiter.Wait(context.Background()); err != nil {
				logger.Error("[Main] Rate limiter error: %s", err)
				continue
			}

			// Retrieve posts (retry batteries included)
			posts, err := booru.GetPosts(*galleryURL, currentPage, *tags, httpClient)
			if err != nil {
				logger.Error("[Main] Failed after retries: %s... Skipping to the next page", err)
				continue
			}

			// Submit posts to the worker pool
			for _, post := range posts {
				select {
				case <-shutdown:
					// No more jobs
					logger.Info("Shutting down...")
					return
				default:
					wg.Add(1) // Track the job
					pool.Submit(NewJob(post))
				}
			}

			// Bump page number
			currentPage++
		}
	}
}
