package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/api/sheets/v4"
)

type spreadsheetWriter struct {
	srv          *sheets.Service
	min_rows     int
	max_rows     int
	max_cols     rune
	batch_length int // Write to spreadsheet every time we get to this number of rows
	values       [][]string
	id           string // Id is the identifier in the sheet's URL
	sheet        string // Sheet is the name of the sheet we're writing to
}

func NewSpreadsheetWriter(ctx context.Context, spreadsheetID string, spreadsheetSheet string) (*spreadsheetWriter, error) {
	log.Printf("Getting new spreadsheetWriter")
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting NewSpreadsheetWriter: %v", err)
	}

	// Start at row 2 to leave the header row in place
	sw := spreadsheetWriter{
		min_rows:     2,
		max_rows:     2,
		max_cols:     'A',
		batch_length: 10,
		id:           spreadsheetID,
		sheet:        spreadsheetSheet,
		srv:          srv,
	}

	// Clear the current contents, from second row on. This should leave the formatting intact
	clearRequest := sheets.BatchClearValuesRequest{
		Ranges: []string{
			fmt.Sprintf("%s!A2:Z500", sw.sheet),
		},
	}
	_, err = srv.Spreadsheets.Values.BatchClear(sw.id, &clearRequest).Do()
	if err != nil {
		log.Printf("clearing spreadsheet values: %v", err)
	}

	// Get the sheet ID
	var sheetID int64
	resp, err := sw.srv.Spreadsheets.Get(sw.id).Do()
	if err != nil {
		log.Printf("getting spreadsheet data: %v", err)
	}

	for _, s := range resp.Sheets {
		log.Printf("Sheet name %s has id %d", s.Properties.Title, s.Properties.SheetId)
		if s.Properties.Title == sw.sheet {
			sheetID = s.Properties.SheetId
		}
	}

	// Add a note in cell A1 of this sheet with the current date
	updateCellsRequest := &sheets.UpdateCellsRequest{
		Range: &sheets.GridRange{
			SheetId:          sheetID,
			StartRowIndex:    0,
			StartColumnIndex: 0,
			EndRowIndex:      1,
			EndColumnIndex:   1,
		},
		Fields: "*",
		Rows: []*sheets.RowData{{
			Values: []*sheets.CellData{{
				Note: fmt.Sprintf("Last updated: %s", time.Now().Format("2006-January-02")),
			}},
		}},
	}
	requestBody := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			UpdateCells: updateCellsRequest,
		}},
	}
	_, err = srv.Spreadsheets.BatchUpdate(sw.id, requestBody).Do()
	if err != nil {
		log.Printf("adding spreadsheet note: %v", err)
	}

	return &sw, nil
}

func (sw spreadsheetWriter) Write(p []byte) (n int, err error) {
	log.Printf("spreadsheet writing no-op")
	return len(p), nil
}

func (sw *spreadsheetWriter) WriteRow(record []string) error {
	log.Printf("Appending row to spreadsheet for rider %s, length %d", record[0], len(record))
	sw.values = append(sw.values, record)
	sw.max_rows += 1

	// Ideally we'd work this out based on the data
	sw.max_cols = 'Z'
	log.Printf("Spreadsheet data has %d rows", len(sw.values))

	if len(sw.values) >= sw.batch_length {
		log.Printf("Flush this data")
		sw.Flush()
	}

	return nil
}

func (sw *spreadsheetWriter) Flush() {
	// Start at row 2 to leave the header row intact
	rangeData := fmt.Sprintf("%s!A%d:%c%d", sw.sheet, sw.min_rows, sw.max_cols, sw.max_rows)
	log.Printf("Writing data to spreadsheet range %s, length %d", rangeData, len(sw.values))
	values := make([][]interface{}, len(sw.values))
	for i, row := range sw.values {
		v := make([]interface{}, len(row))
		for j, col := range row {
			v[j] = col
		}
		values[i] = v
	}

	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
	}
	rb.Data = append(rb.Data, &sheets.ValueRange{
		Range:  rangeData,
		Values: values,
	})
	_, err := sw.srv.Spreadsheets.Values.BatchUpdate(sw.id, rb).Do()
	if err != nil {
		log.Printf("writing to spreadsheet: %v", err)
	}

	// Update where we will write to next time, and reset the values
	sw.min_rows = sw.max_rows
	sw.max_rows = sw.min_rows
	sw.values = nil
	return
}

func (sw *spreadsheetWriter) Close() error {
	return nil
}

type rowWriter interface {
	WriteRow(record []string) error
	Flush()
}

func NewRowWriter(w io.Writer) rowWriter {
	sw, ok := w.(*spreadsheetWriter)
	if ok {
		log.Printf("This is a spreadsheetWriter")
		return sw
	}

	log.Printf("This is a csv.Writer")
	m := &myCSV{
		csv.NewWriter(w),
	}

	return m
}

type myCSV struct {
	*csv.Writer
}

func (m *myCSV) WriteRow(record []string) error {
	return m.Writer.Write(record)
}

// func (m *myCSV) Flush() {

// }

// func (r *rowWriter) Write(record []string) error {
// 	log.Printf("Writing row %v", record)
// 	if r.c != nil {
// 		log.Printf("using csv.WRite")
// 		return r.c.Write(record)
// 	}

// 	log.Printf("using sw.WriteRow")
// 	return r.sw.WriteRow(record)
// }

// func (r *rowWriter) Flush() {
// 	if r.c != nil {
// 		log.Printf("using csv.Flush")
// 		r.c.Flush()
// 	}

// 	log.Printf("using sw.Flush")
// 	r.sw.Flush()
// }
