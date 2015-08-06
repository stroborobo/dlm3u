package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/Knorkebrot/m3u"
	"github.com/cheggaaa/pb"
)

const (
	PS = string(os.PathSeparator)
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s list.m3u targetdir\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	input, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(10)
	}

	pls, err := m3u.Parse(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(20)
	}
	target := flag.Arg(1)
	stat, err := os.Stat(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(30)
	} else if !stat.IsDir() {
		fmt.Fprintln(os.Stderr, "Error: not a directory:", target)
		os.Exit(40)
	}

	for _, song := range pls {
		fmt.Print(song.Title + ":")
		u, err := url.Parse(song.Path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: malformed url:", song.Path, "skipping.")
			continue
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			fmt.Fprintln(os.Stderr, "Error: not a valid http(s) url:", song.Path, "skipping.")
			continue
		}

		resp, err := http.Get(song.Path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: http get failed:", err)
			continue
		}

		path := target + PS + path.Base(u.Path)

		fi, err := os.Stat(path)
		if err == nil && fi.Size() == resp.ContentLength {
			resp.Body.Close()
			fmt.Println(" already downloaded, skipping.")
			continue
		}

		fmt.Println("")
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: cannot open", path+":", err)
			resp.Body.Close()
			os.Exit(50)
		}

		bar := pb.New(int(resp.ContentLength))
		bar.SetUnits(pb.U_BYTES)
		bar.SetMaxWidth(79)
		bar.Start()
		writer := io.MultiWriter(file, bar)
		_, err = io.Copy(writer, resp.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: download failed:", err)
			resp.Body.Close()
			file.Close()
			os.Remove(path)
			continue
		}
		resp.Body.Close()
		file.Close()
		bar.Finish()
	}
}
