package onearchiver

import (
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TODO symlink and hardlink
func _testListingUnpackedArchive(metaObj *ArchiveMeta, unpackObj *ArchiveUnpack, archiveFilesAssertionArr []string, directoryFilesAssertionArr []string) {
	destination := unpackObj.Destination

	Convey("recursive=true | Asc - it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
		}

		result, err := GetArchiveFileList(metaObj, _listObj)

		So(err, ShouldBeNil)

		var itemsArr []string

		for _, item := range result {
			itemsArr = append(itemsArr, item.FullPath)
		}

		So(itemsArr, ShouldResemble, archiveFilesAssertionArr)
	})

	Convey("Read the extracted directory  - it should not throw an error", func() {
		var filePathList []filePathListSortInfo

		err := filepath.Walk(destination, func(path string, info os.FileInfo, err error) error {
			if destination == path {
				return nil
			}

			var pathSplitted [2]string

			if !info.IsDir() {
				pathSplitted = [2]string{filepath.Dir(path), filepath.Base(path)}
			} else {
				path = fixDirSlash(true, path)
				_dir := filepath.Dir(path)

				pathSplitted = [2]string{_dir, ""}
			}

			filePathList = append(filePathList, filePathListSortInfo{
				IsDir:         info.IsDir(),
				FullPath:      path,
				splittedPaths: pathSplitted,
			})

			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		_sortPath(&filePathList, OrderDirAsc)

		var itemsArr []string

		for _, x := range filePathList {
			_path := strings.Replace(x.FullPath, destination, "", -1)
			_path = strings.TrimLeft(_path, "/")

			itemsArr = append(itemsArr, _path)
		}

		So(itemsArr, ShouldResemble, directoryFilesAssertionArr)
	})
}

func _testUnpacking(metaObj *ArchiveMeta, ph *ProgressHandler) {
	Convey("Warm up test | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
		}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, assertionArr, assertionArr)
		})
	})

	Convey("gitIgnore | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{"a.txt"}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			assertionArr := []string{"mock_dir1/", "mock_dir1/1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, assertionArr, assertionArr)
		})
	})

	Convey("fileList | 1 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})

	Convey("fileList | 2 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})

	Convey("fileList | 3 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})

	Convey("fileList | 4 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/a.txt"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})

	Convey("fileList | 5 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1", "mock_dir1/a.txt"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, archiveFilesAssertionArr)
		})
	})

	Convey("fileList | 6 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt", "mock_dir1/a.txt"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})

	Convey("fileList | 7 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt", "mock_dir1/3/"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})

	Convey("fileList | 8 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2", "mock_dir1/3"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})

	Convey("fileList | 9 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt", "mock_dir1/3/b.txt"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})

	Convey("gitIgnorePattern | fileList | 1 | It should not throw an error", func() {
		_destination := newTempMocksDir("mock_test_file1", true)

		unpackObj := &ArchiveUnpack{
			FileList:    []string{"mock_dir1/2/b.txt", "mock_dir1/3/b.txt"},
			Destination: _destination,
		}

		metaObj.GitIgnorePattern = []string{"mock_dir1/3"}

		err := StartUnpacking(metaObj, unpackObj, ph)

		So(err, ShouldBeNil)

		Convey("List the archive files", func() {
			archiveFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			directoryFilesAssertionArr := []string{"mock_dir1/", "mock_dir1/2/", "mock_dir1/2/b.txt"}

			_testListingUnpackedArchive(metaObj, unpackObj, archiveFilesAssertionArr, directoryFilesAssertionArr)
		})
	})
}

func TestUnpacking(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping 'TestUnpacking' testing in short mode")
	//}

	ph := ProgressHandler{
		onReceived: func(pInfo *ProgressInfo) {
			//fmt.Printf("received: %v\n", pInfo)
		},
		onError: func(err error, pInfo *ProgressInfo) {
			//fmt.Printf("error: %e\n", err)
		},
		onCompleted: func(pInfo *ProgressInfo) {
			//elapsed := time.Since(pInfo.StartTime)
			//
			//fmt.Println("observable is closed")
			//fmt.Printf("Time taken to unpack the archive: %s", elapsed)
		},
	}

	Convey("Unpacking | No encryption - ZIP", t, func() {
		filename := getTestMocksAsset("mock_test_file1.zip")

		metaObj := &ArchiveMeta{Filename: filename, Password: ""}

		_testUnpacking(metaObj, &ph)
	})

	Convey("Unpacking | Encryption - ZIP", t, func() {
		filename := getTestMocksAsset("mock_enc_test_file1.zip")

		metaObj := &ArchiveMeta{Filename: filename, Password: "1234567"}

		_testUnpacking(metaObj, &ph)
	})

	Convey("Unpacking | Tar", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpacking(_metaObj, &ph)
	})

	Convey("Unpacking | Tar.gz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.gz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpacking(_metaObj, &ph)
	})

	Convey("Unpacking | Tar.bz2", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.bz2")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpacking(_metaObj, &ph)
	})

	Convey("Unpacking | Tar.br (brotli)", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.br")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpacking(_metaObj, &ph)
	})

	Convey("Unpacking | Tar.lz4", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.lz4")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpacking(_metaObj, &ph)
	})

	Convey("Unpacking | Tar.sz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.sz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpacking(_metaObj, &ph)
	})

	Convey("Unpacking | Tar.xz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.xz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpacking(_metaObj, &ph)
	})

	Convey("Unpacking | Tar.zst (zstd)", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.zst")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testUnpacking(_metaObj, &ph)
	})
}
