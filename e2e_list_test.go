package onearchiver

import (
	"github.com/ganeshrvel/yeka_zip"
	. "github.com/smartystreets/goconvey/convey"
	"path/filepath"
	"strings"
	"testing"
)

func _testCompressedFileListing(_metaObj *ArchiveMeta, isMacOSArchive bool, destinationFilename string, passwords []string) {
	Convey("General tests", func() {
		Convey("Incorrect listDirectoryPath - it should throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "qwerty/",
				Recursive:         true,
				OrderBy:           OrderByName,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			_, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeError)
		})
	})

	Convey("gitIgnore", func() {
		Convey("gitIgnore | 1 - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "",
			}

			_metaObj.GitIgnorePattern = []string{destinationFilename}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			var assertionArr []string

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("gitIgnore | 2  - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "",
			}

			_metaObj.GitIgnorePattern = []string{"some dummy"}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{destinationFilename}

			So(itemsArr, ShouldResemble, assertionArr)
		})
	})

	Convey("ListDirectoryPath  - it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         false,
			OrderBy:           OrderByName,
			OrderDir:          OrderDirDesc,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var itemsArr []string

		for _, item := range result {
			itemsArr = append(itemsArr, item.FullPath)
		}

		assertionArr := []string{destinationFilename}

		So(itemsArr, ShouldResemble, assertionArr)
	})
}

