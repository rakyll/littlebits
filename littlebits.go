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

type Reader struct {
	dev *portaudio.DeviceInfo
	s   *portaudio.Stream
	buf []byte
}

// NewReader creates a new io.Reader to read from the littleBits
// USB I/O module. If no name is given, the default audio device
// name, "KORG 2ch Audio Device". You can specify a specific name
// if you have more than a single USB I/O module connected to your
// computer.
// bufferSize allows you allocate a buffer for the specific reader
// that will be reused while reading from the module.
// All readers must be closed after you are done with reading by
// calling (*Reader).Close to free the underlying resources.
func NewReader(name string, bufferSize int) (*Reader, error) {
	if name == "" {
		name = defaultName
	}
	buf := make([]byte, bufferSize)
	dev, s, err := initDevice(name, true, buf)
	if err != nil {
		return nil, err
	}
	return &Reader{dev: dev, s: s, buf: buf}, nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if len(p) > len(r.buf) {
		return 0, fmt.Errorf("p is exceeding reader buffer size limit = %v", len(r.buf))
	}
	if err := r.s.Start(); err != nil {
		return 0, err
	}
	if err := r.s.Read(); err != nil {
		return 0, err
	}
	if err := r.s.Stop(); err != nil {
		return 0, err
	}
	copy(p, r.buf[:len(p)])
	return len(p), nil // TODO(jbd): is there a possibility we can read less than buffer size?
}

func (r *Reader) Close() error {
	// TODO(jbd): auto terminate portaudio if no devices are being used?
	return r.s.Close()
}

// type Writer struct {
// 	dev *portaudio.DeviceInfo
// 	s   *portaudio.Stream
// 	buf []byte
// }

// func NewWriter() (*Writer, error) {
// 	dev, s, buf, err := initDevice(false)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := s.Start(); err != nil {
// 		return nil, err
// 	}
// 	for i := range buf {
// 		buf[i] = 10
// 	}
// 	return &Writer{dev: dev, s: s, buf: buf}, nil
// }

// func (w *Writer) Write(p []byte) (n int, err error) {
// 	if len(p) > maxBufferSize {
// 		return 0, fmt.Errorf("buffer size cannot be larger than %d", maxBufferSize)
// 	}
// 	panic("not yet implemented")
// }

// func (w *Writer) Close() error {
// 	w.s.Stop()
// 	return w.s.Close()
// }
