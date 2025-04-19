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

package cli

import (
	"Unbewohnte/gobooru-downloader/internal/config"
	"Unbewohnte/gobooru-downloader/internal/core"
	"fmt"
)

type CLI struct {
	config     *config.Config
	downloader *core.Downloader
}

func NewCLI(cfg *config.Config) *CLI {
	return &CLI{
		config:     cfg,
		downloader: core.NewDownloader(cfg),
	}
}

func (cli *CLI) printBanner() {
	fmt.Print(
		` ██████╗  ██████╗ ██████╗  ██████╗  ██████╗ ██████╗ ██╗   ██╗      ██████╗ ██╗    ██╗
██╔════╝ ██╔═══██╗██╔══██╗██╔═══██╗██╔═══██╗██╔══██╗██║   ██║      ██╔══██╗██║    ██║
██║  ███╗██║   ██║██████╔╝██║   ██║██║   ██║██████╔╝██║   ██║█████╗██║  ██║██║ █╗ ██║
██║   ██║██║   ██║██╔══██╗██║   ██║██║   ██║██╔══██╗██║   ██║╚════╝██║  ██║██║███╗██║
╚██████╔╝╚██████╔╝██████╔╝╚██████╔╝╚██████╔╝██║  ██║╚██████╔╝      ██████╔╝╚███╔███╔╝
 ╚═════╝  ╚═════╝ ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝       ╚═════╝  ╚══╝╚══╝ 
` + fmt.Sprintf("%s\n", core.VERSION))
}

func (cli *CLI) Run() error {
	if cli.config.Version {
		cli.config.PrintVersion(core.VERSION)
		return nil
	}

	cli.printBanner()
	return cli.downloader.Run()
}

func (cli *CLI) Stop() error {
	return cli.downloader.Stop()
}
