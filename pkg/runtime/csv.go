package runtime

import (
	"encoding/csv"
	"io"
)

type CSVReader struct {
	r *csv.Reader
}

func NewCSVReader(r io.Reader) *CSVReader {
	return &CSVReader{r: csv.NewReader(r)}
}

func (cr *CSVReader) ReadAll() ([][]string, error) {
	return cr.r.ReadAll()
}

type CSVWriter struct {
	w *csv.Writer
}

func NewCSVWriter(w io.Writer) *CSVWriter {
	return &CSVWriter{w: csv.NewWriter(w)}
}

func (cw *CSVWriter) WriteAll(records [][]string) error {
	return cw.w.WriteAll(records)
}
