package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	filename   string
	parentPath string
}

// zipFileEntry represents data stored about a file to be zipped
type zipFileEntry struct {
	path             string
	compressedSize   uint64
	uncompressedSize uint64
	date             string
	time             string
	timestamp        time.Time
}

// hasFileEntry check for duplicate absolute paths. Files could be put in more than
// once since zip allows multiple dir/path args.
func hasFileEntry(check fileEntry, feList *[]fileEntry) (found bool) {
	for _, fe := range *feList {
		if check.fullPath() == fe.fullPath() {
			return true
		}
	}
	return
}

// does file exist at destination path
func (fe *fileEntry) existsLocally() (exists bool) {
	if _, err := os.Stat(fe.fullPath()); os.IsNotExist(err) {
		return
	}
	exists = true

	return
}

// fullPath get full path for an entry
func (fe *fileEntry) writeToFile(data []byte) (err error) {
	os.MkdirAll(fe.fullDir(), os.ModePerm)

	file, err := os.Create(fe.fullPath())
	if err != nil {
		return
	}
	byteReader := bytes.NewReader(data)
	reader := bufio.NewReader(byteReader)
	_, err = io.Copy(file, reader)
	if err != nil {
		return
	}

	return
}

// fullPath get full path for an entry
func (fe *fileEntry) fullDir() (fullPath string) {
	fullPath = filepath.Dir(fe.fullPath())
	return
}

// fullPath get full path for an entry
func (fe *fileEntry) fullPath() (fullPath string) {
	fullPath = filepath.Join(args.Dir, fe.parentPath, fe.filename)
	return
}

// archivePath get path for archive for entry
func (fe *fileEntry) archivePath() (archivePath string) {
	archivePath = filepath.Join(fe.parentPath, fe.filename)
	return
}

func hasZipFileEntry(path string, feList *[]zipFileEntry) (found bool, fe zipFileEntry) {
	if len(*feList) == 0 {
		return
	}
	for _, fe = range *feList {
		if path == fe.path {
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

// Walk a file or a directory and gatehr file entries and error messages
func walkAllFilesInDir(path string, fileEntries *[]fileEntry, errorMsgs *[]string) (err error) {
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		*errorMsgs = append(*errorMsgs, err.Error())
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		*errorMsgs = append(*errorMsgs, err.Error())
		return
	}
	defer file.Close()

	rootPath, err := filepath.Abs(path)
	rootPath = filepath.Dir(rootPath)

	var curDir string
	var basePath string

	if !fileInfo.IsDir() {
		fe := fileEntry{}
		fe.rootPath = rootPath

		abs, _ := filepath.Abs(path)
		parentPath := filepath.Dir(abs)

		fe.rootPath = filepath.Dir(rootPath)
		fe.parentPath = parentPath

		fe.filename = fileInfo.Name()
		fe.isBareFile = true

		if !hasFileEntry(fe, fileEntries) {
			*fileEntries = append(*fileEntries, fe)
		}
		return
	}

	return filepath.Walk(path, func(path string, info os.FileInfo, e error) (err error) {
		if err != nil {
			*errorMsgs = append(*errorMsgs, e.Error())
			return err
		}

		if info.Name() == filepath.Base(path) {
			basePath, err = filepath.Abs(path)
			basePath = filepath.Dir(basePath)
		}
		if info.IsDir() {
			curDir = info.Name()
			curDir = filepath.Join(basePath, curDir)
		}
		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() {
			fe := fileEntry{}
			fe.parentPath = curDir
			// fe.rootPath = rootPath

			fe.filename = filepath.Join(info.Name())
			// start with base path since it is a directory
			*fileEntries = append(*fileEntries, fe)
		}
		return
	})
}

func unzip(path string) (err error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	var write = func(filePath string, f *zip.File) (err error) {
		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return
		}

		var dstFile *os.File
		dstFile, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return
		}

		defer dstFile.Close()

		var fileInArchive io.ReadCloser

		fileInArchive, err = f.Open()
		if err != nil {
			return
		}
		defer fileInArchive.Close()

		if _, err = io.Copy(dstFile, fileInArchive); err != nil {
			return
		}

		return
	}

	var fileEntries = []fileEntry{}
	var errorMsgs = []string{}

	// Populate list of files
	for _, path := range args.SourceFiles {
		// fmt.Println(args.SourceFiles)
		walkAllFilesInDir(path, &fileEntries, &errorMsgs)
	}
	if len(fileEntries) == 0 {
		fmt.Fprintln(os.Stderr, colour(brightRed, "no valid files found"))
		os.Exit(1)
	}

	zipFileEntries, err := zipFileList(path)

	for _, f := range archive.File {
		filePath := filepath.Join(args.Dir, f.Name)
		fmt.Println("unzipping file ", filePath)

		// if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
		// 	fmt.Println("invalid file path")
		// 	return
		// }
		if f.FileInfo().IsDir() {
			// fmt.Println("creating directory...")
			// os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return
		}

		err = write(filePath, f)
		if err != nil {
			return
		}
		// var dstFile *os.File
		// dstFile, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		// if err != nil {
		// 	return
		// }

		// var fileInArchive io.ReadCloser

		// fileInArchive, err = f.Open()
		// if err != nil {
		// 	return
		// }

		// if _, err = io.Copy(dstFile, fileInArchive); err != nil {
		// 	return
		// }
	}

	return
}

// printEntries of a zip file
func printEntries(path string) (err error) {
	zipFileEntries, err := zipFileList(path)
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
			file.path,
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
func zipFileList(path string) (entries []zipFileEntry, err error) {
	zf, err := zip.OpenReader(path)
	if err != nil {
		return
	}

	defer zf.Close()

	entries = make([]zipFileEntry, 0, 10)

	for _, file := range zf.File {
		entry := zipFileEntry{}
		entry.path = file.Name
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
	p := arg.MustParse(&args)

	if !args.Update && !args.Freshen {
		args.Update = true
	}
	if args.Update && args.Freshen {
		p.Fail("both -u (update) and -f (freshen) specified")
	}

	// Handle printing list of files in archive
	if args.List {
		err := printEntries(args.Zipfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, colour(brightRed, err.Error()))
		}
		os.Exit(0)
	}

	err := unzip(args.Zipfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, colour(brightRed, err.Error()))

	}
}
