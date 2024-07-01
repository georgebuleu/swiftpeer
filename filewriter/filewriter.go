package filewriter

import (
	"os"
	"syscall"
)

type FileWriter struct {
	File *os.File
	Data []byte
	Size int
}

func New(path string, size int) (*FileWriter, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	if err = f.Truncate(int64(size)); err != nil {
		f.Close()
		return nil, err
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &FileWriter{File: f, Data: data, Size: size}, nil
}

func (fw *FileWriter) WriteAt(data []byte, offset int) error {
	if offset+len(data) > fw.Size {
		return syscall.EINVAL // Offset out of range
	}
	copy(fw.Data[offset:], data)
	return nil
}

func (fw *FileWriter) Sync() error {
	return fw.File.Sync()
}

func (fw *FileWriter) Close() error {
	if err := syscall.Munmap(fw.Data); err != nil {
		return err
	}
	return fw.File.Close()
}
