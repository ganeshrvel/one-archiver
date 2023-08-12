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

```go
filename := "test.zip"

if exist := onearchiver.FileExists(filename); !exist {
    fmt.Printf("file does not exist: %s\n", filename)

    return
}

am := &onearchiver.ArchiveMeta{
    Filename:         filename,
    Password:         "",
    GitIgnorePattern: []string{},
}

ar := &onearchiver.ArchiveRead{
    ListDirectoryPath: "test-directory/",
    Recursive:         true,
    OrderBy:           onearchiver.OrderByName,
    OrderDir:          onearchiver.OrderDirAsc,
}

result, err := onearchiver.GetArchiveFileList(am, ar)

if err != nil {
    fmt.Printf("Error occured: %+v\n", err)

    return
}

fmt.Printf("Result: %+v\n", result)
```


**Is encrypted**

```go
filename := "test.enc.zip"
//filename := "test.encrypted.rar"

if exist := onearchiver.FileExists(filename); !exist {
    fmt.Printf("file does not exist %s\n", filename)

    return
}

am := &onearchiver.ArchiveMeta{
    Filename: filename,
    Password: "1234567",
}

result, err := onearchiver.IsArchiveEncrypted(am)

if err != nil {
    fmt.Printf("Error occured: %+v\n", err)

    return
}

fmt.Printf("Result; IsEncrypted: %v, IsValidPassword: %v\n", result.IsEncrypted, result.IsValidPassword)
```



**Pack**

```go
import (
	"fmt"
	"github.com/ganeshrvel/yeka_zip"
	"time"
)

filename := "/path/pack.zip"
path1 := "directory1"
path2 := "directory2"

am := &onearchiver.ArchiveMeta{
    Filename:         filename,
    GitIgnorePattern: []string{},
    Password:         "",
    EncryptionMethod: zip.StandardEncryption,
}

ap := &onearchiver.ArchivePack{
    FileList: []string{path1, path2},
}

ph := &onearchiver.ProgressHandler{
    OnReceived: func(pInfo *onearchiver.ProgressInfo) {
        fmt.Printf("received: %v\n", pInfo)
    },
    OnError: func(err error, pInfo *onearchiver.ProgressInfo) {
        fmt.Printf("error: %e\n", err)
    },
    OnCompleted: func(pInfo *onearchiver.ProgressInfo) {
        elapsed := time.Since(pInfo.StartTime)

        fmt.Println("observable is closed")
        fmt.Printf("Time taken to create the archive: %s", elapsed)
    },
}

err := onearchiver.StartPacking(am, ap, ph)
if err != nil {
    fmt.Printf("Error occured: %+v\n", err)

    return
}

fmt.Printf("Result: %+v\n", "Success")
```


**Unpack**

```go
import (
	"fmt"
	"github.com/ganeshrvel/yeka_zip"
	"time"
)

filename := "/path/pack.zip"
destination := "arc_test_pack/"

am := &onearchiver.ArchiveMeta{
    Filename:         filename,
    Password:         "",
    GitIgnorePattern: []string{},
}

au := &onearchiver.ArchiveUnpack{
    FileList:    []string{},
    Destination: tempDir,
}

ph := &onearchiver.ProgressHandler{
    OnReceived: func(pInfo *onearchiver.ProgressInfo) {
        fmt.Printf("received: %v\n", pInfo)
    },
    OnError: func(err error, pInfo *onearchiver.ProgressInfo) {
        fmt.Printf("error: %e\n", err)
    },
    OnCompleted: func(pInfo *onearchiver.ProgressInfo) {
        elapsed := time.Since(pInfo.StartTime)

        fmt.Println("observable is closed")
        fmt.Printf("Time taken to unpack the archive: %s", elapsed)
    },
}

err := onearchiver.StartUnpacking(am, au, ph)
if err != nil {
    fmt.Printf("Error occured: %+v\n", err)

    return
}

fmt.Printf("Result: %+v\n", "Success")

```


### Credits
- mholt/archiver (https://github.com/mholt/archiver)
- yeka/zip (https://github.com/yeka/zip)
