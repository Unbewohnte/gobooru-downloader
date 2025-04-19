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

package gui

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"Unbewohnte/gobooru-downloader/internal/config"
	"Unbewohnte/gobooru-downloader/internal/core"
	"Unbewohnte/gobooru-downloader/internal/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	app        fyne.App
	window     fyne.Window
	config     *config.Config
	downloader *core.Downloader

	// UI
	startStopBtn    *widget.Button
	statusLabel     *widget.Label
	downloaded      binding.String
	speed           binding.String
	downloadedGB    binding.String
	elapsedTime     binding.String
	consoleOutput   *widget.Entry
	consoleMessages []string
}

func NewGUI(cfg *config.Config) *GUI {
	g := &GUI{
		app:             app.New(),
		config:          cfg,
		downloaded:      binding.NewString(),
		speed:           binding.NewString(),
		downloadedGB:    binding.NewString(),
		elapsedTime:     binding.NewString(),
		consoleOutput:   widget.NewMultiLineEntry(),
		consoleMessages: make([]string, 0),
	}

	g.window = g.app.NewWindow("Gobooru Downloader")
	g.window.Resize(fyne.NewSize(600, 400))

	// Initialize UI elements
	g.consoleOutput.Wrapping = fyne.TextWrapWord

	// Initialize values
	g.downloaded.Set("0")
	g.speed.Set("0 KB/s")
	g.downloadedGB.Set("0.00 GB")
	g.elapsedTime.Set("00:00:00")

	// Redirect logger output to "console"
	logger.SetOutput(&consoleWriter{gui: g})

	g.setupUI()
	g.setupMenu()
	g.downloader = core.NewDownloader(cfg)
	return g
}

func (g *GUI) setupMenu() {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Settings", g.showSettings),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() { g.app.Quit() }),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", g.showAbout),
	)

	mainMenu := fyne.NewMainMenu(
		fileMenu,
		helpMenu,
	)
	g.window.SetMainMenu(mainMenu)
}

func (g *GUI) setupUI() {
	// Status area
	g.statusLabel = widget.NewLabel("Ready to start")
	g.statusLabel.Wrapping = fyne.TextWrapWord

	// Start/Stop button
	g.startStopBtn = widget.NewButton("Start Download", g.toggleDownload)
	g.startStopBtn.Importance = widget.HighImportance

	// Stats display
	statsGrid := container.New(layout.NewFormLayout())
	statsGrid.Add(widget.NewLabel("Downloaded:"))
	statsGrid.Add(widget.NewLabelWithData(g.downloaded))

	statsGrid.Add(widget.NewLabel("Speed:"))
	statsGrid.Add(widget.NewLabelWithData(g.speed))

	statsGrid.Add(widget.NewLabel("Downloaded GB:"))
	statsGrid.Add(widget.NewLabelWithData(g.downloadedGB))

	statsGrid.Add(widget.NewLabel("Elapsed Time:"))
	statsGrid.Add(widget.NewLabelWithData(g.elapsedTime))

	// Console output with scroll
	consoleScroll := container.NewScroll(g.consoleOutput)
	consoleScroll.SetMinSize(fyne.NewSize(0, 200))

	// Main layout
	content := container.NewBorder(
		nil,           // Top
		consoleScroll, // Bottom (console)
		nil,           // Left
		nil,           // Right
		container.NewVBox(
			widget.NewSeparator(),
			g.statusLabel,
			container.NewCenter(g.startStopBtn),
			widget.NewSeparator(),
			statsGrid,
			widget.NewSeparator(),
			layout.NewSpacer(),
		),
	)

	g.window.SetContent(content)
}

func (g *GUI) updateUIAfterStop() {
	g.startStopBtn.SetText("Start Download")
	g.startStopBtn.Enable()
	g.statusLabel.SetText("Download stopped")
}

func (g *GUI) updateProgress(ctx context.Context) {
	startTime := time.Now()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if g.downloader == nil || !g.downloader.IsRunning() {
				return
			}

			progress := g.downloader.GetProgress()
			fyne.Do(func() {
				g.downloaded.Set(fmt.Sprintf("%d/%d", progress.Downloaded, progress.Total))
				g.speed.Set(fmt.Sprintf("%.1f KB/s", progress.SpeedKBps))

				// Update GB downloaded
				g.downloadedGB.Set(fmt.Sprintf("%.2f GB", progress.DownloadedGB))

				// Update elapsed time
				elapsed := time.Since(startTime)
				hours := int(elapsed.Hours())
				minutes := int(elapsed.Minutes()) % 60
				seconds := int(elapsed.Seconds()) % 60
				g.elapsedTime.Set(fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds))
			})
		}
	}
}

