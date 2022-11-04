## Concurrent Download Manager

A downloader built in go that uses multithreading for fast downloading of large files.

Steps to run:

1. Clone the git repository.
2. cd to the directory and run "go build".

Usage:

`./go-cdm -t <number of threads>`

## NOTE:

1. Max threads default to 30.
2. Potential errors may arise for files that do not support part downloading. (It is still a TODO)
