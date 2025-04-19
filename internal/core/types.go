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

import "Unbewohnte/gobooru-downloader/internal/booru"

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
