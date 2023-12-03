package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

func testProgress() {
	pw := progress.NewWriter()
	pw.SetAutoStop(false)
	pw.SetTrackerLength(34)
	// pw.SetTrackerLength(100)
	pw.SetMessageWidth(2)
	pw.SetNumTrackersExpected(10)
	pw.SetSortBy(progress.SortByMessage)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsDefault
	pw.Style().Options.Separator = ""
	pw.Style().Visibility.ETA = false
	pw.Style().Visibility.ETAOverall = false
	pw.Style().Visibility.Percentage = false
	pw.Style().Visibility.Speed = false
	pw.Style().Visibility.SpeedOverall = false
	pw.Style().Visibility.Time = false
	pw.Style().Visibility.TrackerOverall = false
	pw.Style().Visibility.Value = false
	pw.Style().Visibility.Pinned = false
	pw.Style().Chars.BoxLeft = "|"
	pw.Style().Chars.BoxRight = "🏁"
	pw.Style().Chars.Finished = "-"
	pw.Style().Chars.Finished25 = "-"
	pw.Style().Chars.Finished50 = "-"
	pw.Style().Chars.Finished75 = "-"
	pw.Style().Chars.Unfinished = " "
	// pw.Style().Chars.Indeterminate = progress.IndeterminateIndicatorCycle

	go pw.Render()

	go func() {
		trackers := []*progress.Tracker{}
		for i := 0; i < 10; i++ {
			tracker := progress.Tracker{Message: fmt.Sprintf("%02d", i+1), Total: 100, Units: progress.UnitsDefault, DeferStart: false}
			r := int64(rand.Int()%100 + 1)
			tracker.SetValue(r)
			pw.AppendTracker(&tracker)
			trackers = append(trackers, &tracker)
		}

		go func() {
			for {
				for _, tracker := range trackers {
					if tracker.Value() == 99 {
						tracker.SetValue(0)
					} else {
						tracker.Increment(1)
					}
				}
				time.Sleep(400 * time.Millisecond)
			}
		}()
	}()
}

func main() {

	testProgress()
	newDone := make(chan bool)
	<-newDone
}
