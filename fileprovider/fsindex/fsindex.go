package fsindex

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type FsFileInfo struct {
	Path          string
	LastSynced    time.Time
	Updated       time.Time
	Deleted       bool
	TrackingValue int64
}

func New(path string) *FsIndex {
	return &FsIndex{path: path}
}

type FsIndex struct {
	path string
	db   *sql.DB
}

func (i *FsIndex) MarkDeleted(trackingValue int64) error {
	err := i.Load()
	if err != nil {
		return err
	}

	sql := "UPDATE fileinfo SET deleted = 1, updated = DATETIME(lastsynced, '+1 second') WHERE trackingValue != ? AND deleted = 0"
	_, err = i.db.Exec(sql, trackingValue)
	if err != nil {
		return fmt.Errorf("could not set deleted files: %w", err)
	}

	return nil
}

// We don't use this...
func (i *FsIndex) DeleteFileInfo(path string) error {
	err := i.Load()
	if err != nil {
		return err
	}

	_, err = i.db.Exec(`DELETE FROM fileinfo WHERE path = ?`, path)
	if err != nil {
		return fmt.Errorf("error deleting fileinfo: %w", err)
	}

	return nil
}

func (i *FsIndex) GetFileInfo(path string) (FsFileInfo, error) {
	err := i.Load()
	if err != nil {
		return FsFileInfo{}, err
	}

	row := i.db.QueryRow(`SELECT path, lastsynced, updated, deleted, trackingvalue FROM fileinfo WHERE path = ?`, path)
	fi := FsFileInfo{}
	err = row.Scan(&fi.Path, &fi.LastSynced, &fi.Updated, &fi.Deleted, &fi.TrackingValue)
	if err != nil {
		return FsFileInfo{}, fmt.Errorf("error querying database: %w", err)
	}

	return fi, nil
}

func (i *FsIndex) SetFileInfo(fi FsFileInfo) error {
	err := i.Load()
	if err != nil {
		return err
	}

	_, err = i.db.Exec(
		`INSERT INTO fileinfo (path, lastsynced, updated, deleted, trackingvalue)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(Path) DO UPDATE SET 
			lastsynced = excluded.lastsynced,
			updated = excluded.updated,
			deleted = excluded.deleted,
			trackingvalue = excluded.trackingvalue;`,
		fi.Path, fi.LastSynced, fi.Updated, fi.Deleted, fi.TrackingValue,
	)
	if err != nil {
		return fmt.Errorf("cannot insert: %w", err)
	}

	return nil
}

// todo add a updatePath statement
func (i *FsIndex) UpdatePath(oldPath, newPath string) error {
	err := i.Load()
	if err != nil {
		return err
	}

	sql := "UPDATE fileinfo SET path = ? WHERE path = ?"
	res, err := i.db.Exec(sql, newPath, oldPath)
	if err != nil {
		return fmt.Errorf("could not update path: %w", err)
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return errors.New("no rows updated")
	}
	return nil
}

func (i *FsIndex) GetPage(offset, limit int) ([]FsFileInfo, error) {
	err := i.Load()
	if err != nil {
		return nil, err
	}

	if limit == 0 {
		limit = -1
	}

	sql := "SELECT path, lastsynced, updated, deleted, trackingvalue FROM fileinfo LIMIT ?"
	values := []any{limit}

	if offset > 0 {
		sql += " OFFSET ?"
		values = append(values, offset)
	}

	rows, err := i.db.Query(sql, values...)
	if err != nil {
		return nil, fmt.Errorf("could not get rows: %w", err)
	}

	results := []FsFileInfo{}

	for rows.Next() {
		fi := FsFileInfo{}
		err = rows.Scan(&fi.Path, &fi.LastSynced, &fi.Updated, &fi.Deleted, &fi.TrackingValue)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		results = append(results, fi)
	}

	return results, nil
}

func (i *FsIndex) Close() error {
	if i.db == nil {
		return nil
	}

	err := i.db.Close()
	if err != nil {
		return fmt.Errorf("cannot close db: %w", err)
	}

	i.db = nil
	return nil
}

func (i *FsIndex) Load() error {
	if i.db != nil {
		return nil
	}

	//create connection
	conn, err := sql.Open("sqlite3", i.path)
	if err != nil {
		return fmt.Errorf("cannot connect to sqlite db: %w", err)
	}
	i.db = conn

	_, err = conn.Exec(createScript)
	if err != nil {
		return fmt.Errorf("cannot run create script: %w", err)
	}
	return nil
}
