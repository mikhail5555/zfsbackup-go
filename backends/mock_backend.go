// Copyright © 2016 Prateek Malhotra (someone1@gmail.com)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package backends

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"go.uber.org/zap"

	"github.com/someone1/zfsbackup-go/files"
)

var MockBackendImpl = &MockBackend{}

// MockBackendPrefix is the URI prefix used for the MockBackend.
const MockBackendPrefix = "Mock"

// MockBackend is a special Backend used to Mock the files after they've been uploaded
type MockBackend struct {
	DeleteMock func(ctx context.Context, filename string) error

	inMemoryStore sync.Map
}

// Init will initialize the MockBackend (aka do nothing)
func (d *MockBackend) Init(ctx context.Context, conf *BackendConfig, opts ...Option) error {
	d.inMemoryStore = sync.Map{}
	return nil
}

func (d *MockBackend) Delete(ctx context.Context, filename string) error {
	d.inMemoryStore.Delete(filename)
	return nil
}

// PreDownload should not be used with this backend
func (d *MockBackend) PreDownload(ctx context.Context, objects []string) error {
	return nil
}

// Download should not be used with this backend
func (d *MockBackend) Download(ctx context.Context, filename string) (io.ReadCloser, error) {
	if value, ok := d.inMemoryStore.Load(filename); ok {
		return io.NopCloser(bytes.NewReader(value.([]byte))), nil
	}

	return nil, errors.New("file not found")
}

// Close does nothing for this backend.
func (d *MockBackend) Close() error {
	return nil
}

// List should not be used with this backend
func (d *MockBackend) List(ctx context.Context, prefix string) ([]string, error) {
	allKeys := make([]string, 0)
	d.inMemoryStore.Range(func(key, value interface{}) bool {
		allKeys = append(allKeys, key.(string))
		return true
	})
	return allKeys, nil
}

// Upload will Mock the provided volume, usually found in a temporary folder
func (d *MockBackend) Upload(ctx context.Context, vol *files.VolumeInfo) error {
	zap.S().Infof("Mock backend: Upload volume info: %v", vol.ObjectName)
	buffer := new(bytes.Buffer)
	n, err := io.Copy(buffer, vol)
	if err != nil {
		zap.S().Errorf("Mock backend: Upload volume info failed: %v", err)
		return err
	}
	d.inMemoryStore.Store(vol.ObjectName, buffer.Bytes())
	zap.S().Infof("Mock backend: Upload done for: %v, size: %d", vol.ObjectName, n)
	return nil
}

func (d *MockBackend) Reset() {
	d.inMemoryStore.Clear()
}
