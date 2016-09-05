package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var r = regexp.MustCompile(`GET (\S+)`)

func FindImageParts(fn string, c chan<- ImagePart) {
	file, err := os.Open(fn)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dupMap := make(map[string]bool)
	host := extractHostnameFromFilename(fn)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if URLPath, ok := extractImageURLPath(line); ok && !dupMap[URLPath] {
			dupMap[URLPath] = true

			c <- ImagePart{
				host,
				URLPath,
			}
		}
	}
	close(c)
}

func extractImageURLPath(s string) (URLPath string, isImage bool) {
	matches := r.FindAllStringSubmatch(s, -1)

	isImage = len(matches) > 0 && strings.Contains(matches[0][1], imageURLPrefix)
	if isImage {
		URLPath = matches[0][1]
	}

	return
}

func extractHostnameFromFilename(fn string) string {
	baseName := filepath.Base(fn)

	return "https://" + baseName[strings.Index(baseName, "_")+1:]
}
