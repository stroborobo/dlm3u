package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/Knorkebrot/m3u"
	"github.com/cheggaaa/pb"
	flag "github.com/ogier/pflag"
)

var (
	checkNameOnly bool
	target        string
)

const (
	PS = string(os.PathSeparator)
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s list.m3u targetdir\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.BoolVarP(&checkNameOnly, "name-only", "n", false,
		"Skip existing files only based on name, don't check it's size.")

	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	input, err := os.Open(flag.Arg(0))
	exitErr(err)

	pls, err := m3u.Parse(input)
	exitErr(err)

	target = flag.Arg(1)
	stat, err := os.Stat(target)
	exitErr(err)
	if !stat.IsDir() {
		exitPrint("Error: not a directory:", target)
	}

	for _, song := range pls {
		download(song)
	}
}

func getURL(str string) *url.URL {
	u, err := url.Parse(str)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: malformed url:", str, "skipping.")
		return nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		fmt.Fprintln(os.Stderr, "Error: not a valid http(s) url:", str, "skipping.")
		return nil
	}
	return u
}

func checkExists(path string, size int64) bool {
	if checkNameOnly && size > 0 {
		// already checked
		return false
	} else if !checkNameOnly && size == 0 {
		// check later
		return false
	}

	fi, err := os.Stat(path)
	exists := err == nil
	if !checkNameOnly {
		exists = exists && fi.Size() == size
	}
	if exists {
		fmt.Println(" already downloaded, skipping.")
	}
	return exists
}

func download(song *m3u.Song) {
	u := getURL(song.Path)
	if u == nil {
		return
	}

	title := song.Title
	if title == "" {
		title = path.Base(u.Path)
	}
	fmt.Print(title + ":")

	// target file path
	path := target + PS + path.Base(u.Path)
	if checkExists(path, 0) {
		return
	}

	// start the request
	resp, err := http.Get(song.Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: http get failed:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "Error: http status not ok:", resp.Status)
		return
	}
	if checkExists(path, resp.ContentLength) {
		return
	}

	fmt.Println("")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	exitErr(err, "Error: cannot open", path+":", err)

	// progressbar
	bar := pb.New(int(resp.ContentLength))
	bar.SetUnits(pb.U_BYTES)
	bar.SetMaxWidth(79)
	bar.Start()
	defer bar.Finish()

	writer := io.MultiWriter(file, bar)
	_, err = io.Copy(writer, resp.Body)

	file.Close()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: download failed:", err)
		os.Remove(path)
		return
	}
}
