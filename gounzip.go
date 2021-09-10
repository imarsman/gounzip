package main

import (
	"archive/zip"
	"fmt"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/jwalton/gchalk"
)

// used for colour output
const (
	brightGreen = iota
	brightYellow
	brightBlue
	brightRed
	noColour // Can use to default to no colour output
)

// used to describe a file to be zipped
type fileEntry struct {
	rootPath   string
	parentPath string
	path       string
	isBareFile bool
}

// zipFileEntry represents data stored about a file to be zipped
type zipFileEntry struct {
	name             string
	compressedSize   uint64
	uncompressedSize uint64
	date             string
	time             string
	timestamp        time.Time
}

func hasZipFileEntry(path string, feList *[]zipFileEntry) (found bool, fe zipFileEntry) {
	if len(*feList) == 0 {
		return
	}
	for _, fe = range *feList {
		if path == fe.name {
			found = true
			return
		}
	}
	return
}

// colour get colour output
func colour(colour int, input ...string) (output string) {
	str := fmt.Sprint(strings.Join(input, " "))
	str = strings.Replace(str, "  ", " ", -1)

	// Choose colour for output or none
	switch colour {
	case brightGreen:
		output = gchalk.BrightGreen(str)
	case brightYellow:
		output = gchalk.BrightYellow(str)
	case brightBlue:
		output = gchalk.BrightBlue(str)
	case brightRed:
		output = gchalk.BrightRed(str)
	default:
		output = str
	}

	return
}

// printEntries of a zip file
func printEntries(name string) (err error) {
	zipFileEntries, err := zipFileList(name)
	if err != nil {
		return
	}
	// I am no genius at formatting alignment
	fmt.Printf("%2sCompressed%1sUncompressed%6sDate%7sTime%8sName\n", "", "", "", "", "")
	fmt.Println(strings.Repeat("-", 75))

	var totalCompressed int64 = 0
	var totalUnCompressed int64 = 0
	count := 0
	for _, file := range zipFileEntries {
		fmt.Printf("%12d %11d %12s  %-7s  %-10s\n",
			file.compressedSize,
			file.uncompressedSize,
			file.date,
			file.time,
			file.name,
		)
		totalCompressed += int64(file.compressedSize)
		totalUnCompressed += int64(file.uncompressedSize)
		count++
	}
	fmt.Println(strings.Repeat("-", 75))
	fmt.Printf("%12d%12d%27d\n", totalCompressed, totalUnCompressed, count)
	return
}

// zipFileList get list of files in zipfile
func zipFileList(name string) (entries []zipFileEntry, err error) {
	zf, err := zip.OpenReader(name)
	if err != nil {
		return
	}

	defer zf.Close()

	entries = make([]zipFileEntry, 0, 10)

	for _, file := range zf.File {
		entry := zipFileEntry{}
		entry.name = file.Name
		entry.compressedSize = file.CompressedSize64
		entry.uncompressedSize = file.UncompressedSize64
		dateStr := file.Modified.Format("2006-01-02") // get formatted date
		timeStr := file.Modified.Format("15:04:05")   // get formatted time
		entry.date = dateStr
		entry.time = timeStr
		// likely already UTC but be sure
		entry.timestamp = file.Modified.In(time.UTC)

		entries = append(entries, entry)
	}

	return
}

/*
	The zip utility has a lot of options. It is not know at the time of this
	writing how much of the original utility will be implemented.
*/

// the parameters used by the app
var args struct {
	List    bool `arg:"-l" help:"list entries in zip file" default:"false"`
	Quiet   bool `arg:"-q" help:"suppress normal output"`
	Update  bool `arg:"-u" help:"update directory files and add those not in directory"`
	Freshen bool `arg:"-f" help:"leave existing files in place unless the ones in the archive are newer"`
	// Not currently supported in Go library
	// CompressionLevel uint16   `arg:"-L" derault:"6" help:"compression level (0-9) - defaults to 6" placeholder:"6"`
	Zipfile string `arg:"positional,required" help:"the zip file to extract" placeholder:"zipfile"`
	Dir     string `arg:"-d" placeholder:"DIR" help:"base directory to extract to"`
	// SourceFiles []string `arg:"positional" placeholder:"file"`
}

func main() {
	args.Quiet = false
	arg.MustParse(&args)

}
