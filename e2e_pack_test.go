package onearchiver_test

import (
	. "github.com/ganeshrvel/one-archiver"
	"github.com/ganeshrvel/yeka_zip"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func _testListingPackedArchive(_metaObj *ArchiveMeta, assertionArr []string, excludeSizes map[string]int) {
	Convey("recursive=true | Asc - it should not throw an error", func() {
		_listObj := NewArchiveRead()
		_listObj.ListDirectoryPath = ""
		_listObj.Recursive = true
		_listObj.OrderBy = OrderByFullPath
		_listObj.OrderDir = OrderDirAsc

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var itemsArr []string

		for _, item := range result {
			itemsArr = append(itemsArr, item.FullPath)

			if !item.IsDir {

				if _, exists := excludeSizes[item.FullPath]; exists {
					continue
				}

				So(item.Size, ShouldBeGreaterThan, 0)
			}

		}

		So(itemsArr, ShouldResemble, assertionArr)

	})
}

func _testPackedArchiveAfterUnpacking(t *testing.T, _metaObj *ArchiveMeta, contentsAssertionArr []map[string][]byte, session *Session, passwords []string) {

	_destination := newTempMocksDir("mock_test_file1", true)

	Convey("it should not throw an error", func() {
		unpackObj := NewArchiveUnpack()
		unpackObj.FileList = []string{}
		unpackObj.Destination = _destination
		unpackObj.Passwords = passwords

		err := StartUnpacking(_metaObj, unpackObj, session)

		So(err, ShouldBeNil)

		Convey("Read the extracted directory  - it should not throw an error", func() {
			if len(contentsAssertionArr) == 0 {
				Convey("The assertion array should not be empty", func() {
					t.Fatal("contentsAssertionArr is empty")
				})
				return
			}

			if len(contentsAssertionArr) > 0 {
				contentsArr := make([]map[string][]byte, len(contentsAssertionArr))

				for idx, m := range contentsAssertionArr {
					for key := range m {
						contentsArr[idx] = map[string][]byte{key: nil}
					}
				}

				getContentsUnpackedDirectory(_destination, &contentsArr)

				Convey("Read contents and match them  - it should not throw an error", func() {
					So(contentsArr, ShouldResemble, contentsAssertionArr)
				})
			}
		})
	})
}

func _testPacking(t *testing.T, _metaObj *ArchiveMeta, session *Session, password string, zipEncryptionMethod zip.EncryptionMethod) {

	Convey("gitIgnorePattern | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		_metaObj.GitIgnorePattern = []string{"b.txt"}

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)
	})

	Convey("Single path in 'fileList' | selected - a directory  | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionArr, map[string]int{})

			Convey("Unpack and test archive", func() {
				contentsAssertionArr := []map[string][]byte{
					{"mock_dir1/a.txt": []byte("abc d efg")},
					{"mock_dir1/1/": nil},
					{"mock_dir1/1/a.txt": []byte("abcdefg\n")},
					{"mock_dir1/2/": nil},
					{"mock_dir1/2/b.txt": []byte("123456")},
					{"mock_dir1/3/": nil},
					{"mock_dir1/3/b.txt": []byte("123456")},
					{"mock_dir1/3/2/": nil},
					{"mock_dir1/3/2/b.txt": []byte("123456")},
				}

				_testPackedArchiveAfterUnpacking(t, _metaObj, contentsAssertionArr, session, []string{password})
			})

		})
	})

	Convey("Single path in 'fileList' | selected - a file  | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/a.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"a.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - Multiple directories | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1")
		path2 := getTestMocksAsset("mock_dir2")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt", "mock_dir2/", "mock_dir2/a.txt", "mock_dir2/1/", "mock_dir2/1/a.txt", "mock_dir2/2/", "mock_dir2/2/b.txt", "mock_dir2/3/", "mock_dir2/3/b.txt", "mock_dir2/3/2/", "mock_dir2/3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - Multiple files | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir3/a.txt")
		path2 := getTestMocksAsset("mock_dir3/b.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"a.txt", "b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - single file and a single directory | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/a.txt")
		path2 := getTestMocksAsset("mock_dir1/1/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"a.txt", "1/", "1/a.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - single file and multiple directories | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/a.txt")
		path2 := getTestMocksAsset("mock_dir1/1/")
		path3 := getTestMocksAsset("mock_dir1/2/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2, path3}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"a.txt", "1/", "1/a.txt", "2/", "2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - multiple files and multiple directories | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir3/a.txt")
		path2 := getTestMocksAsset("mock_dir3/b.txt")
		path3 := getTestMocksAsset("mock_dir3/1/")
		path4 := getTestMocksAsset("mock_dir3/2/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2, path3, path4}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"a.txt", "b.txt", "1/", "1/a.txt", "2/", "2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Different levels of parent paths | Multiple paths in 'fileList' | selected - two files | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/a.txt")
		path2 := getTestMocksAsset("mock_dir1/1/a.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"a.txt", "1/", "1/a.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Different levels of parent paths | Multiple paths in 'fileList' | selected - 1 file and 1 directory | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/a.txt")
		path2 := getTestMocksAsset("mock_dir2/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir2/", "mock_dir2/a.txt", "mock_dir2/1/", "mock_dir2/1/a.txt", "mock_dir2/2/", "mock_dir2/2/b.txt", "mock_dir2/3/", "mock_dir2/3/b.txt", "mock_dir2/3/2/", "mock_dir2/3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Different levels of parent paths | Multiple paths in 'fileList' | selected - 2 directories - 1 | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/1")
		path2 := getTestMocksAsset("mock_dir2/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"mock_dir1/", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir2/", "mock_dir2/a.txt", "mock_dir2/1/", "mock_dir2/1/a.txt", "mock_dir2/2/", "mock_dir2/2/b.txt", "mock_dir2/3/", "mock_dir2/3/b.txt", "mock_dir2/3/2/", "mock_dir2/3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Different levels of parent paths | Multiple paths in 'fileList' | selected - 2 directories - 2 | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/1")
		path2 := getTestMocksAsset("mock_dir3/dir_1/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"mock_dir1/", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir3/", "mock_dir3/dir_1/", "mock_dir3/dir_1/a.txt", "mock_dir3/dir_1/1/", "mock_dir3/dir_1/1/a.txt", "mock_dir3/dir_1/2/", "mock_dir3/dir_1/2/b.txt", "mock_dir3/dir_1/3/", "mock_dir3/dir_1/3/b.txt", "mock_dir3/dir_1/3/2/", "mock_dir3/dir_1/3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Different levels of parent paths | Multiple paths in 'fileList' | selected - 2 files and 2 directories | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/1")
		path2 := getTestMocksAsset("mock_dir2/3/2/b.txt")
		path3 := getTestMocksAsset("mock_dir3/dir_1/1/a.txt")
		path4 := getTestMocksAsset("mock_dir3/dir_1/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2, path3, path4}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"mock_dir1/", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir2/", "mock_dir2/3/", "mock_dir2/3/2/", "mock_dir2/3/2/b.txt", "mock_dir3/", "mock_dir3/dir_1/", "mock_dir3/dir_1/a.txt", "mock_dir3/dir_1/1/", "mock_dir3/dir_1/1/a.txt", "mock_dir3/dir_1/2/", "mock_dir3/dir_1/2/b.txt", "mock_dir3/dir_1/3/", "mock_dir3/dir_1/3/b.txt", "mock_dir3/dir_1/3/2/", "mock_dir3/dir_1/3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - same 2 files and 2 directories | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/1")
		path2 := getTestMocksAsset("mock_dir1/1")
		path3 := getTestMocksAsset("mock_dir2/3/2/b.txt")
		path4 := getTestMocksAsset("mock_dir2/3/2/b.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2, path3, path4}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"mock_dir1/", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir2/", "mock_dir2/3/", "mock_dir2/3/2/", "mock_dir2/3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - 1 file and 1 directory from the same nested parent | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir2/3/")
		path2 := getTestMocksAsset("mock_dir2/3/2/b.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"b.txt", "2/", "2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - 1 file and 1 directory from the same nested parent - 2 | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir2/3/2/b.txt")
		path2 := getTestMocksAsset("mock_dir2/3/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"b.txt", "2/", "2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - same file multiple times | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir3/b.txt")
		path2 := getTestMocksAsset("mock_dir3/b.txt")
		path3 := getTestMocksAsset("mock_dir3/b.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2, path3}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("Multiple paths in 'fileList' | selected - same directory multiple times | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir3/")
		path2 := getTestMocksAsset("mock_dir3/")
		path3 := getTestMocksAsset("mock_dir3/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1, path2, path3}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionList := []string{"a.txt", "b.txt", "1/", "1/a.txt", "2/", "2/b.txt", "dir_1/", "dir_1/a.txt", "dir_1/1/", "dir_1/1/a.txt", "dir_1/2/", "dir_1/2/b.txt", "dir_1/3/", "dir_1/3/b.txt", "dir_1/3/2/", "dir_1/3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionList, map[string]int{})
		})
	})

	Convey("no files in the pathlist |d It should not throw an error", func() {
		_packObj := NewArchivePack()
		_packObj.FileList = []string{}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionArr := []string(nil)

			_testListingPackedArchive(_metaObj, assertionArr, map[string]int{})
		})
	})

	Convey("symlink | directory | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir4/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionArr := []string{"a.txt", "b.txt", "1/", "1/a.txt", "2/", "2/b.txt", "3/", "3/b.txt", "3/2/", "3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionArr, map[string]int{"b.txt": 0})
		})
	})

	Convey("symlink mock_dir5 | directory with bad and good symlink | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir5/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}
		_packObj.Password = password
		_packObj.ZipEncryptionMethod = zipEncryptionMethod

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed Archive files", func() {
			assertionArr := []string{"a.txt", "b.txt", "cc.txt", "1/", "1/a.txt", "2/", "2/b.txt", "3/", "3/b.txt", "3/2/", "3/2/b.txt"}

			_testListingPackedArchive(_metaObj, assertionArr, map[string]int{"b.txt": 0, "cc.txt": 0})
		})
	})

	Convey("Unpack and test archive", func() {

		contentsAssertionArr := []map[string][]byte{
			{"a.txt": []byte("abc d efg")},
			{"1/": nil},
			{"1/a.txt": []byte("abcdefg\n")},
			{"2/": nil},
			{"2/b.txt": []byte("123456")},
			{"3/": nil},
			{"3/b.txt": []byte("123456")},
			{"3/2/": nil},
			{"3/2/b.txt": []byte("123456")},
		}

		_testPackedArchiveAfterUnpacking(t, _metaObj, contentsAssertionArr, session, []string{password})
	})
}

