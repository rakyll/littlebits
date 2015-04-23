package littlebits

import (
	"errors"
	"fmt"
	"strings"

	"code.google.com/p/portaudio-go/portaudio"
)

const name = "KORG 2ch Audio Device"

const maxBufferSize = 1024

func init() {
	portaudio.Initialize() // handle error
}

type Reader struct {
	dev *portaudio.DeviceInfo
	s   *portaudio.Stream
	buf []byte
}

func NewReader() (*Reader, error) {
	i, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	var dev *portaudio.DeviceInfo
	for _, info := range i {
		// TODO(jbd): support multiple littebits usb I/O devices
		if strings.Contains(info.Name, name) {
			dev = info
			break
		}
	}
	if dev == nil {
		return nil, errors.New("no little bits usb I/O found")
	}

	buf := make([]byte, maxBufferSize)
	p := portaudio.LowLatencyParameters(dev, dev)
	s, err := portaudio.OpenStream(p, buf, []byte{})
	if err != nil {
		return nil, err
	}
	return &Reader{dev: dev, s: s, buf: buf}, nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if len(p) > maxBufferSize {
		return 0, fmt.Errorf("buffer size cannot be larger than %d", maxBufferSize)
	}
	if err := r.s.Start(); err != nil {
		return 0, err
	}
	defer r.s.Stop() // TODO(jbd): Error if stop fails?
	if err := r.s.Read(); err != nil {
		return 0, err
	}
	copy(p, r.buf[:len(p)])
	return len(p), nil // TODO(jbd): is there a possibility we can read less than buffer size?
}

func (r *Reader) Close() error {
	return r.s.Close()
	// TODO(jbd): auto terminate portaudio if no devices are
	// being used.
}

type Writer struct {
	dev *portaudio.DeviceInfo
}
