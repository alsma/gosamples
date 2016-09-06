package main

import (
	"flag"
	"fmt"
	"github.com/alsma/gosamples/logpuzzle"
	"github.com/alsma/gosamples/logpuzzle/parser"
	"github.com/alsma/gosamples/logpuzzle/utils"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
)

const (
	defaultDataDirectoryRelativePath             = "data/logpuzzle"
	targetDirName                                = "target"
	targetDirPerm                    os.FileMode = 0700
	prefferableOutputFormat                      = "png"
)

func main() {
	var (
		dataDir   string
		targetDir string
	)

	flag.StringVar(&dataDir, "data-dir", fmt.Sprintf("./%s/source", defaultDataDirectoryRelativePath), "Path to directory containing log files")
	flag.StringVar(&targetDir, "target-dir", fmt.Sprintf("./%s/%s", defaultDataDirectoryRelativePath, targetDirName), "Path to directory where result should be placed in")

	flag.Parse()

	utils.CheckReadableDirectoryExists(dataDir)
	utils.EnsureDirectoryExists(targetDir, targetDirPerm)

	var wg sync.WaitGroup

	files, _ := ioutil.ReadDir(dataDir)
	for _, f := range files {
		if !f.Mode().IsRegular() {
			continue
		}

		c := make(chan parser.ImagePart)
		go parser.FindImageParts(fmt.Sprintf("%s/%s", dataDir, f.Name()), c)
		res := logpuzzle.CompilePuzzle(c)

		wg.Add(1)

		go func(name string) {
			i := <-res

			r := regexp.MustCompile("[^a-z]")
			name = r.ReplaceAllLiteralString(name, "_")
			err := logpuzzle.SaveImage(i, fmt.Sprintf("%s/%s.%s", targetDir, name, prefferableOutputFormat))

			if err != nil {
				panic(err)
			}

			wg.Done()
		}(f.Name())
	}

	wg.Wait()
}
