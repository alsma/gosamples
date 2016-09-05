package logpuzzle

import (
	"bytes"
	"github.com/alsma/gosamples/logpuzzle/parser"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"

	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path/filepath"
)

type partsStream <-chan parser.ImagePart

func CompilePuzzle(p partsStream) <-chan string {
	c := make(chan string, 1)

	go func() {
		var wg sync.WaitGroup
		downloaded := make(chan *downloadResult)

		for part := range p {
			// increment latch count
			wg.Add(1)

			// download each image in own goroutine
			go func() {
				downloaded <- download(part)
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

		//p := image.Point{0,0}

		for i := range downloaded {
			partIMG, _, err := image.Decode(bytes.NewReader(i.data))
			if err != nil {
				panic(err)
			}

			URLIndexedMap[i.URL] = partIMG
			orderedURLs = append(orderedURLs, i.URL)

			//p.Add()
			upperBound := partIMG.Bounds().Max
			sumWidth = sumWidth + upperBound.X
			if upperBound.Y > maxHeight {
				maxHeight = upperBound.Y
			}
		}

		//fmt.Printf("%v", orderedURLs)
		//sort.Sort(UrlsByPart(orderedURLs))
		sort.Strings(orderedURLs)
		//fmt.Printf("/n/n%v", orderedURLs)

		//resRect := image.Rect(0, 0, sumWidth, maxHeight)
		resRect := image.Rect(0, 0, 800, 600)
		resRGBA := image.NewRGBA(resRect)

		curPosition := 0
		for _, i := range orderedURLs {
			IMG := URLIndexedMap[i]

			draw.Draw(resRGBA, resRGBA.Bounds(), IMG, image.Point{curPosition, 0}, draw.Src)
			curPosition = curPosition - IMG.Bounds().Max.X
			fmt.Printf("%v", curPosition)
		}

		// TODO think how to guess extension
		out, err := os.Create(fmt.Sprintf("./output%s", filepath.Ext(orderedURLs[0])))
		if err != nil {
			panic(err)
		}
		defer out.Close()

		png.Encode(out, resRGBA)

		c <- out.Name()
	}()

	return c
}

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

func extractValuablePartOfUrl(s string) string {
	return s
}
