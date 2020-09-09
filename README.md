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

	if exist := fileExists(filename); !exist {
		fmt.Printf("file does not exist: %s\n", filename)

		return
	}

	_metaObj := &ArchiveMeta{
		Filename: filename,
		GitIgnorePattern: []string{},
	}

	_listObj := &ArchiveRead{
		ListDirectoryPath: "test-directory/",
		Recursive:         true,
		OrderBy:           OrderByName,
		OrderDir:          OrderDirAsc,
	}

	result, err := GetArchiveFileList(_metaObj, _listObj)

	if err != nil {
		fmt.Printf("An error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", result)
```


**Is encrypted**

```shell script
	filename := "test.enc.zip"
	//filename := "test.enc.rar"

	if exist := fileExists(filename); !exist {
		fmt.Printf("file does not exist %s\n", filename)

		return
	}

	_metaObj := &ArchiveMeta{
	  Filename: filename,
      Password: "",
    }

	result, err := isArchiveEncrypted(_metaObj)

	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result; isEncrypted: %v, isValidPassword: %v\n", result.isEncrypted, result.isValidPassword)
```



**Pack**

```shell script
func Pack() {
	filename := "/path/pack.zip"
	path1 := "directory1"
	path2 := "directory2"

	_metaObj := &ArchiveMeta{
		Filename:         filename,
		GitIgnorePattern: []string{"git.log"},
		Password:         "",
		EncryptionMethod: zip.StandardEncryption,
	}

	_packObj := &ArchivePack{
		FileList: []string{path1, path2},
	}

	ph := ProgressHandler{
		onReceived: func(pInfo *ProgressInfo) {
			fmt.Printf("received: %v\n", pInfo)
		},
		onError: func(err error, pInfo *ProgressInfo) {
			fmt.Printf("error: %e\n", err)
		},
		onCompleted: func(pInfo *ProgressInfo) {
			elapsed := time.Since(pInfo.startTime)

			fmt.Println("observable is closed")
			fmt.Printf("Time taken to create the archive: %s", elapsed)
		},
	}

	err := startPacking(_metaObj, _packObj, &ph)
	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", "Success")
}
```


**Unpack**

```shell script
func Unpack() {
	filename := "/path/pack.zip"
	destination := "arc_test_pack/"

	_metaObj := &ArchiveMeta{
		Filename:         filename,
		Password:         "",
		GitIgnorePattern: []string{},
	}

	_unpackObj := &ArchiveUnpack{
		FileList:    []string{}, // archive specific files in the directory
		Destination: destination,
	}

    ph := ProgressHandler{
		onReceived: func(pInfo *ProgressInfo) {
			fmt.Printf("received: %v\n", pInfo)
		},
		onError: func(err error, pInfo *ProgressInfo) {
			fmt.Printf("error: %e\n", err)
		},
		onCompleted: func(pInfo *ProgressInfo) {
			elapsed := time.Since(pInfo.startTime)

			fmt.Println("observable is closed")
			fmt.Printf("Time taken to unpack the archive: %s", elapsed)
		},
	}

	err := startUnpacking(_metaObj, _unpackObj, &ph)
	if err != nil {
		fmt.Printf("Error occured: %+v\n", err)

		return
	}

	fmt.Printf("Result: %+v\n", "Success")
}
```


### Credits
mholt/archiver (https://github.com/mholt/archiver)
yeka/zip (https://github.com/yeka/zip)
