# buttercup

> Sync me up, butterup

Backup local folders somewhere remote. But with privacy and security in mind. Buttercup let's you client-side encrypt your data before shipping it off to some cloud storage provider. Without the passphrase nobody will be able to see your data. (That includes you!)

All your files are locally encrypted before being sent off. Files can be synced to any s3-compatible storage (AWS S3, Digital Ocean Spaces, etc.), or just to some local folder like an external harddrive or a NAS.

**Don't forget your passphrase, as you will not be able to recover your files without it!**

# Current state: Beta

Use at your own risk. If you lose data it's not my fault.

Feel free to report issues, feature requests or PRs.

It has been tested on Linux. I expect it to work just fine on macOs, but I haven't tried. Windows... who knows, it's windows.

# Features

- Sync multiple devices
- Client-side encryption (like, actually private)
- Works with any s3-compatible cloud provider

# Installation

tl;dr:

```bash
make install
```

See [install guide](./guides/installation.md).

# Basic Usage

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

- [ ] Create tests for pushing when locked by someone else
- [ ] Keep permissions the same
- [ ] check code coverage for glaring holes
- [ ] Page indexes so we don't pull potentially millions of files into memory
- [ ] Some setup for new users
- [ ] Some Service / Monitoring for automatic syncing
- [ ] Make password optional so you get asked every time
- [ ] For the local folders, store index somewhere else.
- [ ] Add command to reset a local or remote
      Reset local is just delete the index
      Reset remote is delete the entire fucking thing and push
- [ ] Add command to force remove lock
- [ ] Add command to push/pull individual files
- [ ] Add command to ls remote

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
      # Backup to s3
      type: s3
      s3Config:
        # Passphrase for encryption / decryption.
        passphrase: somelongpassphrasethatsreallysecure
        # S3 Access Key
        accessKey: youraccesskey
        # S3 Secret Key
        secretKey: yoursecretkey
        # S3 Bucket name
        bucket: bucketname
        # Optionally a path within the bucket
        basePath: my-buttercup-backups
        # The endpoint (without the bucket name)
        endpoint: https://sgp1.digitaloceanspaces.com
        # Should be false for Digital Ocean. Your mileage may vary with other providers
        forcePathStyle: false
        # S3 Region
        region: sgp1
  - name: alt
    local:
      type: filesystem
      fsConfig:
        path: /home/someuser/secret-docs
    remote:
      # Backup to the local filesystem, use encryption
      type: encrypted-filesystem
      efsConfig:
        path: /media/somedevice/encrypted-docs
        # Passphrase for encryption / decryption.
        passphrase: somelongpassphrasethatsreallysecure
```