func _testCompressedFilePacking(t *testing.T, _metaObj *ArchiveMeta, session *Session, packedFileName string) {

	Convey("Single path in 'fileList' | selected - a file  | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/a.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed compressed file", func() {
			assertionArr := []string{packedFileName}

			_testListingPackedArchive(_metaObj, assertionArr, map[string]int{packedFileName: 0})

		})

		Convey("Unpack the compressed file and test", func() {

			contentsAssertionArr := []map[string][]byte{
				{packedFileName: []byte("abc d efg")},
			}

			_testPackedArchiveAfterUnpacking(t, _metaObj, contentsAssertionArr, session, []string{})
		})
	})

	Convey("gitIgnorePattern | It should NOT throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/a.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}

		_metaObj.GitIgnorePattern = []string{"b.txt"}

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed compressed file", func() {
			assertionArr := []string{packedFileName}

			_testListingPackedArchive(_metaObj, assertionArr, map[string]int{packedFileName: 0})
		})
	})

	Convey("gitIgnorePattern | It should throw an error", func() {
		path1 := getTestMocksAsset("mock_dir1/2/b.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}

		_metaObj.GitIgnorePattern = []string{"b.txt"}

		err := StartPacking(_metaObj, _packObj, session)
		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "atleast a single file is required for creating a compress file")
	})

	Convey("no files in the pathlist | It should throw an error", func() {
		_packObj := NewArchivePack()
		_packObj.FileList = []string{}

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "atleast a single file is required for creating a compress file")
	})

	Convey("Multiple paths in 'fileList' | It should throw an error", func() {
		path1 := getTestMocksAsset("mock_dir3/")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "only a single file be packed to a compressed file, no directories are allowed")
	})

	Convey("Multiple paths in 'fileList' | selected - 2 files | It should throw an error", func() {
		path2 := getTestMocksAsset("mock_dir2/3/2/b.txt")
		path3 := getTestMocksAsset("mock_dir3/dir_1/1/a.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path2, path3}

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "only a single file can be packed to a compressed file")
	})

	Convey("symlink | b.txt | It should not throw an error", func() {
		path1 := getTestMocksAsset("mock_dir5/b.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeNil)

		Convey("List Packed compressed files", func() {
			assertionArr := []string{packedFileName}

			_testListingPackedArchive(_metaObj, assertionArr, map[string]int{"b.txt": 0, packedFileName: 0})
		})
	})

	Convey("symlink | cc.txt | It should throw an error", func() {
		path1 := getTestMocksAsset("mock_dir5/cc.txt")
		_packObj := NewArchivePack()
		_packObj.FileList = []string{path1}

		err := StartPacking(_metaObj, _packObj, session)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "no such file or directory")
	})
}

