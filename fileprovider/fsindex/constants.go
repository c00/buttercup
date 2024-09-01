package fsindex

const createScript = `CREATE TABLE IF NOT EXISTS fileinfo (
	path TEXT PRIMARY KEY NOT NULL,
	lastsynced DATETIME NOT NULL,
	updated DATETIME NOT NULL,
	deleted BOOLEAN NOT NULL,
	trackingvalue INTEGER NULL
);`