func _testArchiveListing(_metaObj *ArchiveMeta, isMacOSArchive bool, modTimeAssertionParentPathArr []string, passwords []string) {
	Convey("General tests", func() {
		Convey("Incorrect listDirectoryPath - it should throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "qwerty/",
				Recursive:         true,
				OrderBy:           OrderByName,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			_, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeError)
		})
	})

	Convey("OrderByName", func() {
		Convey("Asc | recursive=false - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/",
				Recursive:         false,
				OrderBy:           OrderByName,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/1/", "mock_dir1/2/", "mock_dir1/3/", "mock_dir1/a.txt"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Desc | recursive=false | 1 - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/3/",
				Recursive:         false,
				OrderBy:           OrderByName,
				OrderDir:          OrderDirDesc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/3/b.txt", "mock_dir1/3/2/"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Desc | recursive=false | 2 - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/3",
				Recursive:         false,
				OrderBy:           OrderByName,
				OrderDir:          OrderDirDesc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/3/b.txt", "mock_dir1/3/2/"}

			So(itemsArr, ShouldResemble, assertionArr)
		})
	})

	Convey("gitIgnore", func() {
		Convey("gitIgnore | recursive=true | Asc | 1 - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "",
				Recursive:         true,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			_metaObj.GitIgnorePattern = []string{"a.txt"}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/", "mock_dir1/1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("gitIgnore | recursive=true | Asc | 2  - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/",
				Recursive:         true,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			_metaObj.GitIgnorePattern = []string{"a.txt"}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/1/", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			So(itemsArr, ShouldResemble, assertionArr)
		})
	})

	Convey("empty listDirectoryPath", func() {
		Convey("recursive=true | Asc - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "",
				Recursive:         true,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("recursive=false | Desc - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "",
				Recursive:         false,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirDesc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/"}

			So(itemsArr, ShouldResemble, assertionArr)
		})
	})

	Convey("OrderByFullPath", func() {
		Convey("Asc | recursive=true | 1 - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/",
				Recursive:         true,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Asc | recursive=true | 2 - it should not throw an error", func() {

			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/3",
				Recursive:         true,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Asc | recursive=false | 3 - it should not throw an error", func() {

			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/3",
				Recursive:         false,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirAsc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/3/b.txt", "mock_dir1/3/2/"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Desc | recursive=true | 1 - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/",
				Recursive:         true,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirDesc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/3/2/b.txt", "mock_dir1/3/2/", "mock_dir1/3/b.txt", "mock_dir1/3/", "mock_dir1/2/b.txt", "mock_dir1/2/", "mock_dir1/1/a.txt", "mock_dir1/1/", "mock_dir1/a.txt"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Desc | recursive=false | 2 - it should not throw an error", func() {
			_listObj := &ArchiveRead{
				ListDirectoryPath: "mock_dir1/",
				Recursive:         false,
				OrderBy:           OrderByFullPath,
				OrderDir:          OrderDirDesc,
				Passwords:         passwords,
			}

			result, err := GetArchiveFileList(_metaObj, _listObj)

			So(err, ShouldBeNil)

			var itemsArr []string

			for _, item := range result {
				itemsArr = append(itemsArr, item.FullPath)
			}

			assertionArr := []string{"mock_dir1/3/", "mock_dir1/2/", "mock_dir1/1/", "mock_dir1/a.txt"}

			So(itemsArr, ShouldResemble, assertionArr)
		})

	})

	Convey("Test Parentpath | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "mock_dir1/",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var parentPathArr []string

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			parentPathArr = append(parentPathArr, item.ParentPath)
		}

		assertionArr := []string{"mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
		assertionParentPathArr := []string{"mock_dir1/", "mock_dir1/", "mock_dir1/1/", "mock_dir1/", "mock_dir1/2/", "mock_dir1/", "mock_dir1/3/", "mock_dir1/3/", "mock_dir1/3/2/"}

		So(fullPathArr, ShouldResemble, assertionArr)
		So(parentPathArr, ShouldResemble, assertionParentPathArr)
	})

	Convey("Test Extension | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var ExtensionArr []string

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			ExtensionArr = append(ExtensionArr, item.Extension)
		}

		assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
		assertionParentPathArr := []string{"", "txt", "", "txt", "", "txt", "", "txt", "", "txt"}

		So(fullPathArr, ShouldResemble, assertionArr)
		So(ExtensionArr, ShouldResemble, assertionParentPathArr)
	})

	Convey("Test name | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "mock_dir1/",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var nameArr []string

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			nameArr = append(nameArr, item.Name)
		}

		assertionArr := []string{"mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}
		assertionNameArr := []string{"a.txt", "1", "a.txt", "2", "b.txt", "3", "b.txt", "2", "b.txt"}

		So(fullPathArr, ShouldResemble, assertionArr)
		So(nameArr, ShouldResemble, assertionNameArr)
	})

	Convey("Mode | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var modeArr []string

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			modeArr = append(modeArr, item.Mode.String())
		}

		assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

		for _, item := range modeArr {
			reg := `[drwx-]$`
			So(MatchRegex(item, reg), ShouldBeTrue)
		}

		So(len(modeArr), ShouldResemble, 10)

		So(fullPathArr, ShouldResemble, assertionArr)
	})

	Convey("Size | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var sizeArr []int64

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			sizeArr = append(sizeArr, item.Size)
		}

		var assertionParentPathArr []int64

		assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

		if !isMacOSArchive {
			assertionParentPathArr = []int64{0, 9, 0, 8, 0, 6, 0, 6, 0, 6}
		} else {
			assertionParentPathArr = []int64{0, 9, 0, 9, 0, 6, 0, 6, 0, 6}
		}

		So(fullPathArr, ShouldResemble, assertionArr)
		So(sizeArr, ShouldResemble, assertionParentPathArr)
	})

	Convey("IsDir | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var isDirArr []bool

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			isDirArr = append(isDirArr, item.IsDir)
		}

		assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

		assertionParentPathArr := []bool{true, false, true, false, true, false, true, false, true, false}

		So(fullPathArr, ShouldResemble, assertionArr)
		So(isDirArr, ShouldResemble, assertionParentPathArr)
	})

	Convey("ModTime | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var modTimeArr []string

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			modTimeArr = append(modTimeArr, item.ModTime.Format("2006"))
		}

		assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

		assertionParentPathArr := []string{"2020", "2020", "2020", "2020", "2020", "2020", "2020", "2020", "2020", "2020"}

		if modTimeAssertionParentPathArr != nil {
			assertionParentPathArr = modTimeAssertionParentPathArr
		}

		So(fullPathArr, ShouldResemble, assertionArr)
		So(modTimeArr, ShouldResemble, assertionParentPathArr)
	})

	Convey("Kind | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByKind,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var KindArr []string

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			KindArr = append(KindArr, item.Kind())
		}

		assertionParentPathArr := []string{"Folder", "Folder", "Folder", "Folder", "Folder", "Text Document", "Text Document", "Text Document", "Text Document", "Text Document"}

		So(KindArr, ShouldResemble, assertionParentPathArr)
	})
}

