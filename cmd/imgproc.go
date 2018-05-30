package main

import (
	"flag"
	"log"

	"github.com/anatolym/imgproc"
)

var srcFile = flag.String("s", "", "path to source file")
var resultFile = flag.String("r", "", "path to result CSV file")
var resultN = flag.Int("n", 1, "count of most prevalent colors included to the results")

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

	src, err := imgproc.NewFileSrc(*srcFile)
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()
	csv, err := imgproc.NewCsvResult(*resultFile)
	if err != nil {
		log.Fatal(err)
	}
	for {
		imgItem, err := src.Next()
		if err != nil {
			log.Printf("Image source returns an error: %s", err)
			continue
		} else if imgItem == nil {
			log.Println("Image source returns 'nil' - end of processing")
			break
		}

		results, err := imgproc.Analyze(imgItem.Data, *resultN)
		if err != nil {
			log.Printf("Error during processing image %s: %s", imgItem.Name, err)
		}
		if err := csv.Add(imgItem.Name, results); err != nil {
			log.Printf("Error during saving results for image %s: %s", imgItem.Name, err)
		}
	}
}
