package main

import (
	"flag"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	html "github.com/hexahigh/blalange-go/imgToText/html"
	libimg "github.com/hexahigh/blalange-go/imgToText/libimg"
)

var (
	filePath         = flag.String("f", "", "File path of image to convert to text")
	color            = flag.Bool("c", false, "Color text")
	width            = flag.Int("w", 0, "Width of image")
	dimensions       = flag.String("d", "", "Dimensions of image (w,h)")
	complex          = flag.Bool("complex", false, "Complex character set")
	braille          = flag.Bool("b", false, "Braille output")
	grayscale        = flag.Bool("gs", false, "Grayscale output")
	customMap        = flag.String("m", "", "Custom character set")
	http_server      = flag.Bool("http", false, "run a http server")
	http_server_host = flag.String("http:host", ":8080", "http server host")
)

func init() {
	flag.Parse()
}

func main() {
	if *http_server {
		httpServer()
		os.Exit(0)
	}

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

func httpServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get parameters from URL
		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, "Failed to parse query parameters", http.StatusBadRequest)
			return
		}

		type Params struct {
			Url       string
			Width     int
			Complex   bool
			Braille   bool
			Color     bool
			customMap string
			Output    string
		}

		var params Params

		params.Url = queryParams.Get("u")
		params.Width, _ = strconv.Atoi(queryParams.Get("w"))
		params.Complex = queryParams.Has("complex")
		params.Braille = queryParams.Has("b")
		params.Color = queryParams.Has("c")
		params.Output = queryParams.Get("o")
		params.customMap = queryParams.Get("m")

		fmt.Println(params)

		if params.Url == "" {
			http.Error(w, "Missing 'u' parameter", http.StatusBadRequest)
			return
		}

		// Download file
		resp, err := http.Get(params.Url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Decode image
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		imgset, err := libimg.ConvertToAsciiPixels(img, nil, params.Width, 0, false, false, false, params.Braille, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var asciiset [][]libimg.AsciiChar

		if params.Braille {
			asciiset, err = libimg.ConvertToBrailleChars(imgset, false, params.Color, !params.Color, false, [3]int{255, 255, 255}, 0)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			asciiset, err = libimg.ConvertToAsciiChars(imgset, false, params.Color, !params.Color, params.Complex, false, params.customMap, [3]int{255, 255, 255})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		ascii := flattenAscii(asciiset, params.Color, false)

		result := strings.Join(ascii, "\n")

		// Add the params to the headers
		w.Header().Add("x-params", fmt.Sprintf("%+v", params))

		switch params.Output {
		case "html":
			w.Header().Set("Content-Type", "text/html")
			var result []string
			// Turn each line from ascii into html
			for _, line := range ascii {
				result = append(result, string(html.ConvertToHTML([]byte(line))))
			}
			// Merge them using newline
			result2 := strings.Join(result, "<br>")
			w.Write([]byte(result2))
			return
		default:
			w.Write([]byte(result))
			return
		}

	})
	http.ListenAndServe(*http_server_host, nil)
}
