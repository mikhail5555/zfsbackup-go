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
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/someone1/zfsbackup-go/backends"
	"github.com/someone1/zfsbackup-go/config"
	"github.com/someone1/zfsbackup-go/files"
	"github.com/someone1/zfsbackup-go/zfs"
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

func fakeExecCommand(ctx context.Context, _ *files.JobInfo) *exec.Cmd {
	cs := []string{"run", "./mock_zfs"}
	cmd := exec.CommandContext(ctx, "go", cs...)
	return cmd
}

func SetupMocks(info files.SnapshotInfo) func() {
	l, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(l)

	origExecCommand := zfs.GetZFSSendCommand
	zfs.GetZFSSendCommand = fakeExecCommand

	origsSnapshotCommand := zfs.GetSnapshotsAndBookmarks
	zfs.GetSnapshotsAndBookmarks = func(_ context.Context, _ string) ([]files.SnapshotInfo, error) {
		return []files.SnapshotInfo{info}, nil
	}

	origGetCreationDate := zfs.GetCreationDate
	zfs.GetCreationDate = func(ctx context.Context, target string) (time.Time, error) {
		return info.CreationTime, nil
	}

	return func() {
		zfs.GetZFSSendCommand = origExecCommand
		zfs.GetSnapshotsAndBookmarks = origsSnapshotCommand
		zfs.GetCreationDate = origGetCreationDate
	}
}

func TestBackup(t *testing.T) {
	baseSnapshot := files.SnapshotInfo{Name: "tank/test@snap1", CreationTime: time.Now()}

	undo := SetupMocks(baseSnapshot)
	defer undo()

	tempDir := os.TempDir()
	defer os.RemoveAll(tempDir)

	config.WorkingDir = tempDir

	// Create a job info for testing
	jobInfo := &files.JobInfo{
		VolumeName:         "tank/test",
		VolumeSize:         1, // 1 MiB
		UploadChunkSize:    1,
		Destinations:       []string{fmt.Sprintf("%s://test", backends.MockBackendPrefix)},
		BaseSnapshot:       baseSnapshot,
		MaxParallelUploads: 5,
		MaxBackoffTime:     5 * time.Millisecond,
		MaxRetryTime:       1 * time.Second,
		StartTime:          time.Now(),
		AesEncryptionKey:   "test1234test1234",
	}

	// Run the backup
	err := Backup(t.Context(), jobInfo)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify results
	if len(jobInfo.Volumes) == 0 {
		t.Errorf("Expected at least one volume to be created")
	}

	// Check that ZFS stream bytes were recorded
	if jobInfo.ZFSStreamBytes == 0 {
		t.Errorf("Expected ZFS stream bytes to be recorded")
	}

	// Verify that the manifest was created
	if !jobInfo.EndTime.After(jobInfo.StartTime) {
		t.Errorf("Expected end time to be after start time")
	}
}
