# Installation Guide

## Build and install the binary

To install, the best thing is to compile from source. For this you will need to have `go` [installed](https://go.dev/doc/install) and have the `[go_install_folder]/go/bin` folder in your path.

Then, clone the repo and run:

```bash
make install
```

Check that the binary is installed correctly by running: `buttercup version`.

## Initialize configuration

Initialize a new configuration file:

```bash
buttercup init
```

This creates the configuration file at `~/.buttercup/config.yaml`. You should open the file and set your keys accordingly.

## Setting up S3 remote

See [this guide](./s3-remotes.md). Buttercup should be able to sync with any S3-compatible service.

## Setting up a local folder as a remote

See [this guide](./encrypted-filesystem.md).

## Usage

See [usage guide](./usage.md).
