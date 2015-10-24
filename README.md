carchivum
=========
[![Build Status](https://travis-ci.org/mohae/carchivum.png)](https://travis-ci.org/mohae/carchivum)

## About

Carchivum is a package for working with compressed archives. Archivum is latin for archives. Carchivum is not an application, just a package that other applications can use to add compressed archive support.

These compressed archives can be in either `tar` or `zip` format.

## Supported archival formats
### Default: tar
Tar, tape archive, is the default archive format that carchivum uses. Carchivum does not support all of the compression formats that `tar` does. It may, at some point, support compression formats that tar does not.  Tar archives generated by Carchivum may not be extractable by `tar`; compatibility depends on the algorithm used for compression.  If compatibility with `tar` is desired, make sure the compression format being used is one that `tar` supports.

Carchivum's default compression format for tarballs is gzip.

__In the future, the carchivum archives may be more than a tar, which will make `.car` files incompatible with tar. This will probably be implemented in a manner that continues to support the tar format, but no gurantees. If those does occur, a flag will be added for `tar` compatibility. This flag will not guarantee that `tar` will be able to extract a `.car` file as this will also depend on the compression algorithm used. It will guarantee that the archive is created as a `tar`.__

### zip
The zip format includes compression and its standard extension is `.zip`. No other compression schemes are supported. 

## Supported Compression Algorithms
Carchivum supports a number of compression algorithms. More may be implemented in the future. Carchivum does not support all of the compression algorithms that `tar` does. Carchivum does support some compression algorithms that `tar` does not. If compatibility with `tar` is important to you, make sure that the compression algorithm used is supported by `tar`. By default, Carchivum uses `gzip` for compression; this is compatible with `tar`.

* bzip2
* gzip
* lzw
* lz4

Currently there is no support for specifying the compression level, the defaults compression levels are used.

### Options

```
format       string          The archive format to use. (DEFAULT=tar)
type         string          The compression type to use. This is only used
                             when the archive format is tar. (DEFAULT=gzip)
usefullpath  bool            If files should be archives using their fullpath
                             or relative paths. (DEFAULT=false)
```

## Adding `carchivum` to your application

    import github.com/mohae/carchivum

For an example implementation, please see [car](https://github.com/mohae/car), my cross-platform CLI tool for creating `car` files.

## Enable logging
This package uses the standard log package and logs all errors as `log.Print()` or `log.Printf()` prior to returning.

It may be that this package stops logging directly and only returns the error, relying on the caller to do any logging it deems appropriate. This is a design decision I have punted, for now.

## Functionality wishlist

* Add support for exclude filters.
* Add support for include filters.
* Add support for archiving since a date.
* Add support for archiving using relative datetime.
