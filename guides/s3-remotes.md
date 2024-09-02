# Setup S3 remote

Note that whether or not you trust the s3-provider to not spy on you, is not relevant. The data is encrypted before it ever goes to them. Even if they would leak the data publically, nobody would be able to use it without your passphrase.

Buttercup should be able to sync with any S3-compatible service. So far I've only used Digital Ocean. If you do try it with other providers, a PR would be very welcome to add to this guide.

## Digital Ocean Spaces

In these steps make sure to note the following:

- The endpoint url
- The bucket name
- The Access Key
- The Secret key

### Creating the bucket

1. Create a Digital Ocean account ([referral link](https://m.do.co/c/2c4575a18d8a)) if you don't have one. Log in to the dashboard if you do.
2. Go to [Spaces](https://cloud.digitalocean.com/spaces)
3. Go to "Create Spaces Bucket"
4. Choose a data center region, note down the region (e.g. `sgp1`)
5. Set a name for your bucket. e.g. `buttercup-joe`
6. Leave the rest at the default values and click "Create a Spaces Bucket" at the bottom
7. Once created, note down the "origin endpoint", but remove the bucket name from the endpoint. (e.g. `https://buttercup-joe.sgp1.digitaloceanspaces.com` becomes `https://sgp1.digitaloceanspaces.com`)

### Creating the Access Key and Secret Key

1. On the left menu go down to `API`.
2. Go to "Spaces Keys"
3. Click "Generate New Key"
4. Enter name. e.g. `buttercup-key`
5. Note down the Access Key and Secret Key. (Note that you will not be able to view the secret key again, so make sure to copy it somewhere.)

### Configure Buttercup

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
      type: s3
      # Update stuff down here
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

### Sync!

Now all you have to do is sync it:

```bash
buttercup sync
```
