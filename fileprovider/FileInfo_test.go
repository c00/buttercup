package fileprovider

import (
	"testing"
	"time"
)

func GetDate(hourOffset int) time.Time {
	return time.Date(2020, 6, 10, 12, 0, 0, 0, time.UTC).Add(time.Hour * time.Duration(hourOffset))
}

func TestFileInfo_CompareForPull(t *testing.T) {
	tests := []struct {
		name string

		localUpdated  time.Time
		localSynced   time.Time
		remoteUpdated time.Time

		want    int
		wantErr bool
	}{
		{name: "Same Date", localUpdated: GetDate(0), localSynced: GetDate(0), remoteUpdated: GetDate(0), want: UpToDate},
		{name: "Local later", localUpdated: GetDate(1), localSynced: GetDate(0), remoteUpdated: GetDate(0), want: LocalNewer},
		{name: "Remote later", localUpdated: GetDate(0), localSynced: GetDate(0), remoteUpdated: GetDate(1), want: RemoteNewer},
		{name: "Both later, local newest", localUpdated: GetDate(5), localSynced: GetDate(0), remoteUpdated: GetDate(2), want: ConflictLocalNewer},
		{name: "Both later, remote newest", localUpdated: GetDate(2), localSynced: GetDate(0), remoteUpdated: GetDate(5), want: ConflictRemoteNewer},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := FileInfo{
				Updated: tt.remoteUpdated,
			}
			got, err := rf.Compare(FileInfo{
				Updated:    tt.localUpdated,
				LastSynced: tt.localSynced,
			}, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileInfo.CompareForPull() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileInfo.CompareForPull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileInfo_CompareForPush(t *testing.T) {
	tests := []struct {
		name string

		localUpdated  time.Time
		localSynced   time.Time
		remoteUpdated time.Time

		want    int
		wantErr bool
	}{
		{name: "Same Date", localUpdated: GetDate(0), localSynced: GetDate(0), remoteUpdated: GetDate(0), want: UpToDate},
		{name: "Local later", localUpdated: GetDate(1), localSynced: GetDate(0), remoteUpdated: GetDate(0), want: LocalNewer},
		{name: "Remote later", localUpdated: GetDate(0), localSynced: GetDate(0), remoteUpdated: GetDate(1), want: RemoteNewer},
		{name: "Both later, local newest", localUpdated: GetDate(5), localSynced: GetDate(0), remoteUpdated: GetDate(2), want: ConflictLocalNewer},
		{name: "Both later, remote newest", localUpdated: GetDate(2), localSynced: GetDate(0), remoteUpdated: GetDate(5), want: ConflictRemoteNewer},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			local := FileInfo{
				Updated:    tt.localUpdated,
				LastSynced: tt.localSynced,
			}
			got, err := local.Compare(FileInfo{
				Updated: tt.remoteUpdated,
			}, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileInfo.CompareForPull() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileInfo.CompareForPull() = %v, want %v", got, tt.want)
			}
		})
	}
}
