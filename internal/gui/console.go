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
	"fyne.io/fyne/v2"
)

type consoleWriter struct {
	gui *GUI
}

func (cw *consoleWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	// Remove trailing newline if present
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}

	cw.gui.appendConsoleMessage(msg)
	return len(p), nil
}

func (g *GUI) appendConsoleMessage(msg string) {
	fyne.Do(func() {
		// Keep last 100 messages
		g.consoleMessages = append(g.consoleMessages, msg)
		if len(g.consoleMessages) > 100 {
			g.consoleMessages = g.consoleMessages[1:]
		}

		// Join all messages with newlines
		fullText := ""
		for _, m := range g.consoleMessages {
			fullText += m + "\n"
		}
		g.consoleOutput.SetText(fullText)
		g.consoleOutput.CursorRow = len(g.consoleMessages) - 1 // Auto-scroll
		g.consoleOutput.CursorColumn = 0
	})
}
