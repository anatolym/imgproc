# imgproc - Image processing tool

[![Go Report Card](https://goreportcard.com/badge/github.com/anatolym/imgproc)](https://goreportcard.com/report/github.com/anatolym/imgproc)

Simple tool for defining top N most prevalent colors of given images.

Supported image formats:
* GIF,
* JPEG,
* PNG.

## Usage

```
$ imgproc -h
Usage of ./imgproc:
  -d int
        number of download workers operating simultaneously (default 10)
  -n int
        count of most prevalent colors included to the results (default 1)
  -r string
        path to result CSV file
  -s string
        path to source file with image URLs
  -w int
        number of image processing workers operating simultaneously (default 'count of CPU - 1')
```