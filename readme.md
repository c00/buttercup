# buttercup

> Sync me up, butterup

Backup local folders somewhere remote. But with privacy and security in mind.

All your files are locally encrypted before being sent off. Files can be synced to any s3-compatible storage (AWS S3, Digital Ocean Object Storage, etc.), or just to some local folder like an external harddrive or a NAS.

**Don't forget your password, as you will not be able to recover your files without it!**

# Usage

```bash
# Pull changes for your default folder from its remote.
buttercup pull

# Pull changes for a named source folder from its remote.
buttercup pull [source_name]
```

# Todo

- [x] Expand commands to allow syncing of other folders
- [ ] Implement S3 provider
- [ ] Keep permissions the same
- [x] Are we dealing with deleted files correctly? Pushing and pulling
- [ ] check code coverage for glaring holes
- [ ] Page indexes so we don't pull potentially millions of files into memory
- [x] Loglevels so we can be more or less verbose
