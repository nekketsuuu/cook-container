## What's this?

A toy container which uses overlayfs. **This is only a TOY container!** The original source code of this container was implemented by [@rrreeeyyy and others][1] and modified by @nekketsuuu during Cookpad Spring 1-day Internship 2019.

## Prerequisites

* Linux >= kernel 3.18
    * I tested on Ubuntu 18.04.2 bionic
* Go
    * The root user must be able to run `go`.

## Usage

1. Prepare `/root/overlayfs/lower`. At least, copy `/bin/sh` to `/root/overlayfs/lower/bin`.
    Example (on Ubuntu 18.04.2):

    ```sh
    # Create root-like directories
    sudo mkdir -p /root/overlayfs/lower/bin
    sudo mkdir -p /root/overlayfs/lower/lib
    sudo ln -s /root/overlayfs/lower/lib /root/overlayfs/lower/lib64
    # Copy necessary binaries: copy binary itself and its dependencies
    sudo cp /bin/sh /root/overlayfs/lower/bin/
    ldd /bin/sh  # list dependent libraries
    sudo cp /lib/x86_64-linux-gnu/libc.so.6 /root/overlayfs/lower/lib/
    sudo cp /lib64/ld-linux-x86-64.so.2 /root/overlayfs/lower/lib/
    # (Same for /bin/ls)
    sudo cp /bin/ls /root/overlayfs/lower/bin/
    ldd /bin/ls
    sudo cp /lib/x86_64-linux-gnu/libselinux.so.1 /root/overlayfs/lower/lib/
    sudo cp /lib/x86_64-linux-gnu/libc.so.6 /root/overlayfs/lower/lib/
    sudo cp /lib/x86_64-linux-gnu/libpcre.so.3 /root/overlayfs/lower/lib/
    sudo cp /lib/x86_64-linux-gnu/libdl.so.2 /root/overlayfs/lower/lib/
    sudo cp /lib64/ld-linux-x86-64.so.2 /root/overlayfs/lower/lib/
    sudo cp /lib/x86_64-linux-gnu/libpthread.so.0 /root/overlayfs/lower/lib/
    ```

2. Run `main.go` with `run` arg: `sudo go run main.go run`. Then it will run a container.

After running a container:

```sh
$ sudo go run main.go run
# ls
bin  lib  lib64  proc
# echo foobar > test
# ls
bin  lib  lib64  proc  test
# exit
$ sudo go run main.go run
# ls  # The environment returned to the original! The file `test` was removed.
bin  lib  lib64  proc
```

## Miscellaneous notes for development

* overlayfs
    * <https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt>
    * <https://github.com/moby/moby/blob/master/daemon/graphdriver/overlay/overlay.go>
    * (in Japanese) <http://gihyo.jp/admin/serial/01/linux_containers/0018>
* mount
    * Args format of `syscall.Mount` is different for each filesystem. See `man 2 mount` and `man 8 mount`.

## License

This source code is originally implemented by [@rrreeeyyy and others][1] and modified by @nekketsuuu. Currently there is no OSS-compatible license for this repository.


  [1]: https://github.com/rrreeeyyy/container-internship
