# Setting up a folder on system as a remote

This can be useful if you have a NAS or a USB drive you can backup to.

## Configure Buttercup

Open your configuration yaml at `~/.buttercup/config.yaml`. Update the information under the `remote` key:

```yaml
defaultFolder: default
clientName: mycomputername
folders:
  - name: default
    local:
      type: filesystem
      fsConfig:
        path: /home/myusername/Buttercup
    remote:
      type: encrypted-filesystem
      # Update stuff down here
      efsConfig:
        path: /media/some/folder/backups
        passphrase: somethingrandom
```
