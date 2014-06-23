upload-server
=============

A quick-and-dirty server in Go that accepts and saves file uploads.

###Usage

`upload-server [-log [log-file-path]] [-addr [listen-address]] [-out [output-dir-path]]`

By default, nothing is logged, the server listens on :80, and uploaded files are placed in the current working directory.