func _testArchiveListingInvalidPassword(_metaObj *ArchiveMeta) {
	Convey("Incorrect Password | empty password | - it should throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByName,
			OrderDir:          OrderDirAsc,
			Passwords:         []string{""},
		}

		_, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "invalid password")
	})
	Convey("Incorrect Password | with string | - it should throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByName,
			OrderDir:          OrderDirAsc,
			Passwords:         []string{"wrongpassword"},
		}

		_, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "invalid password")
	})

	Convey("Empty Password - it should throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByName,
			OrderDir:          OrderDirAsc,
			Passwords:         []string{},
		}

		_, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeError)
		So(err.Error(), ShouldContainSubstring, "password is required")
	})

	Convey("Correct Password - it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByName,
			OrderDir:          OrderDirAsc,
			Passwords:         []string{"1234567"},
		}

		_, err := GetArchiveFileList(_metaObj, _listObj)

		So(err, ShouldBeNil)
	})
}

func _testArchiveListingInvalidPasswordCommonArchivesAndZip(_metaObj *ArchiveMeta, passwords []string) {
	Convey("Incorrect Password | common archives and zip - it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByFullPath,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(_metaObj, _listObj)

		var itemsArr []string

		for _, item := range result {
			itemsArr = append(itemsArr, item.FullPath)
		}

		So(err, ShouldBeNil)
		assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

		So(itemsArr, ShouldResemble, assertionArr)
	})
}

