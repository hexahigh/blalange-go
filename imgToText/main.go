package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"strings"

	libimg "github.com/hexahigh/blalange-go/imgToText/libimg"
)

var (
	filePath   = flag.String("f", "", "File path of image to convert to text")
	color      = flag.Bool("c", false, "Color text")
	width      = flag.Int("w", 0, "Width of image")
	dimensions = flag.String("d", "", "Dimensions of image (w,h)")
	complex    = flag.Bool("complex", false, "Complex character set")
	braille    = flag.Bool("b", false, "Braille output")
	grayscale  = flag.Bool("gs", false, "Grayscale output")
	customMap  = flag.String("m", "", "Custom character set")
)

func init() {
	flag.Parse()
}

func main() {
	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "Please specify a file path with -f")
		os.Exit(1)
	}

	bytes, err := os.OpenFile(*filePath, os.O_RDONLY, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	img, _, err := image.Decode(bytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	imgset, err := libimg.ConvertToAsciiPixels(img, nil, *width, 0, false, false, false, *braille, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var asciiset [][]libimg.AsciiChar

	if *braille {
		asciiset, err = libimg.ConvertToBrailleChars(imgset, false, *color, *grayscale, false, [3]int{255, 255, 255}, 0)
	} else {
		asciiset, err = libimg.ConvertToAsciiChars(imgset, false, *color, *grayscale, *complex, false, *customMap, [3]int{255, 255, 255})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	ascii := flattenAscii(asciiset, *color, false)

	result := strings.Join(ascii, "\n")
	fmt.Println(result)

}

func flattenAscii(asciiSet [][]libimg.AsciiChar, colored, toSaveTxt bool) []string {
	var ascii []string

	for _, line := range asciiSet {
		var tempAscii string

		for _, char := range line {
			if toSaveTxt {
				tempAscii += char.Simple
				continue
			}

			if colored {
				tempAscii += char.OriginalColor
			} else {
				tempAscii += char.Simple
			}
		}

		ascii = append(ascii, tempAscii)
	}

	return ascii
}
