package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/anatolym/imgproc"
)

const defaultLoadNum = 10

var srcFile = flag.String("s", "", "path to source file with image URLs")
var resultFile = flag.String("r", "", "path to result CSV file")
var resultN = flag.Int("n", 1, "count of most prevalent colors included to the results")
var loadNum = flag.Int("d", defaultLoadNum, "number of download workers operating simultaneously")
var workersNum = flag.Int("w", 0, "number of image processing workers operating simultaneously (default 'count of CPU - 1')")

func main() {
	flag.Parse()
	if *srcFile == "" {
		log.Fatal("source file is not specified, see option -h")
	}
	if *resultFile == "" {
		log.Fatal("result CSV file is not specified, see option -h")
	}
	if *resultN <= 0 {
		log.Fatal("count of most prevalent colors (-n) should be above 0")
	}
	if *loadNum <= 0 {
		// This option allows to run download in parallel.
		// Assuming download time is greater than processing time.
		*loadNum = defaultLoadNum
		log.Printf("Number of download workers operating simultaneously: %d", *loadNum)
	}
	if *workersNum <= 0 {
		*workersNum = runtime.NumCPU() - 1
		if *workersNum <= 0 {
			*workersNum = 1
		}
		log.Printf("Number of image processing workers operating simultaneously: %d", *workersNum)
	}

	src, err := imgproc.NewFileSrc(*srcFile, *loadNum)
	if err != nil {
		log.Fatal(err)
	}
	csv, err := imgproc.NewCsvResult(*resultFile, *resultN)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Start processing")
	done := make(chan struct{})
	defer close(done)
	// Getting results and saving it to CSV.
	for res := range imgproc.ProcessItems(done, src.GetImgItemCh(done), *resultN, *workersNum) {
		if err := csv.Add(&res); err != nil {
			log.Printf("Error during saving results for image %s: %s", res.Name, err)
		}
	}

	log.Println("Stop processing")
}