func _testOrderByFullPathListing() {
	Convey("OrderByFullPath", func() {
		Convey("Asc | 1 - it should not throw an error", func() {
			var filePathList []filePathListSortInfo

			list := []string{"A/file1.txt",
				"A/B/C/D/file3.txt",
				"A/B/file2.txt",
				"A/B/file1.txt",
				"A/B/123.txt",
				"A/B/C/D/file1.txt",
				"A/file2.txt",
				"A/B/",
				"A/W/X/Y/Z/file1.txt",
				"A/W/file1.txt",
				"A/W/X/file1.txt",
				"A/file3.txt",
				"A/B/C/file1.txt",
				"mock_dir1/3/2/",
				"mock_dir1/3/2/b/",
				"mock_dir1/2/",
				"mock_dir1/1/",
				"mock_dir1/1/2/",
				"mock_dir1/1/a.txt",
				"mock_dir1/3/b.txt",
				"A/W/X/Y/file1.txt",
				"A/B/file2.txt", "A/file5.txt"}

			for _, x := range list {
				isDir := !strings.HasSuffix(x, ".txt")

				var pathSplitted [2]string

				if !isDir {
					pathSplitted = [2]string{filepath.Dir(x), filepath.Base(x)}
				} else {
					pathSplitted = [2]string{filepath.Dir(x), ""}
				}

				filePathList = append(filePathList, filePathListSortInfo{
					IsDir:         isDir,
					FullPath:      x,
					splittedPaths: pathSplitted,
				})
			}

			_sortPath(&filePathList, OrderDirAsc)

			assertionArr := []string{"A/file1.txt", "A/file2.txt", "A/file3.txt", "A/file5.txt", "A/B/", "A/B/123.txt", "A/B/file1.txt", "A/B/file2.txt", "A/B/file2.txt", "A/B/C/file1.txt", "A/B/C/D/file1.txt", "A/B/C/D/file3.txt", "A/W/file1.txt", "A/W/X/file1.txt", "A/W/X/Y/file1.txt", "A/W/X/Y/Z/file1.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/1/2/", "mock_dir1/2/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b/"}

			var itemsArr []string

			for _, x := range filePathList {
				itemsArr = append(itemsArr, x.FullPath)
			}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Asc | 2 - it should not throw an error", func() {
			var filePathList []filePathListSortInfo

			list := []string{"/A/file1.txt",
				"/A/B/C/D/file3.txt",
				"/A/B/file2.txt",
				"/A/B/file1.txt",
				"/A/B/123.txt",
				"/A/B/C/D/file1.txt",
				"/A/file2.txt",
				"/A/B/",
				"/A/W/X/Y/Z/file1.txt",
				"/A/W/file1.txt",
				"/A/W/X/file1.txt",
				"/A/file3.txt",
				"/A/B/C/file1.txt",
				"/mock_dir1/3/2/",
				"/mock_dir1/3/2/b/",
				"/mock_dir1/2/",
				"/mock_dir1/1/",
				"/mock_dir1/1/2/",
				"/mock_dir1/1/a.txt",
				"/mock_dir1/3/b.txt",
				"/A/W/X/Y/file1.txt",
				"/A/B/file2.txt"}

			for _, x := range list {
				isDir := !strings.HasSuffix(x, ".txt")

				var pathSplitted [2]string

				if !isDir {
					pathSplitted = [2]string{filepath.Dir(x), filepath.Base(x)}
				} else {
					pathSplitted = [2]string{filepath.Dir(x), ""}
				}

				filePathList = append(filePathList, filePathListSortInfo{
					IsDir:         isDir,
					FullPath:      x,
					splittedPaths: pathSplitted,
				})
			}

			_sortPath(&filePathList, OrderDirDesc)

			assertionArr := []string{"/mock_dir1/3/2/b/", "/mock_dir1/3/2/", "/mock_dir1/3/b.txt", "/mock_dir1/2/", "/mock_dir1/1/2/", "/mock_dir1/1/a.txt", "/mock_dir1/1/", "/A/W/X/Y/Z/file1.txt", "/A/W/X/Y/file1.txt", "/A/W/X/file1.txt", "/A/W/file1.txt", "/A/B/C/D/file3.txt", "/A/B/C/D/file1.txt", "/A/B/C/file1.txt", "/A/B/file2.txt", "/A/B/file2.txt", "/A/B/file1.txt", "/A/B/123.txt", "/A/B/", "/A/file3.txt", "/A/file2.txt", "/A/file1.txt"}

			var itemsArr []string

			for _, x := range filePathList {
				itemsArr = append(itemsArr, x.FullPath)
			}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Asc | 3 - it should not throw an error", func() {
			var filePathList []filePathListSortInfo

			list := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			for _, x := range list {
				isDir := !strings.HasSuffix(x, ".txt")

				var pathSplitted [2]string

				if !isDir {
					pathSplitted = [2]string{filepath.Dir(x), filepath.Base(x)}
				} else {
					pathSplitted = [2]string{filepath.Dir(x), ""}
				}

				filePathList = append(filePathList, filePathListSortInfo{
					IsDir:         isDir,
					FullPath:      x,
					splittedPaths: pathSplitted,
				})
			}

			_sortPath(&filePathList, OrderDirAsc)

			assertionArr := []string{"mock_dir1/", "mock_dir1/a.txt", "mock_dir1/1/", "mock_dir1/1/a.txt", "mock_dir1/2/", "mock_dir1/2/b.txt", "mock_dir1/3/", "mock_dir1/3/b.txt", "mock_dir1/3/2/", "mock_dir1/3/2/b.txt"}

			var itemsArr []string

			for _, x := range filePathList {
				itemsArr = append(itemsArr, x.FullPath)
			}

			So(itemsArr, ShouldResemble, assertionArr)
		})

		Convey("Desc | 1 - it should not throw an error", func() {
			var filePathList []filePathListSortInfo

			list := []string{"A/file1.txt",
				"A/B/C/D/file3.txt",
				"A/B/file2.txt",
				"A/B/file1.txt",
				"A/B/123.txt",
				"A/B/C/D/file1.txt",
				"A/file2.txt",
				"A/B/",
				"A/W/X/Y/Z/file1.txt",
				"A/W/file1.txt",
				"A/W/X/file1.txt",
				"A/file3.txt",
				"A/B/C/file1.txt",
				"mock_dir1/3/2/",
				"mock_dir1/3/2/b/",
				"mock_dir1/2/",
				"mock_dir1/1/",
				"mock_dir1/1/2/",
				"mock_dir1/1/a.txt",
				"mock_dir1/3/b.txt",
				"A/W/X/Y/file1.txt",
				"A/B/file2.txt"}

			for _, x := range list {
				isDir := !strings.HasSuffix(x, ".txt")

				var pathSplitted [2]string

				if !isDir {
					pathSplitted = [2]string{filepath.Dir(x), filepath.Base(x)}
				} else {
					pathSplitted = [2]string{filepath.Dir(x), ""}
				}

				filePathList = append(filePathList, filePathListSortInfo{
					IsDir:         isDir,
					FullPath:      x,
					splittedPaths: pathSplitted,
				})
			}

			_sortPath(&filePathList, OrderDirDesc)

			assertionArr := []string{"mock_dir1/3/2/b/", "mock_dir1/3/2/", "mock_dir1/3/b.txt", "mock_dir1/2/", "mock_dir1/1/2/", "mock_dir1/1/a.txt", "mock_dir1/1/", "A/W/X/Y/Z/file1.txt", "A/W/X/Y/file1.txt", "A/W/X/file1.txt", "A/W/file1.txt", "A/B/C/D/file3.txt", "A/B/C/D/file1.txt", "A/B/C/file1.txt", "A/B/file2.txt", "A/B/file2.txt", "A/B/file1.txt", "A/B/123.txt", "A/B/", "A/file3.txt", "A/file2.txt", "A/file1.txt"}

			var itemsArr []string

			for _, x := range filePathList {
				itemsArr = append(itemsArr, x.FullPath)
			}

			So(itemsArr, ShouldResemble, assertionArr)
		})
	})
}

