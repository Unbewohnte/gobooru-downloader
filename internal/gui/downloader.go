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
	"Unbewohnte/gobooru-downloader/internal/core"
	"Unbewohnte/gobooru-downloader/internal/logger"
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func (g *GUI) toggleDownload() {
	if g.downloader.IsRunning() {
		g.stopDownload()
	} else {
		g.startDownload()
	}
}

func (g *GUI) startDownload() {
	g.appendConsoleMessage("Starting download...")

	// Update UI first
	fyne.Do(func() {
		g.downloadedGB.Set("0.00 GB")
		g.elapsedTime.Set("00:00:00")
		g.startStopBtn.SetText("Stop Download")
		g.statusLabel.SetText("Downloading...")
		g.startStopBtn.Disable()
	})

	// Run download in background
	go func() {
		// Re-Apply config and create new downloader
		g.config.Apply()
		dl := core.NewDownloader(g.config)
		g.downloader = dl

		// Create communication channels
		done := make(chan struct{})
		errChan := make(chan error, 1)

		// Run download in separate goroutine
		go func() {
			defer close(done)
			errChan <- dl.Run()
		}()

		// Enable button after preparation
		fyne.Do(func() {
			g.startStopBtn.Enable()
		})

		// Wait for completion
		select {
		case err := <-errChan:
			fyne.Do(func() {
				if err != nil {
					logger.Error("Download failed: %v", err)
					dialog.ShowError(err, g.window)
				}
				g.updateUIAfterStop()
			})
		case <-done:
			fyne.Do(g.updateUIAfterStop)
		}
	}()

	// Start non-blocking progress updates
	go g.updateProgress(context.TODO())
}

func (g *GUI) stopDownload() {
	g.appendConsoleMessage("Stopping download...")

	fyne.Do(func() {
		g.startStopBtn.Disable()
		g.statusLabel.SetText("Stopping...")
	})

	go func() {
		err := g.downloader.Stop()
		fyne.Do(func() {
			if err != nil {
				logger.Error("Stop failed: %v", err)
				dialog.ShowError(err, g.window)
			}
			g.updateUIAfterStop()
		})
	}()
}
