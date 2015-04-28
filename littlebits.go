package littlebits

import (
	"errors"
	"fmt"
	"strings"

	"code.google.com/p/portaudio-go/portaudio"
)

const name = "KORG 2ch Audio Device"

const maxBufferSize = 8192

func init() {
	portaudio.Initialize() // handle error
}

func initDevice(in bool) (dev *portaudio.DeviceInfo, s *portaudio.Stream, buf []byte, err error) {
	i, err := portaudio.Devices()
	if err != nil {
		return nil, nil, nil, err
	}

	for _, info := range i {
		// TODO(jbd): support multiple littebits usb I/O devices
		if strings.Contains(info.Name, name) {
			dev = info
			break
		}
	}
	if dev == nil {
		return nil, nil, nil, errors.New("no little bits usb I/O found")
	}

	buf = make([]byte, maxBufferSize)
	p := portaudio.LowLatencyParameters(dev, dev)
	if in {
		s, err = portaudio.OpenStream(p, buf, []byte{})
		return
	}
	s, err = portaudio.OpenStream(p, []byte{}, buf)
	return
}

type Reader struct {
	dev *portaudio.DeviceInfo
	s   *portaudio.Stream
	buf []byte
}

func NewReader() (*Reader, error) {
	dev, s, buf, err := initDevice(true)
	if err != nil {
		return nil, err
	}
	if err := s.Start(); err != nil {
		return nil, err
	}
	return &Reader{dev: dev, s: s, buf: buf}, nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if len(p) > maxBufferSize {
		return 0, fmt.Errorf("buffer size cannot be larger than %d", maxBufferSize)
	}
	if err := r.s.Read(); err != nil {
		return 0, err
	}
	copy(p, r.buf[:len(p)])
	return len(p), nil // TODO(jbd): is there a possibility we can read less than buffer size?
}

func (r *Reader) Close() error {
	r.s.Stop()
	return r.s.Close()
	// TODO(jbd): auto terminate portaudio if no devices are
	// being used.
}

type Writer struct {
	dev *portaudio.DeviceInfo
	s   *portaudio.Stream
	buf []byte
}

func NewWriter() (*Writer, error) {
	dev, s, buf, err := initDevice(false)
	if err != nil {
		return nil, err
	}
	if err := s.Start(); err != nil {
		return nil, err
	}
	for i := range buf {
		buf[i] = 10
	}
	return &Writer{dev: dev, s: s, buf: buf}, nil
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if len(p) > maxBufferSize {
		return 0, fmt.Errorf("buffer size cannot be larger than %d", maxBufferSize)
	}
	// copy(w.buf, p)
	fmt.Println(w.buf)
	if err := w.s.Write(); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *Writer) Close() error {
	w.s.Stop()
	return w.s.Close()
}
