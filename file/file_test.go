// Copyright © 2017 Meltwater
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"log"
	"sync"
	"testing"

	"github.com/inside-track/rabbitio/rmq"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNewInput(t *testing.T) {
	assert := assert.New(t)
	fs = afero.NewMemMapFs()

	fs.MkdirAll("datadir", 0755)
	afero.WriteFile(fs, "datadir/file1.tgz", []byte("mymessage"), 0644)
	afero.WriteFile(fs, "datadir/file2.tgz", []byte("mymessage"), 0644)
	path, err := NewInput("datadir")

	_, err2 := NewInput("datadir_notthere")
	_, fileErr := NewInput("datadir/file2.tgz")

	fs = afero.NewReadOnlyFs(fs)

	assert.Nil(err, "should return no error")
	assert.Len(path.queue, 2, "Should have 2 items")
	assert.NotNil(err2, "should return error on directory not there")
	assert.Nil(fileErr, "should handle a single file input")
}

// TestNewOutput will make sure we create directory when missing.
// also checks that it is able to
func TestNewOutput(t *testing.T) {
	assert := assert.New(t)
	fs = afero.NewMemMapFs()

	fs.MkdirAll("data", 0755)
	_, err := NewOutput("data/creates_directory", 100)

	fs = afero.NewReadOnlyFs(fs)
	_, err2 := NewOutput("data/creates_directory2", 100)

	assert.Nil(err, "should return no error")
	assert.NotNil(err2, "should not be able to create output")
}

func TestPath_Create(t *testing.T) {
	assert := assert.New(t)
	fs = afero.NewMemMapFs()

	p := &Path{
		name:      "mypath",
		batchSize: 100,
	}
	p2 := &Path{
		name:      "mypath2",
		batchSize: 100,
	}

	err := p.create()

	fs = afero.NewReadOnlyFs(fs)
	err2 := p2.create()

	assert.Nil(err, "should return no error")
	assert.NotNil(err2, "should return error")

}

func tarball() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: "file.tgz",
		Mode: 0600,
		Size: int64(len("mybody")),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		log.Fatal(err)
	}
	if _, err := tw.Write([]byte("mybody")); err != nil {
		log.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}

func TestPath_Send(t *testing.T) {
	assert := assert.New(t)
	fs = afero.NewMemMapFs()

	fs.MkdirAll("/datadir", 0755)
	fs.MkdirAll("/nofilesdatadir", 0755)
	fs.MkdirAll("/data", 0755)
	afero.WriteFile(fs, "/datadir/file1.tgz", []byte("mymessage"), 0644)
	afero.WriteFile(fs, "/datadir/file2.tgz", []byte("mymessage"), 0644)
	afero.WriteFile(fs, "/data/tarball.tgz", tarball(), 0644)

	path, _ := NewInput("/datadir/")
	workingPath, _ := NewInput("/data/tarball.tgz")

	noFilesPath := &Path{queue: []string{"nodir/nofile"}}

	// m := make(chan rmq.Message)
	path.Wg = new(sync.WaitGroup)
	workingPath.Wg = new(sync.WaitGroup)

	openErr := noFilesPath.Send(make(chan rmq.Message, 1))
	invErr := path.Send(make(chan rmq.Message, 1))

	ch := make(chan rmq.Message, 1)
	go func() {
		for range ch {
			workingPath.Wg.Done()
		}
	}()
	noErr := workingPath.Send(ch)

	assert.Error(openErr, "should return error as directory and file is not there")
	assert.Error(invErr, "received unexpected error unexpected EOF")
	assert.NoError(noErr, "should return no error")
}

func TestWriteFile(t *testing.T) {
	assert := assert.New(t)
	fs = afero.NewMemMapFs()

	fs.MkdirAll("datadir", 0755)
	myBytes := []byte("mydatawritten")
	err := writeFile(myBytes, "datadir", "datafile")

	fs = afero.NewReadOnlyFs(fs)
	err2 := writeFile(myBytes, "datadir", "datafile")

	assert.Nil(err, "should return no error")
	assert.NotNil(err2, "should return error")
}
