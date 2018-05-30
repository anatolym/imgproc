package imgproc

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// Sourcer is the interface that groups the basic Next and Close methods of an image data source.
type Sourcer interface {
	// Next allows to iterate over the image items of a Sourcer.
	// It should return `nil, nil` at the end of processing list.
	Next() (*ImgItem, error)
	// Close wraps up opened resources.
	Close()
}

// ImgItem represents an image being processed.
type ImgItem struct {
	// Name is the image name, e.g. image filename or URL.
	Name string
	// Data is the image content.
	Data []byte
}

// FileSrc is a Sourcer structure which uses text files as a source of images.
// Each line of a file is image URL.
type FileSrc struct {
	FileName string
	file     *os.File
	reader   *bufio.Reader
}

// NewFileSrc returns new FileSrc object which implements Sourcer interface.
func NewFileSrc(fileName string) (*FileSrc, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("File %s not exists", fileName)
	}
	src := FileSrc{FileName: fileName}
	var err error
	src.file, err = os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Error opening file %s: %s", fileName, err)
	}
	src.reader = bufio.NewReader(src.file)
	return &src, nil
}

// Next returns next ImgItem for processing or nil.
func (src *FileSrc) Next() (*ImgItem, error) {
	line, _, err := src.reader.ReadLine()
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("Cannot read next line: %s", err)
	}
	resp, err := http.Get(string(line))
	if err != nil {
		return nil, fmt.Errorf("Cannot get content of URL: %s", string(line))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Cannot read response body content: %s", err)
	}
	return &ImgItem{string(line), body}, nil
}

// Close wraps up opened resources.
func (src *FileSrc) Close() {
	src.file.Close()
}

// CsvResult represents storage of analysis results.
type CsvResult struct {
	FileName string
}

// NewCsvResult creates new CsvResult.
func NewCsvResult(fileName string) (*CsvResult, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return nil, fmt.Errorf("os.Create returns an error: %s", err)
	}
	file.Close()
	return &CsvResult{fileName}, nil
}

// Add adds new line to CSV result file.
func (r *CsvResult) Add(url string, colors ColorList) error {
	line := url + ","
	for _, color := range colors {
		line += "#" + color.Hex + ","
	}
	return r.append(line)
}

func (r *CsvResult) append(line string) error {
	f, err := os.OpenFile(r.FileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("os.OpenFile returns an error: %s", err)
	}
	if _, err := f.Write([]byte(line + "\n")); err != nil {
		return fmt.Errorf("file.Write returns an error: %s", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("file.Close returns an error: %s", err)
	}
	return nil
}
