orryg
=====

Orryg is an automated backup tool. The goal is to define a set of directories to backup and a set of destinations to backup to.

Orryg is smart: it remembers the last time a backup was made for a directory, and will retrigger a backup at the appropriate time, even if you stopped your computer.

support
-------

Only Windows is supported for now. This is my primary OS, so that's what I support.

Supporting another OS is probably not hard, but it requires splitting up properly the Windows service stuff and the rest.
That will be done later.

destinations support
--------------------

Done:

  * SSH

Planned:

  * Amazon S3
  * OneDrive
  * Google Drive

Maybe:

  * FTP
  * Copy to local Dropbox directory

configuration
-------------

Configuration is done via a JSON file that you have to write by hand. It kinda sucks, but hey, you'll manage.

This is the reference example configuration:

    {
        "scpCopiers": [
            {
                "name": "myserver",
                "params": {
                    "user": "myuser",
                    "host": "myserver.example.com",
                    "port": 22,
                    "privateKeyFile": "",
                    "backupsDir": "/home/myuser/backups"
                }
            }
        ],
        "directories": [
            {
                "frequency": "6h0m0s",
                "originalPath": "C:/Users/myuser/Documents",
                "maxBackups": 10,
                "archiveName": "my_documents"
            },
            {
                "frequency": "6h0m0s",
                "originalPath": "C:/Users/myuser/Pictures",
                "maxBackupAge": "72h",
                "archiveName": "my_pictures"
            }
        ],
        "checkFrequency": "10m0s",
        "cleanupFrequency": "10m0s",
        "dateFormat": "20060201_150405"
    }

There are some parameters that are not immediately obvious:

  * *maxBackups* is the maximum number of backups for this directory on each copier. Orryg removes the oldest backups first.
  * *maxBackupAge* is the oldest a backup can be. If you set this to say 24h, Orryg will remove everything that is older than that.
  * *checkFrequency* is the interval at which Orryg checks if there are out of date directories.
  * *cleanupFrequency* is the interval at which Orryg checks if there are expired backups to delete.
  * *dateFormat* is the format used for the date in the archive name in the backups. It is not a real date.
    * It's based on the date 01/02/2006 15:05:05. You can use any format you want as long as you use the correct numbers.
    * If you have a frequency less than a day, make sure your format includes the hour. Same with minutes.

installation of the service
---------------------------

    orryg.exe install

Now the service should appear in the *Services* app and you can start it.

You can also start and stop it from the command line:

    orryg.exe start
    orryg.exe stop

where are the logs
------------------

By running *Event Viewer* you can follow the execution of Orryg, although the UI is terrible, it works.
There might be a real log file in the future.
