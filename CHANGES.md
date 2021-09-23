# Release notes
All notable changes to this project will be documented in this file.  
This project adheres to [Semantic Versioning](http://semver.org/).

## 0.5.0
- Added `--no-resume` option to avoid resuming an existing upload. The upload will start over if an existing upload already exists.
- Improved logging for many error cases

## 0.4.0
- Upload through new upload servers.
- Updated to Go 1.15 with Go Modules.
- Updated almost all dependencies. Removed deprecated ones.

## 0.3.0
- Support overriding the chunk size uploads are split to to with a new `--chunk-size` option. Defaults to 10 MB.

## 0.2.0
- Support passing token using a `GJPUSH_TOKEN` environment variable
- Support passing token using a global credentials file located in the user's home directory.

## 0.1.0
- **Initial release!**
- Resumable build uploading and creation