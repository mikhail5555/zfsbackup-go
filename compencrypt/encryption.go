package compencrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"sync"
)

var _ io.WriteCloser = (*EncryptionWriter)(nil)

type EncryptionWriter struct {
	w           io.WriteCloser
	destination io.WriteCloser
	key         []byte
	once        sync.Once
}

func NewEncryptionWriter(destination io.WriteCloser, key []byte) *EncryptionWriter {
	return &EncryptionWriter{
		destination: destination,
		key:         key,
	}
}

func (ew *EncryptionWriter) Write(p []byte) (int, error) {
	var onceErr error
	ew.once.Do(func() {
		if len(ew.key) == 0 {
			ew.w = ew.destination
			return
		}

		block, err := aes.NewCipher(ew.key)
		if err != nil {
			onceErr = err
			return
		}

		iv := make([]byte, aes.BlockSize)
		if _, err := rand.Read(iv); err != nil {
			onceErr = err
			return
		}

		stream := cipher.NewCTR(block, iv)

		if _, err := ew.destination.Write(iv); err != nil {
			onceErr = err
			return
		}

		ew.w = cipher.StreamWriter{S: stream, W: ew.destination}
	})

	if onceErr != nil {
		return 0, onceErr
	}

	return ew.w.Write(p)
}

func (ew *EncryptionWriter) Close() error {
	if ew.w != nil {
		return ew.w.Close()
	}
	return nil
}

var _ io.ReadCloser = (*DecryptionReader)(nil)

type DecryptionReader struct {
	r      io.Reader
	source io.ReadCloser
	key    []byte
	once   sync.Once
}

func (dr *DecryptionReader) Close() error {
	return dr.source.Close()
}

func NewDecryptionReader(source io.ReadCloser, key []byte) *DecryptionReader {
	return &DecryptionReader{
		source: source,
		key:    key,
	}
}

func (dr *DecryptionReader) Read(p []byte) (int, error) {
	var onceErr error
	dr.once.Do(func() {
		if len(dr.key) == 0 {
			dr.r = dr.source
			return
		}

		block, err := aes.NewCipher(dr.key)
		if err != nil {
			onceErr = err
			return
		}

		iv := make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(dr.source, iv); err != nil {
			onceErr = err
			return
		}

		stream := cipher.NewCTR(block, iv)
		dr.r = &cipher.StreamReader{S: stream, R: dr.source}
	})
	if onceErr != nil {
		return 0, onceErr
	}

	return dr.r.Read(p)
}

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nwc *nopWriteCloser) Close() error { return nil }
