// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package littlebits contains an io.Reader and io.Writer to read from
// and write to a littleBits circuit. It requires USB I/O module.
package littlebits

import (
	"errors"
	"fmt"
	"strings"

	"code.google.com/p/portaudio-go/portaudio"
)

const defaultName = "KORG 2ch Audio Device"

func init() {
	portaudio.Initialize() // handle error
}

func initDevice(name string, in bool, buf []byte) (dev *portaudio.DeviceInfo, s *portaudio.Stream, err error) {
	i, err := portaudio.Devices()
	if err != nil {
		return nil, nil, err
	}

	for _, info := range i {
		// TODO(jbd): support multiple littebits usb I/O devices
		if strings.Contains(info.Name, name) {
			dev = info
			break
		}
	}
	if dev == nil {
		return nil, nil, errors.New("no little bits usb I/O found")
	}

	p := portaudio.LowLatencyParameters(dev, dev)
	if in {
		s, err = portaudio.OpenStream(p, buf, []byte{})
		return
	}
	s, err = portaudio.OpenStream(p, []byte{}, buf)
	return
}

// Reader is an io.Reader that reads from a littleBits USB I/O module.
type Reader struct {
	dev *portaudio.DeviceInfo
	s   *portaudio.Stream
	buf []byte
}

// NewReader creates a new io.Reader to read from the littleBits
// USB I/O module. If no name is given, the default audio device
// name, "KORG 2ch Audio Device", is used. You can specify a specific name
// if you have more than a single USB I/O module connected to your
// computer.
//
// bufferSize allows you allocate a buffer for the specific reader
// that will be reused while reading from the module.
// All readers must be closed after you are done with reading by
// calling (*Reader).Close to free the underlying resources.
//
// Your USB I/O module needs to be in the "out" mode to work
// properly with a Reader.
func NewReader(name string, bufferSize int) (*Reader, error) {
	if name == "" {
		name = defaultName
	}
	buf := make([]byte, bufferSize)
	dev, s, err := initDevice(name, true, buf)
	if err != nil {
		return nil, err
	}
	if err := s.Start(); err != nil {
		return nil, err
	}
	return &Reader{dev: dev, s: s, buf: buf}, nil
}

// Read reads len(p) bytes from the USB I/O module. If len(p)
// is larger than the reader buffer size, it returns with an
// error. Otherwise, fills up to n bytes it can read from the
// module and returns.
//
// n is always len(p), the call will block until n number
// of bytes are read from the module.
func (r *Reader) Read(p []byte) (n int, err error) {
	if len(p) > len(r.buf) {
		return 0, fmt.Errorf("len(p) is exceeding reader buffer size limit = %v", len(r.buf))
	}
	if err := r.s.Read(); err != nil {
		return 0, err
	}
	copy(p, r.buf[:len(p)])
	return len(p), nil // TODO(jbd): is there a possibility we can read less than buffer size?
}

// Close frees the underyling sources.
func (r *Reader) Close() error {
	if err := r.s.Stop(); err != nil {
		return err
	}
	// TODO(jbd): auto terminate portaudio if no devices are being used?
	return r.s.Close()
}

// A writer is an io.Writer that writes to an littleBits USB I/O module.
// Writer requires the module to be in the "in" mode.
type Writer struct {
	dev *portaudio.DeviceInfo
	s   *portaudio.Stream
	buf []byte
}

// NewWriter returns an a Writer to write to a littleBits USB I/O module.
// If no name is given, the default audio device name,
// "KORG 2ch Audio Device, is used. You have to specificy a name if
// there are multiple USB I/O modules are connected to your computer.
//
// // bufferSize allows you allocate a buffer for the specific writer
// that will be reused while writing to the module.
// All writers must be closed after you are done with reading by
// calling (*Reader).Close to free the underlying resources.
//
// Your USB I/O module needs to be in the "in" mode to work
// properly with a Writer.
func NewWriter(name string, bufferSize int) (*Writer, error) {
	if name == "" {
		name = defaultName
	}
	buf := make([]byte, bufferSize)
	dev, s, err := initDevice(name, false, buf)
	if err != nil {
		return nil, err
	}
	if err := s.Start(); err != nil {
		return nil, err
	}
	return &Writer{dev: dev, s: s, buf: buf}, nil
}

// Write writes len(p) bytes to the USB I/O module. If len(p)
// is larger than the reader buffer size, it returns with an
// error.
func (w *Writer) Write(p []byte) (n int, err error) {
	maxBufferSize := len(w.buf)
	if len(p) > maxBufferSize {
		return 0, fmt.Errorf("len(p) is exceeding reader buffer size limit = %v", maxBufferSize)
	}
	copy(w.buf, p)
	if err := w.s.Write(); err != nil {
		return 0, err
	}
	return len(p), err
}

// Close closes the Writer and frees underlying resources.
func (w *Writer) Close() error {
	if err := w.s.Stop(); err != nil {
		return err
	}
	return w.s.Close()
}
