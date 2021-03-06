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
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/inside-track/rabbitio/rmq"
	"github.com/spf13/afero"
)

// fs is our os.File handler
var fs = afero.NewOsFs()

// Path is a directory file path
type Path struct {
	name      string
	batchSize int
	queue     []string
	Wg        *sync.WaitGroup
}

// NewInput returns a *Path with a queue of files paths, all files in a directory
func NewInput(path string) (*Path, error) {
	fi, err := fs.Stat(path)
	if err != nil {
		// log.Fatalln(err)
		return nil, err
	}

	q := []string{}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, err := afero.ReadDir(fs, path)
		if err != nil {
			return nil, err
			//	log.Fatalf("Couldn't get directory or file: %s", err)
		}
		log.Printf("Found %d file(s) in %s", len(files), path)
		for _, f := range files {
			q = append(q, filepath.Join(path, f.Name()))
		}
	case mode.IsRegular():
		q = append(q, path)
	}

	return &Path{queue: q}, nil
}

func writeFile(b []byte, dir, file string) error {
	filePath := filepath.Join(dir, file)
	err := afero.WriteFile(fs, filePath, b, 0644)
	if err != nil {
		return err
	}
	log.Printf("Wrote %d bytes to %s", len(b), filePath)
	return nil
}

// Send delivers messages to the channel
func (p *Path) Send(messages chan rmq.Message) error {
	var num int

	// loop over the queued up files
	for _, file := range p.queue {
		// open file from the queue
		fh, err := fs.Open(file)
		if err != nil {
			return err
			// log.Fatalf("failed to open file: %s", err)
		}
		// and clean up afterwards
		defer fh.Close()

		tarNum, err := UnPack(p.Wg, fh, messages)
		if err != nil {
			return err
			//log.Fatalf("Failed to unpack: %s ", err)
		}
		log.Printf("Extracted %d Messages from tarball: %s", tarNum, file)
		num = num + tarNum
	}

	p.Wg.Wait()
	close(messages)
	// when all files are read, close
	log.Printf("Total %d Messages from tarballs", num)
	return nil
}

// NewOutput creates a Path to output files in from RabbitMQ
func NewOutput(path string, batchSize int) (*Path, error) {

	p := &Path{
		name:      path,
		batchSize: batchSize,
	}

	if err := p.create(); err != nil {
		return p, err
	}

	return p, nil
}

// Create creates the target directory if missing
func (p *Path) create() error {
	if _, err := fs.Stat(p.name); os.IsNotExist(err) {
		err := fs.MkdirAll(p.name, os.ModePerm)
		if err != nil {
			return err
		}
		log.Println("Created missing directory:", p.name)
	}
	return nil
}

// Receive will handle messages and save to path
func (p *Path) Receive(messages chan rmq.Message, verify chan rmq.Verify) error {

	// create new TarballBuilder
	builder, err := NewTarballBuilder(p.batchSize)
	if err != nil {
		return err
	}

	return builder.Pack(messages, p.name, verify)
}
