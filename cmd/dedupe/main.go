package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func checksum(path string) (cksum string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func main () {
	entryDir := flag.String("directory", "", "the absolute path of the directory containing the files to de-duplicate")
	extensions := flag.String("extensions", "", "comma separated list of file extensions to filter by")
	preferred := flag.String("preferred", "", "the preferred extension of the file to keep in a set of duplicates")
	clean := flag.Bool("clean", false, "delete duplicate files")
	flag.Parse()

	if *entryDir == "" {
		log.Fatalln("path is required")
	}

	if *extensions == "" {
		log.Fatalln("a comma separated list of extensions are required")
	}

	if *clean {
		fmt.Println("WARNING: Files will begin deletion in ten seconds. CTRL+C to stop.")
		time.Sleep(10 * time.Second)
		fmt.Println("Starting...")
	}

	// get extensions
	*extensions = strings.ReplaceAll(*extensions, " ", "")
	extList := strings.Split(*extensions, ",")

	checksums := make(map[string]map[string]bool)
	fileCounts := make(map[string]int)
	fileSizes := make(map[string]int64)

	evalFile := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		for _, ext := range extList {
			if filepath.Ext(path) == "." + ext {
				sum, err := checksum(path)
				if err != nil {
					return fmt.Errorf("checksum error: %s", err)
				}

				files, exists := checksums[sum]
				if exists {
					files[path] = true
				} else {
					checksums[sum] = make(map[string]bool)
					checksums[sum][path] = true
				}
				fileCounts[sum] += 1

				info, err := d.Info()
				if err != nil {
					return err
				}
				fileSizes[path] = info.Size()

				break
			}
		}

		return nil
	}

	if err := filepath.WalkDir(*entryDir, evalFile); err != nil {
		log.Fatalln(err)
	}

	// remove the file path with the preferred extension. this will ensure the desired original is not marked
	// duplicate for deletion.
	if *preferred != "" {
		for _, files := range checksums {
			var found bool
			// if we find a file with the extension we want, remove it so it doesn't get marked as a duplicate later
			for d := range files {
				if filepath.Ext(d) == "." + *preferred {
					//original := "original: " + d
					//files[original] = true
					delete(files, d)
					found = true
					break
				}
			}

			// if we don't have a file with our preferred extension, remove the first file in the map so it isn't deleted
			if !found {
				for d := range files {
					//original := "original: " + d
					//files[original] = true
					delete(files, d)
					break
				}
			}
		}
	} else {
		// otherwise, delete the first file in the list.
		for _, files := range checksums {
			for d := range files {
				//original := "original: " + d
				//files[original] = true
				delete(files, d)
				break
			}
		}
	}

	var dupeSizeTotal int64
	// print out the duplicate paths, or delete them, now that we have preserved at least one file in each files list.
	for sum, dupeFiles := range checksums {
		if len(dupeFiles) > 0 {
			fmt.Println(sum, " -> ", fileCounts[sum])
			for d := range dupeFiles {
				dupeSizeTotal += fileSizes[d]

				if *clean {
					err := os.Remove(d)
					if err != nil {
						log.Fatalln("file deletion error: ", err)
					}
					fmt.Println("deleted: ", d)
				} else {
					fmt.Println(d)
				}
			}
		}
 	}

 	fmt.Println("total duplicate bytes: ", dupeSizeTotal)
}
