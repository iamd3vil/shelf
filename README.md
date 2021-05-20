![Alt text](logo.png "Shelf Logo")

A Good Symlinks Manager. Swap files from any location with symlinks and manage them easily in one place

An approach to manage symlinks that just works out of the box! No more fighting with junk configs/systems to manage your dotfiles across multiple systems.

## Usage

```
NAME:
   shelf - A new cli application

USAGE:
   shelf.bin [global options] command [command options] [arguments...]

DESCRIPTION:
   A Good Symlinks Manager

COMMANDS:
   create, c    creates a Shelf
   track, t     track a file
   clone, cl    clones a shelf
   snapshot, s  creates a snapshot of existing shelves
   restore, r   restores all the links from a shelf
   where, w     prints where the given shelf is
   list, ls     lists all the files tracked by shelf
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

## Tutorial

### Creating a shelf

`shelf` can be used to create multiple shleves. A `shelf` is a collection of files you want to track together.

You can create a shelf using `shelf create` command.

```bash
$ shelf create dotfiles
```

This creates a shelf under `$HOME/.shelves/dotfiles`.

A shelf is just a directory with all the tracked files. Shelfs can be managed using either git or any other backup system.

Also a `shelf.json` is created inside the shelf which keeps track of all the files and their symlink paths. This files needs to be maintained with other files.

### Tracking a file

To start tracking/adding a file to a shelf:
```bash
$ shelf track dotfiles ~/.gitconfig
```

This moves the file from `~/.gitconfig` to `~/.shelves/dotfiles/.gitconfig` and creates a symlink at `~/.gitconfig`.

To move to a shelf directory(for example for running git commands or any other backup commands):

```bash
$ cd `shelf where dotfiles`
```

### Snapshot a shelf

A shelf can be `snapshotted` using `shelf` CLI itself. There are 2 modes to take a snapshot:

#### Using git

To keep a track of your `shelf` using `.git`, you can use

```bash
$ shelf snapshot git dotfiles
```

This will add all the existing files in your shelf directory, create an automated `commit` and push the commit to
the remote branch.

#### Using archive

To use any other backup mechanism, `shelf` comes with a utility to create a `.tar.gz` archive for your shelf.

```bash
$ shelf snapshot archive --output /data/backup/shelves dotfiles
```

This will create an archive file of your shelf in `/data/backup/shelves/dotfiles.tar.gz` and you can use any backup tool
to preserve this data for long term.

### Restoring a shelf

A shelf can be cloned directly into the local system if the shelf has git repo. For cloning:

```bash
$ shelf clone https://github.com/iamd3vil/dotfiles.git
```

After cloning, for restoring all the symlinks:

```bash
$ shelf restore dotfiles
```

If there is already a file/symlink in the restoring path, it will print a warning and skips the file.

For example, if only a single file needs to be restored, the restore command can be run again and the warnings for the other files can be ignored.