func (g *GUI) showSettings() {
	settingsWindow := g.app.NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(500, 340))

	// Create form elements for config
	booruURLEntry := widget.NewEntry()
	booruURLEntry.SetText(g.config.BooruURL.String())

	workersEntry := widget.NewEntry()
	workersEntry.SetText(strconv.Itoa(int(g.config.WorkerCount)))

	outputDirEntry := widget.NewEntry()
	outputDirEntry.SetText(g.config.OutputDir)

	proxyEntry := widget.NewEntry()
	proxyEntry.SetText(g.config.ProxyString)

	tagsEntry := widget.NewEntry()
	tagsEntry.SetText(g.config.Tags)

	fromPageEntry := widget.NewEntry()
	fromPageEntry.SetText(strconv.Itoa(int(g.config.FromPage)))

	maxFileSizeEntry := widget.NewEntry()
	maxFileSizeEntry.SetText(strconv.Itoa(int(g.config.MaxFileSize)))

	downloadLimitGBEntry := widget.NewEntry()
	downloadLimitGBEntry.SetText(strconv.Itoa(int(g.config.DownloadLimitGb)))

	maxRetriesEntry := widget.NewEntry()
	maxRetriesEntry.SetText(strconv.Itoa(int(g.config.MaxRetries)))

	noMetadataCheck := widget.NewCheck("No metadata", func(b bool) { g.config.NoMetadata = b })
	noMetadataCheck.Checked = g.config.NoMetadata

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Booru URL", Widget: booruURLEntry},
			{Text: "Worker Count", Widget: workersEntry},
			{Text: "Output Directory", Widget: outputDirEntry},
			{Text: "Max Retries", Widget: maxRetriesEntry},
			{Text: "Proxy connection string", Widget: proxyEntry},
			{Text: "Tags", Widget: tagsEntry},
			{Text: "From page", Widget: fromPageEntry},
			{Text: "Max file size (MB)", Widget: maxFileSizeEntry},
			{Text: "Download limit (GB)", Widget: downloadLimitGBEntry},
			{Text: "No metadata", Widget: noMetadataCheck},
		},
		OnSubmit: func() {
			// Update config
			workerCount, err := strconv.Atoi(workersEntry.Text)
			if err == nil {
				g.config.WorkerCount = uint(workerCount)
			}

			g.config.OutputDir = outputDirEntry.Text

			maxRetries, err := strconv.Atoi(maxFileSizeEntry.Text)
			if err == nil {
				g.config.MaxRetries = uint(maxRetries)
			}

			g.config.ProxyString = proxyEntry.Text

			g.config.Tags = tagsEntry.Text

			fromPage, err := strconv.Atoi(fromPageEntry.Text)
			if err == nil {
				g.config.FromPage = uint(fromPage)
			}

			maxFileSizeMB, err := strconv.ParseFloat(maxFileSizeEntry.Text, 64)
			if err == nil {
				g.config.MaxFileSize = uint(math.Ceil(maxFileSizeMB))
			}

			downloadLimitGB, err := strconv.ParseFloat(downloadLimitGBEntry.Text, 64)
			if err == nil {
				g.config.DownloadLimitGb = downloadLimitGB
			}

			g.config.NoMetadata = noMetadataCheck.Checked

			settingsWindow.Close()
		},
	}

	settingsWindow.SetContent(form)
	settingsWindow.Show()
}

func (g *GUI) showAbout() {
	dialog.ShowInformation(
		"About",
		fmt.Sprintf(`Gobooru Downloader v%s
Copyright (C) 2025  Kasyanov Nikolay Alexeevich (Unbewohnte)
This program comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it
under certain conditions (see COPYING)`, core.VERSION),
		g.window,
	)
}

func (g *GUI) Run() error {
	g.window.ShowAndRun()
	return nil
}

func (g *GUI) Stop() error {
	return g.downloader.Stop()
}
