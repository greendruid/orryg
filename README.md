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

Installation
------------

For now you need to have Go installed.

    go get github.com/vrischmann/orryg

Usage
-----

Right now it works with a configuration file.

In the future I'll make a web interface so it's more user-friendly and this will disappear.

```json
{
    "checkFrequency": "1m",
    "dateFormat": "2006-02-01_030405",
    "copiers": [{
        "type": "scp",
        "conf": {
            "user": "sphax",
            "host": "192.168.1.34",
            "port": 22,
            "privateKeyFile": "N:/backup_dev.rsa",
            "backupsDir": "/tmp/backups"
        }
    }],
    "directories": [{
        "frequency": "12h",
        "origPath": "N:/Projects",
        "archiveName": "projects"
    }]
}
```
