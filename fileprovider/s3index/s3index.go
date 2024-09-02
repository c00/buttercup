package s3index

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/c00/buttercup/fileprovider/s3client"
	"github.com/c00/buttercup/modifiers"
	_ "github.com/mattn/go-sqlite3"
)

const sqliteIndexName = ".buttercup-index.db"

type S3FileInfo struct {
	Path          string
	LastSynced    time.Time
	Updated       time.Time
	Deleted       bool
	StoredPath    string
	TrackingValue int64
}

func New(s3client *s3client.S3Client, passphrase string) *S3Index {
	return &S3Index{s3client: s3client, passphrase: passphrase}
}

type S3Index struct {
	unencryptedPath string
	passphrase      string
	db              *sql.DB
	s3client        *s3client.S3Client
}

func (i *S3Index) MarkDeleted(trackingValue int64) error {
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
func (i *S3Index) DeleteFileInfo(path string) error {
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

func (i *S3Index) GetFileInfo(path string) (S3FileInfo, error) {
	err := i.Load()
	if err != nil {
		return S3FileInfo{}, err
	}

	row := i.db.QueryRow(`SELECT path, lastsynced, updated, deleted, storedpath, trackingvalue FROM fileinfo WHERE path = ?`, path)
	fi := S3FileInfo{}
	err = row.Scan(&fi.Path, &fi.LastSynced, &fi.Updated, &fi.Deleted, &fi.StoredPath, &fi.TrackingValue)
	if err != nil {
		return S3FileInfo{}, fmt.Errorf("error querying database: %w", err)
	}

	return fi, nil
}

func (i *S3Index) SetFileInfo(fi S3FileInfo) error {
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
func (i *S3Index) UpdatePath(oldPath, newPath string) error {
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

func (i *S3Index) GetPage(offset, limit int) ([]S3FileInfo, error) {
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

	results := []S3FileInfo{}

	for rows.Next() {
		fi := S3FileInfo{}
		err = rows.Scan(&fi.Path, &fi.LastSynced, &fi.Updated, &fi.Deleted, &fi.StoredPath, &fi.TrackingValue)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		results = append(results, fi)
	}

	return results, nil
}

func (i *S3Index) Close() error {
	if i.db == nil {
		return nil
	}

	err := i.db.Close()
	if err != nil {
		return fmt.Errorf("cannot close db: %w", err)
	}

	i.db = nil

	tmpFile, err := os.CreateTemp("", "enc-")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	defer tmpFile.Close()

	//encrypt
	err = modifiers.CompressAndEncryptFile(i.unencryptedPath, tmpFile.Name(), i.passphrase)
	if err != nil {
		return fmt.Errorf("cannot encrypt index: %w", err)
	}

	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("could not reset head to 0: %w", err)
	}

	//push to s3
	err = i.s3client.UploadFile(sqliteIndexName, tmpFile)
	if err != nil {
		return fmt.Errorf("cannot upload index: %w", err)
	}

	//delete unencrypted index
	err = os.Remove(i.unencryptedPath)
	if err != nil {
		return fmt.Errorf("cannot cleanup unencrypted index: %w", err)
	}

	return nil
}

func (i *S3Index) Load() error {
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

	encryptedData, err := i.s3client.DownloadFile(sqliteIndexName)
	if err == nil {
		defer encryptedData.Close()
		decryptedData, err := modifiers.DecryptAndDecompress(encryptedData, i.passphrase)
		if err != nil {
			return fmt.Errorf("cannot decrypt index: %w", err)
		}
		defer decryptedData.Close()

		file, err := os.OpenFile(i.unencryptedPath, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return fmt.Errorf("cannot open unencrypted index: %w", err)
		}

		_, err = io.Copy(file, decryptedData)
		if err != nil {
			return fmt.Errorf("could not write index to file: %w", err)
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
