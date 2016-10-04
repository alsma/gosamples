package compiler

import (
	"bytes"
	"errors"
	"github.com/alsma/gosamples/logpuzzle/parser"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type partsStream <-chan parser.ImagePart

func CompilePuzzle(p partsStream) <-chan *image.RGBA {
	c := make(chan *image.RGBA, 1)

	go func() {
		var wg sync.WaitGroup
		downloaded := make(chan *downloadResult)

		for part := range p {
			_p := part
			// increment latch count
			wg.Add(1)

			// download each image in own goroutine
			go func() {
				downloaded <- download(_p)
				wg.Done()
			}()
		}

		go func() {
			wg.Wait()
			close(downloaded)
		}()

		URLIndexedMap := make(map[string]image.Image)
		orderedURLs := make([]string, 0)
		sumWidth, maxHeight := 0, 0

		for i := range downloaded {
			partIMG, _, err := image.Decode(bytes.NewReader(i.data))
			if err != nil {
				panic(err)
			}

			URLIndexedMap[i.URL] = partIMG
			orderedURLs = append(orderedURLs, i.URL)

			upperBound := partIMG.Bounds().Max
			sumWidth = sumWidth + upperBound.X
			if upperBound.Y > maxHeight {
				maxHeight = upperBound.Y
			}
		}

		sort.Sort(UrlsByPart(orderedURLs))

		resRect := image.Rect(0, 0, sumWidth, maxHeight)
		resRGBA := image.NewRGBA(resRect)

		curPosition := 0
		for _, i := range orderedURLs {
			IMG := URLIndexedMap[i]

			draw.Draw(resRGBA, resRGBA.Bounds(), IMG, image.Point{curPosition, 0}, draw.Src)
			curPosition = curPosition - IMG.Bounds().Max.X
		}

		c <- resRGBA
	}()

	return c
}

func SaveImage(i *image.RGBA, sp string) error {
	out, err := os.Create(sp)
	if err != nil {
		return err
	}
	defer out.Close()

	ext := filepath.Ext(sp)
	switch ext[1:] {
	case "png":
		png.Encode(out, i)
	case "jpeg", "jpg":
		jpeg.Encode(out, i, nil)
	case "gif":
		gif.Encode(out, i, nil)
	default:
		return errors.New("Known formats are: png, jpeg, gif")
	}

	return nil
}

// synchronously download image part
func download(ip parser.ImagePart) *downloadResult {
	URL := ip.String()

	resp, err := http.Get(URL)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Trouble reading reesponse body!")
	}

	return &downloadResult{URL, contents}
}

type downloadResult struct {
	URL  string
	data []byte
}

type UrlsByPart []string

func (urls UrlsByPart) Len() int {
	return len(urls)
}

func (urls UrlsByPart) Less(i, j int) bool {
	url1, url2 := extractValuablePartOfUrl(urls[i]), extractValuablePartOfUrl(urls[j])

	return url1 < url2
}

func (urls UrlsByPart) Swap(i, j int) {
	urls[i], urls[j] = urls[j], urls[i]
}

// extract valuable part of url for sorting
// basically if base filename contains less then 2 dashes use whole name
func extractValuablePartOfUrl(s string) string {
	name := filepath.Base(s)

	if strings.Count(name, "-") < 2 {
		return name
	}

	return s[strings.LastIndexAny(s, "-"):]
}
