# Usage Guide

This guide assumes you have installed and configured your buttercup installation as explained [here](./installation.md).

## Synchronizing

To sync your changes to the remote, simply run `buttercup sync`.

Syncing will first pull any new files on the remote to your local folder, and then push any locally changed files up to the remote.

### Conflicts

If a file has been changed both locally and remotely since they were last synced, then there is conflict. Conflicts are handled while pulling changed from the remote. If a conflict exists, the newest file will be kept, and the other file will be renamed to `[originalfilename].conflict.[extension]`. After that it is just considered another file on the system.

## Pushing and pulling

To only push or pull, use the command `buttercup pull` and `buttercup push`. If you try to push before pulling you will get an error when there are new changes remotely that you don't have locally yet.

## Connecting a new device to an existing remote

To connect a new device to an existing remote, the easiest thing is to just use the same remote configuration. When running the sync command it will simply pull down everything to your new device, and you'll be ready to go.

Make sure you choose a different `clientName` on the second device.

Example:

```yaml
# Configuration on device 1
defaultFolder: default
clientName: donkey # This name should be different on each device
folders:
  - name: default
    local:
      type: filesystem
      fsConfig:
        path: /home/myusername/Buttercup
    remote:
      type: s3
      # This stuff below should be the same on both devices
      s3Config:
        region: yourregion
        bucket: yourbucketname
        endpoint: yourendpoint
        accessKey: youraccesskey
        secretKey: yoursecretkey
        passphrase: somethingrandom
        basePath: my-buttercup-folder
        forcePathStyle: false

# Configuration on device 2
defaultFolder: default
clientName: shrek # This name should be different on each device
folders:
  - name: default
    local:
      type: filesystem
      fsConfig:
        path: /home/myusername/Buttercup
    remote:
      type: s3
      # This stuff below should be the same on both devices
      s3Config:
        region: yourregion
        bucket: yourbucketname
        endpoint: yourendpoint
        accessKey: youraccesskey
        secretKey: yoursecretkey
        passphrase: somethingrandom
        basePath: my-buttercup-folder
        forcePathStyle: false
```
