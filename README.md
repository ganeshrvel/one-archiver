# One Archiver
All-in-one archiver package for Go.

### Supported archive formats
- zip
- tar
- tar.br (brotli)
- tar.bz2 (bz2)
- tar.gz (gzip)
- tar.lz4
- tar.sz (snappy)
- tar.xz
- tar.zst (zstd)
- rar (read-only)

### Format-dependent features
- Create/read/extract an encrypted zip file
- Read/extract an encrypted rar file
- List a specific directory in an archive
- Sort and list files by size, time, name, path
- Extract specific files from an archive
- Gitignore patterns for easy skipping files/directories
- Emits progress while archiving and unarchiving
- Check whether a zip or rar file is encrypted
- Check whether the archive password is correct
- Gzip is multithreaded
- Make all necessary directories
- Open password-protected RAR archives


### Using the library
```shell script
go get github.com/ganeshrvel/one-archiver
```

Install Go mholt package (https://github.com/mholt/archiver/issues/195)
```shell script
cd $GOPATH
go get github.com/pierrec/lz4 && cd $GOPATH/src/github.com/pierrec/lz4 && git fetch && git checkout v3.0.1
```

### APIs
**List an archive**

```shell script
filename := "test.zip"

if exist := one_archiver.FileExists(filename); !exist {
    fmt.Printf("file does not exist: %s\n", filename)

    return
}

am := &one_archiver.ArchiveMeta{
    Filename:         filename,
    Password:         "",
    GitIgnorePattern: []string{},
}

ar := &one_archiver.ArchiveRead{
    ListDirectoryPath: "test-directory/",
    Recursive:         true,
    OrderBy:           one_archiver.OrderByName,
    OrderDir:          one_archiver.OrderDirAsc,
}

result, err := one_archiver.GetArchiveFileList(am, ar)

if err != nil {
    fmt.Printf("Error occured: %+v\n", err)

    return
}

fmt.Printf("Result: %+v\n", result)
```


**Is encrypted**

```shell script
filename := "test.enc.zip"
//filename := "test.encrypted.rar"

if exist := one_archiver.FileExists(filename); !exist {
    fmt.Printf("file does not exist %s\n", filename)

    return
}

am := &one_archiver.ArchiveMeta{
    Filename: filename,
    Password: "1234567",
}

result, err := one_archiver.IsArchiveEncrypted(am)

if err != nil {
    fmt.Printf("Error occured: %+v\n", err)

    return
}

fmt.Printf("Result; IsEncrypted: %v, IsValidPassword: %v\n", result.IsEncrypted, result.IsValidPassword)
```



**Pack**

```shell script
import (
	"fmt"
	"github.com/yeka/zip"
	"time"
)

filename := "/path/pack.zip"
path1 := "directory1"
path2 := "directory2"

am := &one_archiver.ArchiveMeta{
    Filename:         filename,
    GitIgnorePattern: []string{},
    Password:         "",
    EncryptionMethod: zip.StandardEncryption,
}

ap := &one_archiver.ArchivePack{
    FileList: []string{path1, path2},
}

ph := &one_archiver.ProgressHandler{
    onReceived: func(pInfo *one_archiver.ProgressInfo) {
        fmt.Printf("received: %v\n", pInfo)
    },
    onError: func(err error, pInfo *one_archiver.ProgressInfo) {
        fmt.Printf("error: %e\n", err)
    },
    onCompleted: func(pInfo *one_archiver.ProgressInfo) {
        elapsed := time.Since(pInfo.StartTime)

        fmt.Println("observable is closed")
        fmt.Printf("Time taken to create the archive: %s", elapsed)
    },
}

err := one_archiver.StartPacking(am, ap, ph)
if err != nil {
    fmt.Printf("Error occured: %+v\n", err)

    return
}

fmt.Printf("Result: %+v\n", "Success")
```


**Unpack**

```shell script
import (
	"fmt"
	"github.com/yeka/zip"
	"time"
)

filename := "/path/pack.zip"
destination := "arc_test_pack/"

am := &one_archiver.ArchiveMeta{
    Filename:         filename,
    Password:         "",
    GitIgnorePattern: []string{},
}

au := &one_archiver.ArchiveUnpack{
    FileList:    []string{},
    Destination: tempDir,
}

ph := &one_archiver.ProgressHandler{
    onReceived: func(pInfo *ProgressInfo) {
        fmt.Printf("received: %v\n", pInfo)
    },
    onError: func(err error, pInfo *one_archiver.ProgressInfo) {
        fmt.Printf("error: %e\n", err)
    },
    onCompleted: func(pInfo *one_archiver.ProgressInfo) {
        elapsed := time.Since(pInfo.StartTime)

        fmt.Println("observable is closed")
        fmt.Printf("Time taken to unpack the archive: %s", elapsed)
    },
}

err := one_archiver.StartUnpacking(am, au, ph)
if err != nil {
    fmt.Printf("Error occured: %+v\n", err)

    return
}

fmt.Printf("Result: %+v\n", "Success")

```


### Credits
mholt/archiver (https://github.com/mholt/archiver)
yeka/zip (https://github.com/yeka/zip)
