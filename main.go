package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"strconv"

	"github.com/gobwas/glob"
)

var (
	// TODO: improve descriptions
	nameFlag = flag.String("name", "", "name of the file to look for")
	sizeFlag = flag.String("size", "", "size of the file to look for")
	minDepthFlag = flag.String("mindepth", "", "minimum depth")
	maxDepthFlag = flag.String("maxdepth", "", "minimum depth")

)

var (
	nWalked           = 0
	nArchives         = 0
	nInnerFiles       = 0
	totalSize   int64 = 0
	nFound            = 0
)

const dirSep = "/"

func main() {
	flag.Parse()

	fmt.Fprintln(os.Stderr, "######### Looking for name: ", *nameFlag)
	fmt.Fprintln(os.Stderr, "######### Looking for size: ", *sizeFlag)
	fmt.Fprintln(os.Stderr, "######### Looking for min depth: ", *minDepthFlag)
	fmt.Fprintln(os.Stderr, "######### Looking for max depth: ", *maxDepthFlag)
	t0 := time.Now()

	// TODO: cleanup, redundant and not elegant
	matchAll := func(f *zip.File, archiveFilepath string) bool { return true }

	// TODO: optimize for non-glob ?
	matchName := matchAll
	if nameFlag != nil && *nameFlag != "" {
		g := glob.MustCompile(*nameFlag)

		matchName = func(f *zip.File, archiveFilepath string) bool {
			return g.Match(f.Name)
		}
	}

	// TODO: add support for +1M notation
	matchSize := matchAll
	if sizeFlag != nil && *sizeFlag != "" {
		size, err := strconv.Atoi(*sizeFlag)
		if err != nil {
			log.Fatal(err)  // FIXME
		}
		size64 := uint64(size)

		matchSize = func(f *zip.File, archiveFilepath string) bool {
			return f.UncompressedSize64 > size64
		}
	}

	matchDepth := matchAll
	if *minDepthFlag != "" || *maxDepthFlag != "" {
		minDepth := 0
		maxDepth := 0xFFFF
		if val, err := strconv.Atoi(*minDepthFlag); err == nil {
			minDepth = val
		}
		if val, err := strconv.Atoi(*maxDepthFlag); err == nil {
			maxDepth = val
		}
		fmt.Fprintf(os.Stderr, "Min depth: %v  max depth: %v\n", minDepth, maxDepth)

		// TODO : check corner cases here ( / , a, a/, a/b, ...
		matchDepth = func(f *zip.File, archiveFilePath string) bool {
			depth := 1 + strings.Count(f.Name, dirSep) + strings.Count(archiveFilePath, dirSep)
			//fmt.Fprintf(os.Stderr, "######### Depth: %d %s/%s\n", depth, archiveFilePath, f.Name)
			return depth >= minDepth && depth <= maxDepth
		}
	}

	findFunc := func(f *zip.File, p string) bool {
		return matchName(f, p) || (false && matchSize(f, p) && matchDepth(f, p))
	}

	filepath.Walk(".",
		func(fullpath string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			nWalked++
			//fmt.Println(total, archives, "Ext:", filepath.Ext(info.Name()), fullpath)

			if filepath.Ext(fullpath) == ".zip" {
				nArchives++

				// Open a zip archive for reading.
				r, err := zip.OpenReader(fullpath)
				if err != nil {
					log.Fatal(err) // FIXME
				}
				defer r.Close()

				totalSize += info.Size()

				// Iterate through the files in the archive,
				// printing some of their contents.
				for _, f := range r.File {
					nInnerFiles++
					//fmt.Printf("Contents of %s:\n", f.Name)

					if findFunc(f, fullpath) {
						nFound++
						fmt.Printf("%s : %s\n", fullpath, f.Name)
					}
				}
			}
			return nil

		})

	t1 := time.Now()

	fmt.Fprintf(os.Stderr, "\nScanned %d archives, %d files, %.1f MB in %.2f sec\n",
		nArchives, nInnerFiles, float64(totalSize/1024.0/1024.0), t1.Sub(t0).Seconds())
	fmt.Fprintf(os.Stderr, "Found %d matches\n", nFound)
}
