package main

import (
	"github.com/360EntSecGroup-Skylar/excelize"
)

// WriteToCSV ...
func WriteToCSV(filename string, data []*SoccerPlayer) error {
	return nil
}

// WriteToXLS ...
func WriteToXLS(filename string, data []*SoccerPlayer) error {
	xls := excelize.NewFile()
	idx := xls.NewSheet("Sheet1")

	return nil
}
