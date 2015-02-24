carchivum
=========

## Status
Under development. Initial implementation will be complete when [car](https://github.com/mohae/car)'s initial implementation is complete.

## About

Carchivum is a package for working with compressed archives. Archivum is latin for archives. Carchivum is not an application, just a package that other applications can use to add compressed archive support.

Examples of carchivum implementations:


## Supported archival formats
### Default: tar
Tar, tape archive, is the default archive format that carchivum uses. Carchivum does not support all of the compression formats that `tar` does. It may, at some point, support compression formats that tar does not.  If compatibility with `tar` is desired, make sure the compression format being used is one that `tar` supports.

Carchivum's default compression format for tarballs is gzip.

__In the future, the carchivum archives may be more than a tar, which will make `.car` files incompatible with tar. This will probably be implemented in a manner that continues to support the tar format, but no gurantees.__

### zip
The zip format includes compression and its standard extension is `.zip`. No other compression schemes are supported. 

## Supported Compression Algorithms
Carchivum only supports the `gzip` compression algorithm for tarballs. Support for additional compression types may be added. At some point, carchivum may support a compression scheme that is not compatible with `tar`. It is also doubtful that carchivum will ever support all of the compression schemes that `tar` does. If you want to be able to use the created archive with `tar`, make sure the compression scheme used is one that `tar` is compatible with.

* gzip

### Options

```
format       string          The archive format to use. (DEFAULT=tar)
type         string          The compression type to use. This is only used
                             when the archive format is tar. (DEFAULT=gzip)
usefullpath  bool            If files should be archives using their fullpath
                             or relative paths. (DEFAULT=false)
exclude	     string array    A comma separated list of files or extensions to
                             exclude from the archive. exclude is mutually
                             exclusive with include.
include      string array    A comma separated list of files or extensions to
                             include with the archive. include is mutually
                             exclusive with exclude; include takes precedence.
since        string          Only archive files that are either new or have
                             been modified since the value specified.
relative     string          Only archive files that are either new or have
                             been modified in the timeframe relative to now.
```

## Adding `carchivum` to your application


## Enable logging
This package uses the standard log package and logs all errors as `log.Print()` or `log.Printf()` prior to returning.

It may be that this package stops logging directly and only returns the error, relying on the caller to do any logging it deems appropriate. This is a design decision I have punted, for now.

## Functionality wishlist

* Create a gzip compressed tar file from a list of sources and write it to a destination. (COMPLETE)
* Create a zip file from a list of sources and writ it to a destination. (COMPLETE)
* Support using relative or fullpaths for added files. (COMPLETE)
* Extract a zip file.
* Extract a gzip compressed tar file.
* Archive command: create a gzip compressed tar file from the source and delete the source.
	compressed filename will be sourceDir-timeFormat.tgz for tarballs.
* Archive command: create a zip compressed file from the source and delete the source.
	compressed filename will be sourceDir-timeFormat.zip.
* Add support for exclude filters.
* Add support for include filters.
* Add support for archiving since a date.
* Add support for archiving using relative datetime.
