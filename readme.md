# buttercup

> Sync me up, butterup

Backup local folders somewhere remote. But with privacy and security in mind. Has the ability to client-side encrypt your data before shipping it off to some cloud storage provider. Without the passphrase nobody will be able to see your data. (That includes you, so don't lose your passphrase...)

All your files are locally encrypted before being sent off. Files can be synced to any s3-compatible storage (AWS S3, Digital Ocean Object Storage, etc.), or just to some local folder like an external harddrive or a NAS.

**Don't forget your passphrase, as you will not be able to recover your files without it!**

# Current state

Very very alpha. Some stuff needs to happen before it's even usable.

# Usage

```bash
# Show help
buttercup help

# Show version
buttercup version

# Sync your default folder with its remote
buttercup sync

# Sync a named source folder with its remote
buttercup sync [source_name]

# Pull changes for your default folder from its remote.
buttercup pull

# Pull changes for a named source folder from its remote.
buttercup pull [source_name]

# Push changes from your default folder to its remote.
buttercup push

# Push changes from a named source folder to its remote.
buttercup push [source_name]
```

# Todo

- [x] Expand commands to allow syncing of other folders
- [ ] Implement S3 provider
- [ ] Keep permissions the same
- [x] Are we dealing with deleted files correctly? Pushing and pulling
- [ ] check code coverage for glaring holes
- [ ] Page indexes so we don't pull potentially millions of files into memory
- [x] Loglevels so we can be more or less verbose
- [ ] Some ssetup for new uses
- [ ] Some Service / Monitoring for automatic syncing

# Config file

Until the setup commands are added, here's an example yaml configuration to put in `~/.buttercup/config.yaml`:

```yaml
# Default configuration to run commands on if no folder name is given
defaultFolder: default
# Make sure that all your clients have a different name. Used to determine who has the write lock.
clientName: testies
folders:
  - name: default
    local:
      type: filesystem
      fsConfig:
        path: /home/someuser/Buttercup
    remote:
      type: encrypted-filesystem
      efsConfig:
        path: /media/somedevice/encrypted-buttercup-backups
        passphrase: fBDVLCC+Qbf6j22qfjhCsPQnSA7hf+UkPJg3+D6erCA=
  - name: alt
    local:
      type: filesystem
      fsConfig:
        path: /home/someuser/secret-docs
    remote:
      type: encrypted-filesystem
      efsConfig:
        path: /media/somedevice/encrypted-docs
        passphrase: 80khe0B18aFdmW+gWnfcJL61DoNPgDX5zrHGe3w6A68=
```
