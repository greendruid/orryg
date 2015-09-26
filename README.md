orryg
=====

Orryg is an automated backup tool. The goal is to defined a set of directories to backup and a set of destinations to backup to.

The destinations will support:

  * SSH
  * Amazon S3
  * OneDrive
  * Google Drive
  * FTP (maybe)
  * Copy to local Dropbox directory (maybe)

Orryg is smart: it remembers the last time a backup was made for a directory, and will retrigger a backup at the appropriate time, even if you stopped your computer.

daemon dev setup
================

    export GO15VENDOREXPERIMENT=1
    go get github.com/codegangsta/gin

    gin -a 8080

web ui dev setup
================

    npm install -g typescript
    npm install -g tsd

    tsd install react --save --resolve

Run this to automatically rebuilt the web ui.

    tsc -w