func TestPacking(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping 'TestPacking' testing in short mode")
	//}

	ph := &ProgressFunc{
		OnReceived: func(progress *Progress) {
			//fmt.Printf("received: %v\n", progress)
		},
		OnEnded: func(progress *Progress) {
			//elapsed := time.Since(progress.StartTime)

			//fmt.Printf("Time taken to create the archive: %s", elapsed)
		},
	}

	session := NewSession("", ph)

	Convey("Packing | No encryption - ZIP", t, func() {
		filename := newTempMocksAsset("arc_test_pack.zip")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing | Encrypted - ZIP (StandardEncryption)", t, func() {
		filename := newTempMocksAsset("arc_test_stdenc_pack.zip")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "1234567", zip.StandardEncryption)
	})

	Convey("Packing | Encrypted - ZIP (AES128Encryption)", t, func() {
		filename := newTempMocksAsset("arc_test_aes128enc_pack.zip")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "1234567", zip.AES128Encryption)
	})

	Convey("Packing | Encrypted - ZIP (AES256Encryption)", t, func() {
		filename := newTempMocksAsset("arc_test_aes256enc_pack.zip")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "1234567", zip.AES256Encryption)
	})

	Convey("Packing | Encrypted - ZIP (AES192Encryption)", t, func() {
		filename := newTempMocksAsset("arc_test_aes192enc_pack.zip")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "1234567", zip.AES192Encryption)
	})

	Convey("Packing | Tar", t, func() {
		filename := newTempMocksAsset("arc_test_pack.tar")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing | Tar.gz", t, func() {
		filename := newTempMocksAsset("arc_test_pack.tar.gz")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing | Tar.bz2", t, func() {
		filename := newTempMocksAsset("arc_test_pack.tar.bz2")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing | Tar.br (brotli)", t, func() {
		filename := newTempMocksAsset("arc_test_pack.tar.br")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing | Tar.lz4", t, func() {
		filename := newTempMocksAsset("arc_test_pack.tar.lz4")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing | Tar.sz", t, func() {
		filename := newTempMocksAsset("arc_test_pack.tar.sz")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing | Tar.xz", t, func() {
		filename := newTempMocksAsset("arc_test_pack.tar.xz")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing | Tar.zst (zstd)", t, func() {
		filename := newTempMocksAsset("arc_test_pack.tar.zst")

		_metaObj := NewArchiveMeta(filename)

		_testPacking(t, _metaObj, session, "", zip.StandardEncryption)
	})

	Convey("Packing compressed file | GZ", t, func() {
		filename := newTempMocksAsset("arc_test_pack.gz")

		_metaObj := NewArchiveMeta(filename)

		_testCompressedFilePacking(t, _metaObj, session, "arc_test_pack")
	})
	Convey("Packing compressed file | GZ", t, func() {
		filename := newTempMocksAsset("arc_test.a.txt.gz")

		_metaObj := NewArchiveMeta(filename)

		_testCompressedFilePacking(t, _metaObj, session, "arc_test.a.txt")
	})

	Convey("Packing compressed file | Zstd", t, func() {
		filename := newTempMocksAsset("arc_test_pack.zst")
		_metaObj := NewArchiveMeta(filename)
		_testCompressedFilePacking(t, _metaObj, session, "arc_test_pack")
	})

	Convey("Packing compressed file | Xz", t, func() {
		filename := newTempMocksAsset("arc_test_pack.xz")
		_metaObj := NewArchiveMeta(filename)
		_testCompressedFilePacking(t, _metaObj, session, "arc_test_pack")
	})

	Convey("Packing compressed file | sz (Snappy)", t, func() {
		filename := newTempMocksAsset("arc_test_pack.sz")
		_metaObj := NewArchiveMeta(filename)
		_testCompressedFilePacking(t, _metaObj, session, "arc_test_pack")
	})

	Convey("Packing compressed file | Lz4", t, func() {
		filename := newTempMocksAsset("arc_test_pack.lz4")
		_metaObj := NewArchiveMeta(filename)
		_testCompressedFilePacking(t, _metaObj, session, "arc_test_pack")
	})

	Convey("Packing compressed file | Bz2", t, func() {
		filename := newTempMocksAsset("arc_test_pack.bz2")
		_metaObj := NewArchiveMeta(filename)
		_testCompressedFilePacking(t, _metaObj, session, "arc_test_pack")
	})

	Convey("Packing compressed file | BR (Brotli)", t, func() {
		filename := newTempMocksAsset("arc_test_pack.br")
		_metaObj := NewArchiveMeta(filename)
		_testCompressedFilePacking(t, _metaObj, session, "arc_test_pack")
	})
}
