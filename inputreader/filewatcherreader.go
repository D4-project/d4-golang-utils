package inputreader

import (
	"bytes"
	"encoding/base64"
	"github.com/rjeczalik/notify"
	"io"
	"log"
	"os"
)

// FileWatcherReader is an abstraction of a folder watcher
// and behaves like a reader
type FileWatcherReader struct {
	// Folder to watch
	folderfd *os.File
	// Notify Channel
	eic chan notify.EventInfo
	// TearDown channel
	exit chan string
	// Current buffer
	json bool
	// Current file
	curfile *os.File
	// Current state Watching / Reading
	watching bool
	// Insert Separator
	insertsep bool
}

// NewFileWatcherReader creates a new FileWatcherReader
// json specifies whether we now we handle json files
func NewFileWatcherReader(f *os.File, j bool) (*FileWatcherReader, error) {
	r := &FileWatcherReader{
		folderfd: f,
		eic:      make(chan notify.EventInfo, 4096),
		json:     j,
		watching: true,
		insertsep: false,
	}
	// go routine holding the watcher
	go func() {
		if err := notify.Watch("./...", r.eic, notify.InCloseWrite); err != nil {
			log.Fatal(err)
		}
		defer notify.Stop(r.eic)
		<-r.exit
	}()
	return r, nil
}

// Read  waits for InCloseWrite file event uses a bytes reader to copy
// the resulting file encoded in b64 in p
func (fw *FileWatcherReader) Read(p []byte) (n int, err error) {

	// Watching for new files to read
	if fw.watching {
	watchloop:
		for {
			select {
			case ei := <-fw.eic:
				//log.Println("Got event:", ei)
				// New File, let's read its content
				var err error
				fw.curfile, err = os.Open(ei.Path())
				if err != nil {
					log.Fatal(err)
				}
				fw.watching = false
				break watchloop
			case <-fw.exit:
				// Exiting
				return 0, io.EOF
			}
		}
	}

	// Inserting separator
	if fw.insertsep{
		var buf []byte
		buf = append(buf, "\n"...)
		rreader := bytes.NewReader(buf)
		n, err = rreader.Read(p)
		fw.watching = true
		fw.insertsep = false
		log.Println("Inserting file seperator ")
		fw.curfile.Close()
		return n, err
	}

	// Reading
	// if not json it could be anything so we encode it in b64
	if !fw.json {
		// base64 stream encoder
		b64buffer := new(bytes.Buffer)
		b64encoder := base64.NewEncoder(base64.StdEncoding, b64buffer)
		// max buffer size is then 3072 = 4096/4*3
		buf := make([]byte, 3072)
		bytesread, err := fw.curfile.Read(buf)
		// buf is the input
		b64encoder.Write(buf[:bytesread])
		// Close the encoder to flush partially written blocks
		b64encoder.Close()
		log.Println(b64buffer)
		if err == io.EOF {
			fw.insertsep = true
		}
		// Copy from b64buffer to p
		n, err = b64buffer.Read(p[:len(b64buffer.Bytes())])
		log.Println("len b64 buffer: ", len(b64buffer.Bytes()))
	} else {
		n, err = fw.curfile.Read(p)
		if err == io.EOF {
			fw.insertsep = true
		}
	}
	log.Println("nread: ", n)
	return n, err
}

// Teardown is called on error to stop the Reading loop if needed
func (rl *FileWatcherReader) Teardown() {
	rl.exit <- "exit"
}
