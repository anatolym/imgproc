package imgproc

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

// Sourcer is the interface that groups the basic Next and Close methods of an image data source.
type Sourcer interface {
	// GetImgItemCh returns a ImgItemCh channel of pointers to ImgItem objects to be processed.
	GetImgItemCh() ImgItemCh
}

// ImgItem represents an image being processed.
type ImgItem struct {
	// Name is the image name, e.g. image filename or URL.
	Name string
	// Data is the image content.
	Data []byte
}

// ImgItemCh represents channel of pointers to ImgItem.
type ImgItemCh chan *ImgItem

// FileSrc is a Sourcer structure which uses text files as a source of images.
// Each line of a file is image URL.
type FileSrc struct {
	FileName string
	loadNum  int
}

// NewFileSrc returns new FileSrc object which implements Sourcer interface.
func NewFileSrc(fileName string, loadNum int) (*FileSrc, error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("File %s not exists", fileName)
	}
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("Error opening file %s: %s", fileName, err)
	}
	file.Close()
	src := FileSrc{FileName: fileName, loadNum: loadNum}
	return &src, nil
}

// GetImgItemCh returns a ImgItemCh channel of pointers to ImgItem objects to be processed.
func (src *FileSrc) GetImgItemCh() ImgItemCh {
	imgItemCh := make(ImgItemCh, src.loadNum)
	go func() {
		defer close(imgItemCh) // Closing channel at the end of iterating over file lines.
		file, err := os.Open(src.FileName)
		if err != nil {
			log.Printf("Error opening file %s: %s", src.FileName, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		sem := make(chan struct{}, src.loadNum)
		var wg sync.WaitGroup
		for scanner.Scan() {
			// Downloading images in "loadNum" number of goroutines.
			sem <- struct{}{}
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				defer func() { <-sem }()
				resp, err := http.Get(url)
				if err != nil {
					log.Printf("Cannot get content of URL: %s", url)
					return
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Cannot read response body content for url '%s': %s", url, err)
					return
				}
				imgItemCh <- &ImgItem{url, body}

			}(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading file %s: %s", src.FileName, err)
		}
		// Waiting for processing of all line in goroutines.
		wg.Wait()
	}()
	return imgItemCh
}

// ProcessItems processes ImgItem objects coming from ImgItemCh channel in "workersNum"
// goroutines and returns ResultItemCh channel of processed results.
func ProcessItems(imgItemCh ImgItemCh, resultN, workersNum int) ResultItemCh {
	resultItemCh := make(ResultItemCh)
	go func() {
		defer close(resultItemCh)
		sem := make(chan struct{}, workersNum)
		var wg sync.WaitGroup
		for imgItem := range imgItemCh {
			sem <- struct{}{}
			wg.Add(1)
			go func(imgItem *ImgItem) {
				defer wg.Done()
				defer func() { <-sem }()
				results, err := Analyze(imgItem.Data, resultN)
				if err != nil {
					log.Printf("Error during processing image %s: %s", imgItem.Name, err)
					return
				}
				resultItemCh <- &ResultItem{imgItem.Name, results}
			}(imgItem)
		}
		wg.Wait()
	}()
	return resultItemCh
}

// CsvResult represents storage of analysis results.
type CsvResult struct {
	FileName string
	colNum   int
}

// ResultItem represents processing result to be saved in CSV file.
type ResultItem struct {
	Name    string
	Results ColorList
}

// ResultItemCh represents channel of pointers to ResultItem.
type ResultItemCh chan *ResultItem

// NewCsvResult creates new CsvResult.
func NewCsvResult(fileName string, colNum int) (*CsvResult, error) {
	if colNum < 1 {
		return nil, fmt.Errorf("Column number should be above 0")
	}
	file, err := os.Create(fileName)
	if err != nil {
		return nil, fmt.Errorf("os.Create returns an error: %s", err)
	}
	file.Close()
	return &CsvResult{fileName, colNum}, nil
}

// Add adds new line to CSV result file.
func (r *CsvResult) Add(res *ResultItem) error {
	line := res.Name + ","
	for i, color := range res.Results {
		if i >= r.colNum {
			break
		}
		line += "#" + color.Hex + ","
	}
	// Adding emty columns in case the size of ColorList is less than the numeber of result columns.
	for i := 0; i < r.colNum-len(res.Results); i++ {
		line += ","
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