func _testZipArchiveEncryption() {
	Convey("Non Encrypted zip | it should return false", func() {
		filename := getTestMocksAsset("mock_test_file1.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		result, err := PrepareArchive(_metaObj, []string{})

		So(err, ShouldBeNil)

		So(result.IsValidPassword, ShouldBeFalse)
		So(result.IsPasswordRequired, ShouldBeFalse)
		So(result.IsSinglePasswordMode, ShouldBeFalse)
	})

	Convey("Encrypted zip | Check if password required", func() {
		filename := getTestMocksAsset("mock_enc_test_file1.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		result, err := PrepareArchive(_metaObj, []string{})

		So(err, ShouldBeNil)

		So(result.IsValidPassword, ShouldBeFalse)
		So(result.IsPasswordRequired, ShouldBeFalse)
		So(result.IsSinglePasswordMode, ShouldBeFalse)
	})
}

func _testRarArchiveEncryption() {
	Convey("Non Encrypted rar | it should return false", func() {
		filename := getTestMocksAsset("mock_test_file1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		result, err := PrepareArchive(_metaObj, []string{})

		So(err, ShouldBeNil)

		So(result.IsValidPassword, ShouldBeFalse)
		So(result.IsPasswordRequired, ShouldBeFalse)
		So(result.IsSinglePasswordMode, ShouldBeFalse)
	})

	Convey("Encrypted rar | Check if encrypted", func() {
		filename := getTestMocksAsset("mock_enc_test_file1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		result, err := PrepareArchive(_metaObj, []string{})

		So(err, ShouldBeNil)

		So(result.IsValidPassword, ShouldBeFalse)
		So(result.IsPasswordRequired, ShouldBeTrue)
		So(result.IsSinglePasswordMode, ShouldBeTrue)
	})

	Convey("Encrypted rar | wrong password", func() {
		filename := getTestMocksAsset("mock_enc_test_file1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		result, err := PrepareArchive(_metaObj, []string{"123"})

		So(err, ShouldBeNil)

		So(result.IsValidPassword, ShouldBeFalse)
		So(result.IsPasswordRequired, ShouldBeTrue)
		So(result.IsSinglePasswordMode, ShouldBeTrue)
	})

	Convey("Encrypted rar | correct password", func() {
		filename := getTestMocksAsset("mock_enc_test_file1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		result, err := PrepareArchive(_metaObj, []string{"1234567"})

		So(err, ShouldBeNil)

		So(result.IsValidPassword, ShouldBeTrue)
		So(result.IsPasswordRequired, ShouldBeTrue)
		So(result.IsSinglePasswordMode, ShouldBeTrue)
	})
}

func TestArchiveListing(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping 'TestArchiveListing' testing in short mode")
	//}

	Convey("Testing OrderByFullPath", t, func() {
		_testOrderByFullPathListing()
	})

	Convey("macOS Compressed Archive Listing - ZIP", t, func() {
		filename := getTestMocksAsset("mock_mac_test_file1.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, true, nil, passwords)
	})

	Convey("Archive Listing - ZIP", t, func() {
		filename := getTestMocksAsset("mock_test_file1.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing - Encrypted ZIP", t, func() {
		filename := getTestMocksAsset("mock_enc_test_file1.zip")
		_metaObj := &ArchiveMeta{Filename: filename, EncryptionMethod: zip.StandardEncryption}

		_testArchiveListing(_metaObj, false, nil, []string{})
	})

	Convey("Archive Listing - tar.gz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.gz")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | 2 - tar.gz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.gz")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | Tar.br", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.br")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | Tar.bz2", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.bz2")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | Tar.lz4", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.lz4")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | Tar.sz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.sz")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | Tar.xz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.xz")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | Tar.zst", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar.zst")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | Non encrypted Rar", t, func() {
		filename := getTestMocksAsset("mock_test_file1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Archive Listing | Encrypted Rar", t, func() {
		filename := getTestMocksAsset("mock_enc_test_file1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Unpacking compressed file | GZ", t, func() {
		filename := getTestMocksAsset("mock_test_file1.zst")

		metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testCompressedFileListing(metaObj, true, "mock_test_file1", passwords)
	})
	Convey("Unpacking compressed file | GZ", t, func() {
		filename := getTestMocksAsset("mock_test_file1.a.txt.gz")

		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testCompressedFileListing(_metaObj, true, "mock_test_file1.a.txt", passwords)
	})

	Convey("Unpacking compressed file | Zstd", t, func() {
		filename := getTestMocksAsset("mock_test_file1.zst")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testCompressedFileListing(_metaObj, true, "mock_test_file1", passwords)
	})

	Convey("Unpacking compressed file | Xz", t, func() {
		filename := getTestMocksAsset("mock_test_file1.xz")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testCompressedFileListing(_metaObj, true, "mock_test_file1", passwords)
	})

	Convey("Unpacking compressed file | sz (Snappy)", t, func() {
		filename := getTestMocksAsset("mock_test_file1.sz")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testCompressedFileListing(_metaObj, true, "mock_test_file1", passwords)
	})

	Convey("Unpacking compressed file | Lz4", t, func() {
		filename := getTestMocksAsset("mock_test_file1.lz4")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testCompressedFileListing(_metaObj, true, "mock_test_file1", passwords)
	})

	Convey("Unpacking compressed file | Bz2", t, func() {
		filename := getTestMocksAsset("mock_test_file1.bz2")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testCompressedFileListing(_metaObj, true, "mock_test_file1", passwords)
	})

	Convey("Unpacking compressed file | BR (Brotli)", t, func() {
		filename := getTestMocksAsset("mock_test_file1.br")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testCompressedFileListing(_metaObj, true, "mock_test_file1", passwords)
	})
}

func TestArchiveListingPassword(t *testing.T) {
	Convey("Wrong password | Archive Listing - ZIP | it should not throw error", t, func() {
		filename := getTestMocksAsset("mock_enc_test_file1.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListingInvalidPasswordCommonArchivesAndZip(_metaObj, passwords)
	})

	Convey("Wrong password | Archive Listing - RAR", t, func() {
		filename := getTestMocksAsset("mock_enc_test_file1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		_testArchiveListingInvalidPassword(_metaObj)
	})

	Convey("Wrong password | Archive Listing - Common Archives", t, func() {
		filename := getTestMocksAsset("mock_test_file1.tar")
		_metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"wrong"}
		_testArchiveListingInvalidPasswordCommonArchivesAndZip(_metaObj, passwords)
	})
}

func TestWindowsArchiveListing(t *testing.T) {

	Convey("Windows Archive Listing - zip", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		var passwords []string
		_testArchiveListing(_metaObj, false, nil, passwords)
	})

	Convey("Windows Archive Listing - encrypted zip", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_enc.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		_testArchiveListing(_metaObj, false, nil, []string{})
	})

	Convey("Windows Archive Listing - encrypted best zip", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_enc_best.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		_testArchiveListing(_metaObj, false, nil, []string{})
	})

	Convey("Windows Archive Listing - encrypted legacy zip", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_enc_legacy.zip")
		_metaObj := &ArchiveMeta{Filename: filename}

		_testArchiveListing(_metaObj, false, nil, []string{})
	})

	Convey("Windows Archive Listing - encrypted rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_enc_best.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}
		modTimeAssertionParentPathArr := []string{"2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023"}

		_testArchiveListing(_metaObj, false, modTimeAssertionParentPathArr, passwords)
	})

	Convey("Windows Archive Listing | file names Encrypted - Rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_enc_file_names_encrypted.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}
		modTimeAssertionParentPathArr := []string{"2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023"}

		_testArchiveListing(_metaObj, false, modTimeAssertionParentPathArr, passwords)
	})

	Convey("Windows Archive Listing - rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		modTimeAssertionParentPathArr := []string{"2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023"}

		var passwords []string
		_testArchiveListing(_metaObj, false, modTimeAssertionParentPathArr, passwords)
	})

	Convey("Windows Archive Listing solid archive - rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_solid_archive_lock.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		modTimeAssertionParentPathArr := []string{"2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023"}

		var passwords []string
		_testArchiveListing(_metaObj, false, modTimeAssertionParentPathArr, passwords)
	})

	Convey("Windows Archive Listing 1024mb dict best - rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_1024mb_dict_best.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		modTimeAssertionParentPathArr := []string{"2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023"}

		var passwords []string

		_testArchiveListing(_metaObj, false, modTimeAssertionParentPathArr, passwords)
	})

	Convey("Windows Archive Listing best enc - rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir1_best_enc.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}

		modTimeAssertionParentPathArr := []string{"2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023", "2023"}

		_testArchiveListing(_metaObj, false, modTimeAssertionParentPathArr, passwords)
	})

	Convey("Windows Archive Listing RAR4 file - rar", t, func() {
		filename := getTestMocksAsset("windows_mocks/mock_dir_rar4.rar")
		_metaObj := &ArchiveMeta{Filename: filename}

		passwords := []string{"1234567"}

		_testArchiveListing(_metaObj, false, nil, passwords)
	})

}

func TestArchiveEncryption(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping 'TestArchiveEncryption' testing in short mode")
	//}

	Convey("Zip Archive Encryption", t, func() {
		_testZipArchiveEncryption()
	})

	Convey("Rar Archive Encryption", t, func() {
		_testRarArchiveEncryption()
	})
}

func _testListingSymlinkCommonArchives(metaObj *ArchiveMeta, passwords []string) {
	Convey("Kind | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByKind,
			OrderDir:          OrderDirAsc,
			Passwords:         passwords,
		}

		result, err := GetArchiveFileList(metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var KindArr []string

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			KindArr = append(KindArr, item.Kind())
		}

		assertionParentPathArr := []string{"Alias", "Alias", "Folder", "Folder", "Folder", "Folder", "Text Document", "Text Document", "Text Document", "Text Document", "Text Document"}

		So(KindArr, ShouldResemble, assertionParentPathArr)
	})
}

