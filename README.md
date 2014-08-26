carchivum
=========

Carchivum is a package for working with compressed archives. Archivum is latin for archives. Carchivum is not an application, just a package that other applications can use to add compressed archive support.

Examples of carchivum implementations:

* [baller](https://github.com/mohae/baller): is a command-line tool for creating compressed archives of targeted sources. It also supports automatic deletion of the archived files. 

## Supported archival formats
### Default: tar
Tar, tape archive, is the default archive format that carchivum uses. Carchivum does not support all of the compression formats that `tar` does. It may, at some point, support compression formats that tar does not.  If compatibility with `tar` is desired, make sure the compression format being used is one that `tar` supports.

Carchivum's default compression format for tarballs is bz2, and uses the `.bz2` extension.

### zip
The zip format includes compression and its standard extension is `.zip`. No other compression schemes are supported. 

## Supported Compression Algorithms
Carchivum supports a variety of compression algorithms, with the possibility of more being added in the futrure. At some point, carchivum may support a compression scheme that is not compatible with `tar`. It is also doubtful that carchivum will ever support all of the compression schemes that `tar` does. If you want to be able to use the created archive with `tar`, make sure the compression scheme used is one that `tar` is compatible with.

By default the bz2 compression scheme is used with tarballs with the `.bz2` extension.

* bz2
* gzip

## Installing carchivum
Assumptions: Go is installed and working.

    $ go get https://github.com/mohae/carchivum

## Adding to your application

    import (
            "fmt"

            arch "github.com/mohae/carchivum"
    )

    func main() {
            // Create a new archiver using carchivum's defaults.
            archiver := arch.NewArchiver()
     
            // Create an archive of the current directory using the current
            // directory name and datetime. The suffix will automatically be
	    // appended to.  
            err := archiver.Create("archive{{ .Datetime }}", "."
            if err != nil {
                    fmt.Println(err.Error())
                    return
            }

            // Change the comression scheme. GZip is a constant for gzip. All
            // supported compression schemes have constants, or you can just
            // pass a matching string in.
	    err = archiver.SetCompression(GZip)
            if err != nil {
                    fmt.Println(err.Error())
                    return
            }

            // Change the datetime format used. This uses Go's format; check
            // docs on specifics about valid values.
            err = archiver.SetDatetimeFormat(RFC822)
            if err != nil {
                    fmt.Println(err.Error())
                    return
            }

            // Now create the archive
            err = archiver.Create("archive{{ .Datetime ))", "." )
            if err != nil {
                    fmt.Println(err.Error())
                    return
            }


            // Delete works the same way, just use Delete() instead. 
            // Remember running delete will result in the path being deleted!
            // This will do the same as the prior command, but also delete the
            // directory. 
            //
            // RFC822's time resolution is only to minutes. If there is a 
            // collision on the filename, a random 4 digit number will be 
            // appended to the new filename, unless overwrite_on_collision is
            // set to true.
            err = archiver.Delete("archive{{ .Datetime }}", ".")
            if err !- nil {
                      fmt.Println(err.Error())
            }

## Enable logging
Logging is disabled by default. As a package, whether carchivum logs or not and at what level is not up to it, its up to you and your application. Carchivum implemnts seelog for logging and its implementation is consistent with seelog's guidelines at https://github.com/cibun/seelog/wiki/Writing-libraries-with-seelog.

## Notes:

