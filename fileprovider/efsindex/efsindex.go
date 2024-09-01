package efsindex

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/c00/buttercup/modifiers"
	_ "github.com/mattn/go-sqlite3"
)

type EfsFileInfo struct {
	Path          string
	LastSynced    time.Time
	Updated       time.Time
	Deleted       bool
	StoredPath    string
	TrackingValue int64
}

func New(path string, passphrase string) *EfsIndex {
	return &EfsIndex{encryptedPath: path, passphrase: passphrase}
}

type EfsIndex struct {
	encryptedPath   string
	unencryptedPath string
	passphrase      string
	db              *sql.DB
}

func (i *EfsIndex) MarkDeleted(trackingValue int64) error {
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
func (i *EfsIndex) DeleteFileInfo(path string) error {
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

func (i *EfsIndex) GetFileInfo(path string) (EfsFileInfo, error) {
	err := i.Load()
	if err != nil {
		return EfsFileInfo{}, err
	}

	row := i.db.QueryRow(`SELECT path, lastsynced, updated, deleted, storedpath, trackingvalue FROM fileinfo WHERE path = ?`, path)
	fi := EfsFileInfo{}
	err = row.Scan(&fi.Path, &fi.LastSynced, &fi.Updated, &fi.Deleted, &fi.StoredPath, &fi.TrackingValue)
	if err != nil {
		return EfsFileInfo{}, fmt.Errorf("error querying database: %w", err)
	}

	return fi, nil
}

func (i *EfsIndex) SetFileInfo(fi EfsFileInfo) error {
	err := i.Load()
	if err != nil {
		return err
	}

	_, err = i.db.Exec(
		`INSERT INTO fileinfo (path, lastsynced, updated, deleted, storedpath, trackingvalue)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(Path) DO UPDATE SET 
			lastsynced = excluded.lastsynced,
			updated = excluded.updated,
			deleted = excluded.deleted,
			storedpath = excluded.storedpath,
			trackingvalue = excluded.trackingvalue;`,
		fi.Path, fi.LastSynced, fi.Updated, fi.Deleted, fi.StoredPath, fi.TrackingValue,
	)
	if err != nil {
		return fmt.Errorf("cannot insert: %w", err)
	}

	return nil
}

// todo add a updatePath statement
func (i *EfsIndex) UpdatePath(oldPath, newPath string) error {
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

func (i *EfsIndex) GetPage(offset, limit int) ([]EfsFileInfo, error) {
	err := i.Load()
	if err != nil {
		return nil, err
	}

	if limit == 0 {
		limit = -1
	}

	sql := "SELECT path, lastsynced, updated, deleted, storedpath, trackingvalue FROM fileinfo LIMIT ?"
	values := []any{limit}

	if offset > 0 {
		sql += " OFFSET ?"
		values = append(values, offset)
	}

	rows, err := i.db.Query(sql, values...)
	if err != nil {
		return nil, fmt.Errorf("could not get rows: %w", err)
	}

	results := []EfsFileInfo{}

	for rows.Next() {
		fi := EfsFileInfo{}
		err = rows.Scan(&fi.Path, &fi.LastSynced, &fi.Updated, &fi.Deleted, &fi.StoredPath, &fi.TrackingValue)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		results = append(results, fi)
	}

	return results, nil
}

func (i *EfsIndex) Close() error {
	if i.db == nil {
		return nil
	}

	err := i.db.Close()
	if err != nil {
		return fmt.Errorf("cannot close db: %w", err)
	}

	i.db = nil

	//encrypt
	err = modifiers.CompressAndEncryptFile(i.unencryptedPath, i.encryptedPath, i.passphrase)
	if err != nil {
		return fmt.Errorf("cannot encrypt index: %w", err)
	}

	//delete unencrypted index
	err = os.Remove(i.unencryptedPath)
	if err != nil {
		return fmt.Errorf("cannot cleanup unencrypted index: %w", err)
	}

	return nil
}

func (i *EfsIndex) Load() error {
	if i.db != nil {
		return nil
	}

	if i.unencryptedPath == "" {
		randBytes := make([]byte, 16)
		_, err := rand.Read(randBytes)
		if err != nil {
			return fmt.Errorf("cannot get random bytes: %w", err)
		}

		randStr := hex.EncodeToString(randBytes)
		i.unencryptedPath = path.Join(os.TempDir(), "buttercup-"+randStr+".db")
	}

	_, err := os.Stat(i.encryptedPath)
	if err == nil {
		//Decrypt the db
		err := modifiers.DecryptAndDecompressFile(i.encryptedPath, i.unencryptedPath, i.passphrase)
		if err != nil {
			return fmt.Errorf("error decrypting index file: %w", err)
		}
	}

	//create connection
	conn, err := sql.Open("sqlite3", i.unencryptedPath)
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
