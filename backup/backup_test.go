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

package backup

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"os"
	"testing"
	"time"

	"github.com/someone1/zfsbackup-go/backends"
	"github.com/someone1/zfsbackup-go/files"
)

// Truly a useless backend
type mockBackend struct{}

func (m *mockBackend) Init(ctx context.Context, conf *backends.BackendConfig, opts ...backends.Option) error {
	return nil
}

func (m *mockBackend) Upload(ctx context.Context, vol *files.VolumeInfo) error {
	// make sure we can read the volume
	_, err := io.ReadAll(vol)
	return err
}

func (m *mockBackend) List(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}

func (m *mockBackend) Close() error { return nil }

func (m *mockBackend) PreDownload(ctx context.Context, objects []string) error { return nil }

func (m *mockBackend) Download(ctx context.Context, filename string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockBackend) Delete(ctx context.Context, filename string) error { return nil }

type errTestFunc func(error) bool

func nilErrTest(e error) bool { return e == nil }

func TestRetryUploadChainer(t *testing.T) {
	_, goodVol, badVol, err := prepareTestVols()
	if err != nil {
		t.Fatalf("error preparing volumes for testing - %v", err)
	}

	testCases := []struct {
		name  string
		vol   *files.VolumeInfo
		valid errTestFunc
	}{
		{
			name:  "good",
			vol:   goodVol,
			valid: nilErrTest,
		},
		{
			name:  "bad",
			vol:   badVol,
			valid: os.IsNotExist,
		},
	}

	j := &files.JobInfo{
		MaxParallelUploads: 1,
		MaxBackoffTime:     5 * time.Millisecond,
		MaxRetryTime:       1 * time.Second,
	}

	for idx, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			b := &mockBackend{}
			if err := b.Init(t.Context(), nil); err != nil {
				t.Errorf("%d: Expected error %v, got %v", idx, nil, err)
				return
			}

			in := make(chan *files.VolumeInfo, 1)
			out, wg := retryUploadChainer(t.Context(), in, b, j, "mock://")
			in <- testCase.vol
			close(in)
			outVol := <-out
			if errResult := wg.Wait(); !testCase.valid(errResult) {
				t.Errorf("%d: error %v id not pass validation function", idx, errResult)
			} else if errResult == nil {
				// Verify we got the same vol we passed in!
				if outVol != testCase.vol {
					t.Errorf("did not get same volume passed in back out")
				}
			}
		})
	}
}

func prepareTestVols() (payload []byte, goodVol, badVol *files.VolumeInfo, err error) {
	payload = make([]byte, 10*1024*1024)
	if _, err = rand.Read(payload); err != nil {
		return
	}
	reader := bytes.NewReader(payload)
	goodVol, err = files.CreateSimpleVolume()
	if err != nil {
		return
	}
	_, err = io.Copy(goodVol, reader)
	if err != nil {
		return
	}
	err = goodVol.Close()
	if err != nil {
		return
	}

	badVol, err = files.CreateSimpleVolume()
	if err != nil {
		return
	}
	err = badVol.Close()
	if err != nil {
		return
	}

	return payload, goodVol, badVol, err
}