func _testListingCompressedFile(metaObj *ArchiveMeta, destinationFilename string) {
	Convey("Kind | it should not throw an error", func() {
		_listObj := &ArchiveRead{
			ListDirectoryPath: "",
			Recursive:         true,
			OrderBy:           OrderByKind,
			OrderDir:          OrderDirAsc,
		}

		result, err := GetArchiveFileList(metaObj, _listObj)

		So(err, ShouldBeNil)

		var fullPathArr []string
		var KindArr []string

		for _, item := range result {
			fullPathArr = append(fullPathArr, item.FullPath)
			KindArr = append(KindArr, item.Kind())
		}

		assertionArr := []string{destinationFilename}

		assertionParentPathArr := []string{"Unknown"}

		So(fullPathArr, ShouldResemble, assertionArr)
		So(KindArr, ShouldResemble, assertionParentPathArr)
	})
}

func TestSymlinkListing(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping 'TestPacking' testing in short mode")
	//}

	Convey("Listing | No encryption - ZIP", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename: filename,
		}

		var passwords []string
		_testListingSymlinkCommonArchives(_metaObj, passwords)
	})

	Convey("Listing | Encrypted - ZIP (StandardEncryption)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_stdenc_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename:         filename,
			EncryptionMethod: zip.StandardEncryption,
		}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Encrypted - ZIP (AES128Encryption)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_aes128enc_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename:         filename,
			EncryptionMethod: zip.AES128Encryption,
		}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Encrypted - ZIP (AES256Encryption)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_aes256enc_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename:         filename,
			EncryptionMethod: zip.AES256Encryption,
		}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Encrypted - ZIP (AES192Encryption)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_aes192enc_pack.zip")

		_metaObj := &ArchiveMeta{
			Filename:         filename,
			EncryptionMethod: zip.AES192Encryption,
		}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Tar", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Tar.gz", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.gz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Tar.bz2", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.bz2")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Tar.br (brotli)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.br")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Tar.lz4", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.lz4")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Tar.sz", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.sz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Tar.xz", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.xz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing | Tar.zst (zstd)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.tar.zst")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingSymlinkCommonArchives(_metaObj, []string{})
	})

	Convey("Listing compressed file | GZ", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.gz")

		_metaObj := &ArchiveMeta{Filename: filename}

		_testListingCompressedFile(_metaObj, "arc_test_pack")
	})

	Convey("Listing compressed file | Zstd", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.zst")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testListingCompressedFile(_metaObj, "arc_test_pack")
	})

	Convey("Listing compressed file | Xz", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.xz")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testListingCompressedFile(_metaObj, "arc_test_pack")
	})

	Convey("Listing compressed file | sz (Snappy)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.sz")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testListingCompressedFile(_metaObj, "arc_test_pack")
	})

	Convey("Listing compressed file | Lz4", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.lz4")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testListingCompressedFile(_metaObj, "arc_test_pack")
	})

	Convey("Listing compressed file | Bz2", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.bz2")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testListingCompressedFile(_metaObj, "arc_test_pack")
	})

	Convey("Listing compressed file | BR (Brotli)", t, func() {
		filename := getTestMocksAsset("symlink_tests/arc_test_pack.br")
		_metaObj := &ArchiveMeta{Filename: filename}
		_testListingCompressedFile(_metaObj, "arc_test_pack")
	})
}
