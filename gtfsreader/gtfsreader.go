package gtfsreader

import (
	"encoding/csv"
	"os"
)

type GTFSReader struct {
	reader *csv.Reader
	Headers map[string]int
	file *os.File
}

func NewReader( filename string ) *GTFSReader {
	file, err := os.Open(filename)
        if err != nil {
//                fmt.Println("Error opening stops file")
                return nil
        }
	reader := &GTFSReader{
		reader: csv.NewReader( file ),
		file: file,
	};
	headers, err := reader.reader.Read();
	if err != nil {
		file.Close();
		return nil;
	}
	reader.Headers = make(map[string]int);
	for idx, value := range headers {
		reader.Headers[value] = idx;
	}
	return reader;
}

func (r *GTFSReader) Read() (record []string, err error){
	return r.reader.Read();
}

func (r *GTFSReader) Close() {
	r.file.Close();
}
