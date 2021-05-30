# dedupe
- De-duplicates files, not directories, by their sha1 checksum. 
- Always runs dry-run unless the `-clean` flag is explicitly set
- Prints out the total number of duplicate bytes
## Building
- [Install Go](https://golang.org/dl/)
- `cd cmd/dedupe`
- `go build`
## Using
- `cd cmd/dedupe`
- See `./dedupe -h` for options