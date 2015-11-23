orryg
=====

Orryg is an automated backup tool. The goal is to define a set of directories to backup and a set of destinations to backup to.

Orryg is smart: it remembers the last time a backup was made for a directory, and will retrigger a backup at the appropriate time, even if you stopped your computer.

support
-------

Only Windows is supported for now. This is my primary OS, so that's what I support.
Supporting another OS is probably not hard, but will be done later.

copiers support
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

On Windows, all data is stored in the registry. I'm building the UI right now, and it will feature a way to configure everything.

If you want to test right now you'll have to create the registry keys by hand.
