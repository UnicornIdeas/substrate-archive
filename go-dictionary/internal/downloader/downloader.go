package downloader

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pierrec/lz4"
)

type DownServer struct {
	URL                string
	OutputFolder       string
	CurrentSnapshot    string
	DecompressedFolder string
	RocksdbFolder      string
}

type WriteCounter struct {
	Total uint64
}

func InitDownServer(url string, currentSnapshot string) DownServer {
	downServer := DownServer{}
	downServer.URL = url
	downServer.OutputFolder = "./snapshots/"
	err := os.MkdirAll(downServer.OutputFolder, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	downServer.DecompressedFolder = "./decompressed/"
	err = os.MkdirAll(downServer.DecompressedFolder, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	downServer.CurrentSnapshot = currentSnapshot
	downServer.RocksdbFolder = "./rocksdbs/"
	err = os.MkdirAll(downServer.RocksdbFolder, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	return downServer
}

func (ds *DownServer) DownloadSnapshot() error {
	fileURL, err := url.Parse(ds.URL)
	if err != nil {
		return err
	}

	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	// Create output folder
	err = os.MkdirAll(ds.OutputFolder, os.ModePerm)
	if err != nil {
		return err
	}
	// Create blank file
	file, err := os.Create(ds.OutputFolder + fileName)
	if err != nil {
		return err
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(ds.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Printf("Downloaded a file %s with size %d", fileName, size)

	return nil
}

func (ds *DownServer) DecompressSnapshot() error {
	// open input file
	fin, err := os.Open(ds.OutputFolder + ds.CurrentSnapshot)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fin.Close(); err != nil {
			panic(err)
		}
	}()

	// make an lz4 read buffer
	r := lz4.NewReader(fin)

	ds.CurrentSnapshot = strings.ReplaceAll(ds.CurrentSnapshot, ".lz4", "")
	// open output file
	fout, err := os.Create(ds.DecompressedFolder + ds.CurrentSnapshot)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fout.Close(); err != nil {
			panic(err)
		}
	}()

	// make a write buffer
	w := bufio.NewWriter(fout)

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := w.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}
	return nil
}

func (ds *DownServer) UntarSnapshot() error {
	if strings.Contains(ds.CurrentSnapshot, ".lz4") {
		ds.CurrentSnapshot = strings.ReplaceAll(ds.CurrentSnapshot, ".lz4", "")
	}
	file, err := os.Open(ds.DecompressedFolder + ds.CurrentSnapshot)
	if err != nil {
		return err
	}

	// gzr, err := gzip.NewReader(file)
	// if err != nil {
	// 	log.Println("gzr error")
	// 	return err
	// }
	// defer gzr.Close()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(ds.RocksdbFolder, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
