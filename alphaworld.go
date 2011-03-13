package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

type AlphaWorld struct {
	worldDir string
}

func (w *AlphaWorld) OpenChunk(x, z int) (io.ReadCloser, os.Error) {
	var file, fileErr = os.Open(chunkPath(w.worldDir, x, z), os.O_RDONLY, 0666)
	if fileErr != nil {
		return nil, fileErr
	}
	var decompressor, gzipErr = gzip.NewReader(file)
	if gzipErr != nil {
		file.Close()
		return nil, gzipErr
	}
	return &ReadCloserPair{decompressor, file}, nil
}

type AlphaChunkPool struct {
	chunkMap map[string]bool
	worldDir string
}

func (p *AlphaChunkPool) Pop(x, z int) bool {
	var chunkFilename = chunkPath(p.worldDir, x, z)
	var _, exists = p.chunkMap[chunkFilename]
	p.chunkMap[chunkFilename] = false, false
	return exists
}

func (p *AlphaChunkPool) Remaining() int {
	return len(p.chunkMap)
}

func (w *AlphaWorld) ChunkPool() (ChunkPool, os.Error) {
	var errors = make(chan os.Error, 5)
	var done = make(chan bool)
	go func() {
		for error := range errors {
			fmt.Fprintln(os.Stderr, error) // TODO: return errors
		}
		done <- true
	}()
	var v = &visitor{make(map[string]bool)}
	filepath.Walk(w.worldDir, v, errors)
	close(errors)
	<-done
	return &AlphaChunkPool{v.chunks, w.worldDir}, nil
}

type visitor struct {
	chunks map[string]bool
}

func (v *visitor) VisitDir(dir string, f *os.FileInfo) bool {
	return true
}

func (v *visitor) VisitFile(file string, f *os.FileInfo) {
	var match, err = path.Match("c.*.dat", path.Base(file))
	if match && err == nil {
		v.chunks[file] = true
	}
}
