# rollcage

[![Build Status](https://travis-ci.org/cactus/rollcage.png?branch=master)](https://travis-ci.org/cactus/rollcage)

## About

![status](.tools/alpha.png)

Proof of concept pseudo-reimplementation of [iocage][1] in Go, with intent to
focus on "thickjails".

[currently implemented commands][3]

## Usage

Create a config file at `/usr/local/etc/rollcage.conf`:

    # zfsroot is the location of the zfs dataset where jails live
    zfsroot = "tank/iocage"

Example layout:

    $ zfs list -r -d 1 tank/iocage
    NAME                   USED  AVAIL  REFER  MOUNTPOINT
    tank/iocage            9.01G   436G   112K  /iocage
    tank/iocage/.defaults    96K   436G    96K  /iocage/.defaults
    tank/iocage/download     96K   436G    96K  /iocage/download
    tank/iocage/jails      8.74G   436G    96K  /iocage/jails
    tank/iocage/log         128K   436G   128K  /iocage/log
    tank/iocage/releases    276M   436G    96K  /iocage/releases
    

## Commands

The current list of supported commands is as follows:

    chroot      Chroot into jail, without actually starting the jail itself
    console     Execute login to have a shell inside the jail.
    destroy     destroy jail
    df          List disk space related information
    exec        Execute login to have a shell inside the jail.
    get         get list of properties
    list        List all jails
    reboot      reboot jail
    restart     restart jail
    rollback    Rollback jail to a particular snapshot
    runtime     show runtime configuration of a jail
    set         Set a property to a given value
    snaplist    List all snapshots belonging to jail
    snapremove  Remove snapshots belonging to jail
    snapshot    Create a zfs snapshot for jail
    start       start jail
    stop        stop jail
    update      update a jail to the latest patchlevel
    version     Print the version

## Building

Currently required to build:

*   a working Go (1.5 recommended) install
*   [gb][2]
*   make (bsdmake, not gnu-make)

Building:

    $ make
    Available targets:
      help                this help
      clean               clean up
      all                 build binaries and man pages
      test                run tests
      cover               run tests with cover output
      build               build all binaries
      man                 build all man pages

    $ make build
    Restoring deps...
    Building rollcage...
    ...

    $ bin/rollcage version
    rollcage no-version (go1.5.1,gc-amd64)


[1]: https://github.com/iocage/iocage
[2]: http://getgb.io
[3]: https://gist.github.com/cactus/542d14aa96e86355ce7d
