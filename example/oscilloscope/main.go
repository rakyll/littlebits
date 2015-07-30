// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"time"

	ui "github.com/gizak/termui"
	"github.com/rakyll/littlebits"
)

const bufSize = 1024

func main() {
	r, err := littlebits.NewReader("", bufSize)
	if err != nil {
		panic(err)
	}

	err = ui.Init()
	if err != nil {
		panic(err)
	}

	defer func() {
		ui.Close()
		r.Close()
	}()

	lc := ui.NewLineChart()

	vals := make([]float64, bufSize/2)
	go func() {
		for {
			w := ui.TermWidth()
			var buf = make([]byte, w*2)
			n, err := r.Read(buf)
			if err != nil {
				continue
			}

			// Reader returns values for two channels, we can avoid one.
			for i := 0; i < n/2; i++ {
				vals[i] = float64(buf[i*2])
			}
			ui.Render(lc)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	lc.PaddingLeft = 2
	lc.Mode = "dot"
	lc.Data = vals[:]
	lc.HasBorder = false
	lc.Width = ui.TermWidth()
	lc.Height = ui.TermHeight()
	lc.LineColor = ui.ColorCyan | ui.AttrReverse

	ui.Render(lc)
	<-ui.EventCh()
}
