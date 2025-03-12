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
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"Unbewohnte/gobooru-downloader/booru"
	"Unbewohnte/gobooru-downloader/logger"
	"Unbewohnte/gobooru-downloader/proxy"
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
)

type Job struct {
	PostURL url.URL
}

func NewJob(postURL url.URL) Job {
	return Job{
		PostURL: postURL,
	}
}

type Result struct {
	Success bool            `json:"success"`
	Info    *booru.PostInfo `json:"info"`
}

func NewResult(success bool, postInfo *booru.PostInfo) Result {
	return Result{
		Success: success,
		Info:    postInfo,
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
			return NewResult(false, nil)
		}

		// Retry logic with exponential backoff
		var err error
		var postInfo *booru.PostInfo
		for attempt := 1; uint(attempt) <= *maxRetries; attempt++ {
			postInfo, err = booru.ProcessPost(httpClient, j.PostURL)
			if err == nil {
				break
			}

			logger.Error("[Worker] Attempt %d on %s: %s", attempt, j.PostURL.String(), err)

			// Exponential backoff
			if uint(attempt) != *maxRetries {
				time.Sleep(time.Duration(attempt*attempt) * time.Second)
			}
		}

		if err != nil {
			logger.Error("[Worker] Failed after retries: %s", err)
			return NewResult(false, nil)
		}

		// Save image
		imageHash, err := booru.SaveImage(
			httpClient,
			postInfo.ImageURL,
			*outputDir,
		)
		if err != nil {
			logger.Error("[Worker] Failed to save image: %s", err)
			return NewResult(false, nil)
		}

		// Save metadata to file
		err = booru.SaveMetadataJson(
			booru.NewMetadata(*postInfo, j.PostURL.Hostname(), imageHash),
			filepath.Join(*outputDir, "metadata.json"),
		)
		if err != nil {
			logger.Error("[Worker] Failed to save metadata for %s: %s", imageHash, err)
			return NewResult(false, nil)
		}

		return NewResult(true, postInfo)
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
				logger.Info("[Result] Done with %s", result.Info.ImageURL)
			} else {
				logger.Warning("[Result] Fail")
			}
			wg.Done() // Mark job as done
		}
	}()

	// Get the engine runnin'
	galleryURL, _ := url.Parse(*booruURL)
	var currentPage uint64 = 0

	for {
		select {
		case <-shutdown:
			// Stop accepting new jobs
			logger.Info("Shutting down...")
			return
		default:
			// Retrieve a new gallery page and find posts
			currentPage++

			// Create a copy of the original galleryURL to preserve its query
			pageURL := *galleryURL
			query := pageURL.Query()
			query.Set("page", fmt.Sprintf("%d", currentPage))
			pageURL.RawQuery = query.Encode()

			logger.Info("[Main] On %s", pageURL.String())

			// Wait for the rate limiter
			if err := limiter.Wait(context.Background()); err != nil {
				logger.Error("[Main] Rate limiter error: %s", err)
				continue
			}

			// Retry logic with exponential backoff
			var posts []string
			var err error
			for attempt := 1; uint(attempt) <= *maxRetries; attempt++ {
				posts, err = booru.GetPosts(httpClient, pageURL)
				if err == nil {
					break
				}

				// Log the error
				logger.Error("[Main] Attempt %d on %s: %s", attempt, pageURL.String(), err)

				// Exponential backoff
				if uint(attempt) != *maxRetries {
					time.Sleep(time.Duration(attempt*attempt) * time.Second)
				}
			}

			if err != nil {
				logger.Error("[Main] Failed after retries: %s", err)
				continue
			}

			// Submit posts to the worker pool
			pageURL.RawQuery = ""
			pageURL.Path = "/"
			for _, post := range posts {
				select {
				case <-shutdown:
					// No more jobs
					logger.Info("Shutting down...")
					return
				default:
					postURL, err := url.Parse(pageURL.String() + post[1:])
					if err != nil {
						logger.Error("[Main] Constructed an invalid post URL: %s. Skipping all posts for this gallery", err)
						break
					}

					wg.Add(1) // Track the job
					pool.Submit(NewJob(*postURL))
				}
			}
		}
	}
}
