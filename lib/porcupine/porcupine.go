package porcupine

// #cgo LDFLAGS: -lpv_porcupine
// #include <stdlib.h>
// #include "pv_porcupine.h"
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

var (
	ErrOutOfMemory = errors.New("porcupine: out of memory")
	ErrIOError = errors.New("porcupine: IO error")
	ErrInvalidArgument = errors.New("porcupine: invalid argument")
	ErrUnknownStatus = errors.New("unknown status code")
)

func SampleRate() int {
	tmp := C.pv_sample_rate()
	return int(tmp)
}

func FrameLength() int {
	tmp := C.pv_porcupine_frame_length()
	return int(tmp)
}

type Porcupine interface {
	Process(frame []int16) (string, error)
	Close()
}

func New(modelPath string, keyword *Keyword) (Porcupine, error) {
	var h *C.struct_pv_porcupine
	mf := C.CString(modelPath)
	kf := C.CString(keyword.FilePath)
	sensitivity := C.float(keyword.Sensitivity)

	defer func() {
		C.free(unsafe.Pointer(mf))
		C.free(unsafe.Pointer(kf))
	}()

	status := C.pv_porcupine_init(mf, C.int32_t(1), &kf, &sensitivity, &h)
	if err := checkStatus(status); err != nil {
		return nil, err
	}

	return &SingleKeywordHandle{
		handle: &handle{h: h},
		kw:     keyword,
	}, nil
}

type handle struct {
	once sync.Once
	h    *C.struct_pv_porcupine
}

func (h *handle) Close() {
	h.once.Do(func() {
		C.pv_porcupine_delete(h.h)
		h.h = nil
	})
}

type Keyword struct {
	Value       string
	FilePath    string
	Sensitivity float32
}

type SingleKeywordHandle struct {
	*handle
	kw *Keyword
}

func (s *SingleKeywordHandle) Process(data []int16) (string, error) {
	var result C.int32_t
	status := C.pv_porcupine_process(s.handle.h, (*C.int16_t)(unsafe.Pointer(&data[0])), &result)
	if err := checkStatus(status); err != nil || int(result) == -1 {
		return "", err
	}

	return s.kw.Value, nil
}

func checkStatus(status C.pv_status_t) error {
	switch status {
	case C.PV_STATUS_SUCCESS:
		return nil
	case C.PV_STATUS_OUT_OF_MEMORY:
		return ErrOutOfMemory
	case C.PV_STATUS_INVALID_ARGUMENT:
		return ErrInvalidArgument
	case C.PV_STATUS_IO_ERROR:
		return ErrIOError
	default:
		return ErrUnknownStatus
	}
}
