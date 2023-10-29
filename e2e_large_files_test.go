package onearchiver_test

import (
	"crypto/rand"
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	. "github.com/ganeshrvel/one-archiver"
	zip "github.com/ganeshrvel/yeka_zip"
	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"log"
	"os"
	"path"
	"testing"
	"time"
)

type largeFileTestInfoFileType string

const (
	largeFileTestInfoFileTypeFile    largeFileTestInfoFileType = "file"
	largeFileTestInfoFileTypeSymlink largeFileTestInfoFileType = "symlink"
	//largeFileTestInfoFileTypeHardlink largeFileTestInfoFileType = "hardlink"
)

type largeFileTestingType string

const (
	largeFileTestingTypePacking   largeFileTestingType = "packing"
	largeFileTestingTypeUnpacking largeFileTestingType = "unpacking"
)

type largeFileTestInfo struct {
	fileSize int
	filename string
	fileType largeFileTestInfoFileType
}

type CustomFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

const _ProgressStreamDebounceTime = 2

// Implementing the os.FileInfo interface

func (c *CustomFileInfo) Name() string       { return c.name }
func (c *CustomFileInfo) Size() int64        { return c.size }
func (c *CustomFileInfo) Mode() os.FileMode  { return c.mode }
func (c *CustomFileInfo) ModTime() time.Time { return c.modTime }
func (c *CustomFileInfo) IsDir() bool        { return c.isDir }
func (c *CustomFileInfo) Sys() interface{}   { return c.sys }

func (lf *largeFileTestInfo) fullpath(destination string) string {
	return path.Join(destination, lf.filename)
}

func createTestFile(lf largeFileTestInfo, destination string) {
	fullpath := lf.fullpath(destination)
	if exist := FileExists(fullpath); exist {
		return
	}

	fileSize := lf.fileSize

	file, err := os.Create(fullpath)
	if err != nil {
		log.Panic("Error creating large TestFile:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Panic(err)
		}
	}(file)

	// Copy random data to file
	_, err = io.CopyN(file, rand.Reader, int64(fileSize))
	if err != nil {
		log.Panic("Error writing to file:", err)
		return
	}

}

func getIndexOfLargeFile(lfiArr []largeFileTestInfo, filepath string, largeFilesMocksDir string) (largeFileTestInfo, int, bool) {
	return lo.FindIndexOf(lfiArr, func(item largeFileTestInfo) bool {
		return item.fullpath(largeFilesMocksDir) == filepath
	})
}

func getUpdatedCounter[T comparable](om *orderedmap.OrderedMap[string, []T], currentFilepath string, countItem T) *orderedmap.OrderedMap[string, []T] {
	val, exists := om.Get(currentFilepath)
	if !exists {
		om.Set(currentFilepath, []T{countItem})

		return om
	}

	_, valExists := lo.Find(val, func(item T) bool {
		return countItem == item
	})

	if !valExists {
		val1 := append(val, countItem)
		om.Set(currentFilepath, val1)

	}
	return om
}

func _testLargeFilesPacking(t *testing.T, packingOutputFullFilepath string, lfiArr []largeFileTestInfo, paths largeFileTestsPaths, testFileInfo largeFileTests, zipEncryptionMethod zip.EncryptionMethod, canResumeTransfer, isCompressedFile bool) error {
	Convey("it should not throw an error", func() {

		largeFilesMocksDir := paths.largeFilesMocksDir
		for _, lf := range lfiArr {
			if lf.fileType == largeFileTestInfoFileTypeFile {
				createTestFile(lf, largeFilesMocksDir)
			} else if lf.fileType == largeFileTestInfoFileTypeSymlink {
				err := os.Symlink(paths.symlinkPath, lf.fullpath(largeFilesMocksDir))
				if err != nil {
					log.Panic(err)
				}
			}
		}

		expectedFilePaths := lo.Map(lfiArr, func(lfi largeFileTestInfo, index int) string {
			return lfi.fullpath(largeFilesMocksDir)
		})
		largeFilesMocksFileInfo := getFilesInDirectory(largeFilesMocksDir, expectedFilePaths)

		var fileList []string
		if !isCompressedFile {
			fileList = []string{largeFilesMocksDir}
		} else {
			fileList = []string{lfiArr[0].fullpath(largeFilesMocksDir)}
		}

		err := _largeFilesCommonTests(t, largeFileTestingTypePacking, lfiArr, expectedFilePaths, largeFilesMocksFileInfo, canResumeTransfer, fileList, packingOutputFullFilepath, testFileInfo.pwd, zipEncryptionMethod, paths, testFileInfo)
		if err != nil {
			log.Panic(err)
		}

		return

	})

	return nil
}

func _testLargeFilesUnpacking(t *testing.T, packingOutputFullFilepath string, paths largeFileTestsPaths, testFileInfo largeFileTests, zipEncryptionMethod zip.EncryptionMethod, canResumeTransfer bool) error {

	Convey("Testing unpacking", func() {
		Convey("it should not throw an error", func() {
			archivesInPackingOutputDir := getFilesInDirectory(paths.packingOutputDirPath, []string{packingOutputFullFilepath})
			_archivesInPackingOutputDir := *archivesInPackingOutputDir

			So(len(_archivesInPackingOutputDir), ShouldEqual, 1)

			var archiveFullpath string
			fileInfoMap := make(map[string]os.FileInfo)
			var expectedFilePaths []string
			var archivedFileInfoArr []largeFileTestInfo

			for fp := range _archivesInPackingOutputDir {
				archiveFullpath = fp
			}

			_metaObj := NewArchiveMeta(archiveFullpath)
			_listObj := NewArchiveRead()
			_listObj.ListDirectoryPath = ""
			_listObj.Recursive = true
			_listObj.OrderBy = OrderByFullPath
			_listObj.OrderDir = OrderDirAsc
			_listObj.Passwords = testFileInfo.pwd.passwords

			fileListResult, listerr := GetArchiveFileList(_metaObj, _listObj)
			So(listerr, ShouldBeNil)

			for _, info := range fileListResult {

				fp := path.Join(paths.unpackingOutputDirPath, GetMD5Hash(testFileInfo.title), testFileInfo.filename, info.FullPath)

				if info.IsDir {
					continue
				}

				fileType := largeFileTestInfoFileTypeFile

				if info.Kind() == "Alias" {
					fileType = largeFileTestInfoFileTypeSymlink
				}

				archivedFileInfoArr = append(archivedFileInfoArr, largeFileTestInfo{
					fileSize: int(info.Size),
					filename: info.Name,
					fileType: fileType,
				})

				expectedFilePaths = append(expectedFilePaths, fp)

				mockFileInfo := &CustomFileInfo{
					name:    info.Name,
					size:    info.Size,
					mode:    info.Mode,
					modTime: info.ModTime,
					isDir:   info.IsDir,
				}

				fileInfoMap[fp] = mockFileInfo
			}

			err := _largeFilesCommonTests(t, largeFileTestingTypeUnpacking, archivedFileInfoArr, expectedFilePaths, &fileInfoMap, canResumeTransfer, []string{}, packingOutputFullFilepath, testFileInfo.pwd, zipEncryptionMethod, paths, testFileInfo)
			if err != nil {
				log.Panic(err)
			}

			return
		})

		return
	})

	return nil

}

func _testLargeFilesStartPacking(ph *ProgressFunc, fileList []string, filename string, pwd largeFileTestsPassword, zipEncryptionMethod zip.EncryptionMethod) error {
	metaObj := NewArchiveMeta(filename)
	session := NewSession("", ph)

	packObj := NewArchivePack()
	packObj.FileList = fileList
	packObj.Password = pwd.getSinglePassword()
	packObj.ZipEncryptionMethod = zipEncryptionMethod
	packObj.ProgressStreamDebounceTime = _ProgressStreamDebounceTime

	err := StartPacking(metaObj, packObj, session)

	So(err, ShouldBeNil)

	return nil

}

func _testLargeFilesStartUnpacking(ph *ProgressFunc, filename string, pwd largeFileTestsPassword, paths largeFileTestsPaths, testFileInfo largeFileTests) error {
	metaObj := NewArchiveMeta(filename)
	session := NewSession("", ph)

	destinationPath := path.Join(paths.unpackingOutputDirPath, GetMD5Hash(testFileInfo.title), testFileInfo.filename)
	unpackObj := NewArchiveUnpack()
	unpackObj.Passwords = pwd.passwords
	unpackObj.ProgressStreamDebounceTime = _ProgressStreamDebounceTime
	unpackObj.Destination = destinationPath

	err := StartUnpacking(metaObj, unpackObj, session)

	So(err, ShouldBeNil)

	return nil

}

func _largeFilesCommonTests(t *testing.T, testType largeFileTestingType, lfiArr []largeFileTestInfo, expectedFilePaths []string, fileInfoMap *map[string]os.FileInfo, canResumeTransfer bool, fileList []string, filename string, pwd largeFileTestsPassword, zipEncryptionMethod zip.EncryptionMethod, paths largeFileTestsPaths, testFileInfo largeFileTests) error {

	var prevLatestSentTime int64
	counterCurrentFilepath := orderedmap.NewOrderedMap[string, []string]()
	counterSentFilesCount := orderedmap.NewOrderedMap[string, []int64]()
	counterSentFilesCountPercentage := orderedmap.NewOrderedMap[string, []float64]()
	counterTotalSize := orderedmap.NewOrderedMap[string, []int64]()
	counterSentSize := orderedmap.NewOrderedMap[string, []int64]()
	counterSentSizeProgressPercentage := orderedmap.NewOrderedMap[string, []float64]()
	counterCurrentFileSize := orderedmap.NewOrderedMap[string, []int64]()
	counterCurrentFileSentSize := orderedmap.NewOrderedMap[string, []int64]()
	counterCurrentFileProgressSizePercentage := orderedmap.NewOrderedMap[string, []float64]()
	counterProgressStatus := orderedmap.NewOrderedMap[string, []ProgressStatus]()
	lastCurrentFileSize := int64(0)

	filesTotalSize := int64(0)
	for _, v := range *fileInfoMap {
		filesTotalSize += v.Size()
	}

	ph := &ProgressFunc{
		OnReceived: func(progress *Progress) {
			So(progress.StartTime.Year(), ShouldBeGreaterThanOrEqualTo, 2023)
			So(progress.LatestSentTime.UnixNano(), ShouldBeGreaterThanOrEqualTo, prevLatestSentTime)
			prevLatestSentTime = progress.LatestSentTime.UnixNano()

			So(progress.CurrentFilepath, ShouldNotBeEmpty)

			So(progress.TotalFiles, ShouldEqual, len(lfiArr))
			So(progress.ProgressStreamDebounceTime, ShouldEqual, _ProgressStreamDebounceTime)
			So(progress.CanResumeTransfer, ShouldEqual, canResumeTransfer)

			getUpdatedCounter(counterCurrentFilepath, progress.CurrentFilepath, progress.CurrentFilepath)
			getUpdatedCounter(counterSentFilesCount, progress.CurrentFilepath, progress.SentFilesCount)
			getUpdatedCounter(counterSentFilesCountPercentage, progress.CurrentFilepath, progress.SentFilesCountPercentage)
			getUpdatedCounter(counterTotalSize, progress.CurrentFilepath, progress.TotalSize)
			getUpdatedCounter(counterSentSize, progress.CurrentFilepath, progress.SentSize)
			getUpdatedCounter(counterSentSizeProgressPercentage, progress.CurrentFilepath, progress.SentSizeProgressPercentage)

			getUpdatedCounter(counterCurrentFileSize, progress.CurrentFilepath, progress.CurrentFileSize)
			lastCurrentFileSize = progress.CurrentFileSize

			getUpdatedCounter(counterCurrentFileSentSize, progress.CurrentFilepath, progress.CurrentFileSentSize)
			getUpdatedCounter(counterCurrentFileProgressSizePercentage, progress.CurrentFilepath, progress.CurrentFileProgressSizePercentage)
			getUpdatedCounter(counterProgressStatus, progress.CurrentFilepath, progress.ProgressStatus)

			So(progress.CurrentFilepath, ShouldBeIn, expectedFilePaths)
			So(progress.ProgressCancelReason, ShouldEqual, ProgressCancelReasonNone)
		},
		OnEnded: func(progress *Progress) {
			So(progress.TotalFiles, ShouldEqual, len(lfiArr))
			So(progress.SentFilesCountPercentage, ShouldEqual, 100)
			So(progress.SentSizeProgressPercentage, ShouldEqual, 100)

			So(progress.SentFilesCount, ShouldEqual, len(lfiArr))
			So(progress.ProgressStatus, ShouldEqual, ProgressStatusCompleted)

			So(progress.CurrentFileSize, ShouldEqual, lastCurrentFileSize)
			So(progress.CurrentFileSentSize, ShouldEqual, lastCurrentFileSize)
			So(progress.CurrentFileProgressSizePercentage, ShouldEqual, 100)

			So(progress.TotalSize, ShouldEqual, filesTotalSize)
			So(progress.SentSize, ShouldEqual, filesTotalSize)
			So(progress.ProgressCancelReason, ShouldEqual, ProgressCancelReasonNone)
		},
	}

	if testType == largeFileTestingTypePacking {
		err := _testLargeFilesStartPacking(ph, fileList, filename, pwd, zipEncryptionMethod)
		if err != nil {
			log.Panic(err)
		}
	}
	if testType == largeFileTestingTypeUnpacking {
		err := _testLargeFilesStartUnpacking(ph, filename, pwd, paths, testFileInfo)
		if err != nil {
			log.Panic(err)
		}
	}

	t.Logf("\nCurrentFilepath incrementing")
	counter := 0
	for el := counterCurrentFilepath.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value
		currentFilepath := el.Key

		currentFileLocalInfo := (*fileInfoMap)[currentFilepath]
		currentFileName := currentFileLocalInfo.Name()

		for _, value := range itemsArr {
			So(value, ShouldEndWith, currentFileName)
		}

		counter++
	}

	So(counter, ShouldBeGreaterThan, 0)

	t.Logf("\nCurrentFilepath incrementing")
	counter = 0
	for el := counterSentFilesCount.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value

		So(itemsArr, ShouldResemble, []int64{itemsArr[0], itemsArr[0] + 1})

		counter++
	}

	So(counter, ShouldBeGreaterThan, 0)

	t.Logf("\nSentFilesCountPercentage incrementing")
	lastPercValue := float64(0)
	counter = 0

	for el := counterSentFilesCountPercentage.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value

		So(itemsArr[0], ShouldEqual, lastPercValue)

		for idx, value := range itemsArr {
			if idx == 0 {
				continue
			}
			So(value, ShouldBeGreaterThan, itemsArr[idx-1])
		}

		last, err := lo.Last(itemsArr)
		if err != nil {
			log.Panic(err)
		}
		lastPercValue = last

		counter++
	}
	So(counter, ShouldBeGreaterThan, 0)
	So(lastPercValue, ShouldEqual, 100)

	t.Logf("\nTotalSize incrementing")
	counter = 0
	for el := counterTotalSize.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value

		last, err := lo.Last(itemsArr)
		if err != nil {
			log.Panic(err)
		}

		So(last, ShouldEqual, filesTotalSize)

		counter++
	}

	So(counter, ShouldBeGreaterThan, 0)

	t.Logf("\nSentSize incrementing")
	counter = 0
	totalSize := int64(0)
	for el := counterSentSize.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value
		currentFilepath := el.Key

		last, err := lo.Last(itemsArr)
		if err != nil {
			log.Panic(err)
		}

		currentFileLocalInfo := (*fileInfoMap)[currentFilepath]
		currentFileSize := currentFileLocalInfo.Size()

		if testFileInfo.deflectiveProgress {
			So(last, ShouldBeBetweenOrEqual, min0Int64(totalSize+currentFileSize-10), totalSize+currentFileSize+10)
		} else {
			So(last, ShouldEqual, totalSize+currentFileSize)
		}

		totalSize = last

		counter++
	}

	if testFileInfo.deflectiveProgress {
		So(totalSize, ShouldBeBetweenOrEqual, min0Int64(totalSize-10), totalSize+10)
	} else {
		So(totalSize, ShouldEqual, filesTotalSize)
	}

	So(counter, ShouldBeGreaterThan, 0)

	t.Logf("\nSentSizeProgressPercentage incrementing")
	lastPercValue = float64(0)
	counter = 0

	for el := counterSentSizeProgressPercentage.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value

		if testFileInfo.deflectiveProgress {
			So(itemsArr[0], ShouldBeBetweenOrEqual, min0Float64(lastPercValue-10), lastPercValue+10)
		} else {
			So(itemsArr[0], ShouldEqual, lastPercValue)
		}

		for idx, value := range itemsArr {
			if idx == 0 {
				continue
			}

			if testFileInfo.deflectiveProgress {
				So(value, ShouldBeGreaterThanOrEqualTo, min0Float64(itemsArr[idx-1]-10))
			} else {
				So(value, ShouldBeGreaterThan, itemsArr[idx-1])
			}

		}

		last, err := lo.Last(itemsArr)
		if err != nil {
			log.Panic(err)
		}
		lastPercValue = last

		counter++
	}
	So(counter, ShouldBeGreaterThan, 0)
	So(lastPercValue, ShouldEqual, 100)

	t.Logf("\nCurrentFileSize incrementing")
	counter = 0

	for el := counterCurrentFileSize.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value
		currentFilepath := el.Key

		first := itemsArr[0]
		last, err := lo.Last(itemsArr)
		if err != nil {
			log.Panic(err)
		}

		currentFileLocalInfo := (*fileInfoMap)[currentFilepath]
		currentFileSize := currentFileLocalInfo.Size()

		So(first, ShouldEqual, 0)
		So(last, ShouldEqual, currentFileSize)
		counter++
	}

	So(counter, ShouldBeGreaterThan, 0)

	t.Logf("\nCurrentFileSentSize incrementing")
	counter = 0
	totalSize = int64(0)
	for el := counterCurrentFileSentSize.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value
		currentFilepath := el.Key

		last, err := lo.Last(itemsArr)
		if err != nil {
			log.Panic(err)
		}

		currentFileLocalInfo := (*fileInfoMap)[currentFilepath]
		currentFileSize := currentFileLocalInfo.Size()

		if testFileInfo.deflectiveProgress {
			So(last, ShouldBeBetweenOrEqual, min0Int64(currentFileSize-10), currentFileSize+10)
		} else {
			So(last, ShouldEqual, currentFileSize)
		}

		totalSize += last

		counter++
	}

	if testFileInfo.deflectiveProgress {
		So(totalSize, ShouldBeBetweenOrEqual, min0Int64(filesTotalSize-10), filesTotalSize+10)
	} else {
		So(totalSize, ShouldEqual, filesTotalSize)
	}

	So(counter, ShouldBeGreaterThan, 0)

	t.Logf("\nCurrentFileProgressSizePercentage incrementing")
	totalPercSent := float64(0)
	counter = 0

	for el := counterCurrentFileProgressSizePercentage.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value

		So(itemsArr[0], ShouldEqual, 0)

		for idx, value := range itemsArr {
			if idx == 0 {
				continue
			}
			So(value, ShouldBeGreaterThan, itemsArr[idx-1])
		}

		last, err := lo.Last(itemsArr)
		if err != nil {
			log.Panic(err)
		}
		So(last, ShouldEqual, 100)
		totalPercSent += last

		counter++
	}
	So(counter, ShouldBeGreaterThan, 0)
	So(totalPercSent, ShouldEqual, 100*len(lfiArr))

	t.Logf("\nProgressStatus incrementing")
	counter = 0

	for el := counterProgressStatus.Front(); el != nil; el = el.Next() {
		itemsArr := el.Value

		for _, value := range itemsArr {
			So(value, ShouldEqual, ProgressStatusRunning)
		}

		counter++
	}
	So(counter, ShouldBeGreaterThan, 0)

	return nil

}

type largeFileTests struct {
	title               string
	filename            string
	pwd                 largeFileTestsPassword
	zipEncryptionMethod zip.EncryptionMethod

	// This is a workaround.
	// When extracting zipped files with multiple passwords, if the initial password is incorrect
	// but subsequent passwords are correct, the progress for the current file's size should have already been emitted.
	// Before invoking the progress correction API, this progress would have been emitted, leading to confusion in the progress statistics.
	// To prevent misleading test results, we will agree to any results between an allowed threshold
	deflectiveProgress bool
}

type largeFileTestsPaths struct {
	largeFilesMocksDir     string
	packingOutputDirPath   string
	unpackingOutputDirPath string
	symlinkPath            string
}

type largeFileTestsPassword struct {
	passwords []string
}

func (lft *largeFileTestsPassword) hasPasswords() bool {
	return len(lft.passwords) > 0
}

func (lft *largeFileTestsPassword) getSinglePassword() string {
	if !lft.hasPasswords() {
		return ""
	}
	return lft.passwords[0]
}

func (lft *largeFileTests) withDestinationPath(destination string) string {
	return path.Join(destination, lft.filename)
}

func getPaths(destinationDirId string) largeFileTestsPaths {
	largeFilesMocksDir := newTempMocksDir(path.Join("large_files_test_mock", destinationDirId), false)
	packingOutputDirPath := newTempMocksDir(path.Join("large_files_test_packing_output", destinationDirId), false)
	unpackingOutputDirPath := newTempMocksDir(path.Join("large_files_test_unpacking_output", destinationDirId), false)

	return largeFileTestsPaths{
		largeFilesMocksDir:     largeFilesMocksDir,
		packingOutputDirPath:   packingOutputDirPath,
		unpackingOutputDirPath: unpackingOutputDirPath,
		symlinkPath:            path.Join("..", "symlink_target.txt"),
	}
}

func getRarPaths(destinationDirId string) largeFileTestsPaths {
	largeFilesMocksDir := getTestMocksAsset(path.Join("large_files_test"))
	unpackingOutputDirPath := newTempMocksDir(path.Join("large_files_test_unpacking_output", "rar", destinationDirId), false)

	return largeFileTestsPaths{
		largeFilesMocksDir:     largeFilesMocksDir,
		packingOutputDirPath:   largeFilesMocksDir,
		unpackingOutputDirPath: unpackingOutputDirPath,
		symlinkPath:            path.Join("..", "symlink_target.txt"),
	}
}

func TestLargeFiles(t *testing.T) {
	_ = newTempMocksDir("large_files_test_mock", true)
	_ = newTempMocksDir("large_files_test_packing_output", true)
	_ = newTempMocksDir("large_files_test_unpacking_output", true)

	Convey("Large files testing", t, func() {
		archiveFilesLargeFileTestsArr := []largeFileTests{
			{
				title:    "No encryption - ZIP",
				filename: "noenc_pack.zip",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "STD encryption - ZIP",
				filename: "std_enc_pack.zip",
				pwd: largeFileTestsPassword{
					passwords: []string{"1234567"},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "AES256Encryption encryption - ZIP",
				filename: "AES256_enc_pack.zip",
				pwd: largeFileTestsPassword{
					passwords: []string{"1234567"},
				},
				zipEncryptionMethod: zip.AES256Encryption,
			}, {
				title:    "AES192Encryption encryption - ZIP",
				filename: "AES192_enc_pack.zip",
				pwd: largeFileTestsPassword{
					passwords: []string{"1234567"},
				},
				zipEncryptionMethod: zip.AES192Encryption,
			}, {
				title:    "tar",
				filename: "pack.tar",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "Tar.gz",
				filename: "pack.tar.gz",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "Tar.bz2",
				filename: "pack.tar.bz2",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "Tar.br",
				filename: "pack.tar.br",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "Tar.lz4",
				filename: "pack.tar.lz4",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "Tar.sz",
				filename: "pack.tar.sz",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "Tar.xz",
				filename: "pack.tar.xz",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "Tar.zst",
				filename: "pack.tar.zst",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			},
		}
		compressedFilesLargeFileTestsArr := []largeFileTests{
			{
				title:    "gz",
				filename: "pack.file.gz",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "zst",
				filename: "pack.file.zst",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "xz",
				filename: "pack.file.xz",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "sz",
				filename: "pack.file.sz",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "lz4",
				filename: "pack.file.lz4",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "bz2",
				filename: "pack.file.bz2",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			}, {
				title:    "br",
				filename: "pack.file.br",
				pwd: largeFileTestsPassword{
					passwords: []string{},
				},
				zipEncryptionMethod: zip.StandardEncryption,
			},
		}

		lfiArr := []largeFileTestInfo{
			{
				fileSize: 60 * 1000,
				filename: "60kb",
				fileType: largeFileTestInfoFileTypeFile,
			},
			{
				fileSize: 40 * 1000,
				filename: "40kb",
				fileType: largeFileTestInfoFileTypeFile,
			},
			{
				fileSize: 10 * 1000,
				filename: "10kb",
				fileType: largeFileTestInfoFileTypeFile,
			}, {
				fileSize: 0,
				filename: "0b",
				fileType: largeFileTestInfoFileTypeFile,
			}, {
				fileSize: 10,
				filename: "symlink_target.txt",
				fileType: largeFileTestInfoFileTypeFile,
			},
			{
				filename: path.Join("dir1", "symlink_1"),
				fileType: largeFileTestInfoFileTypeSymlink,
			},
		}

		Convey("archive files multiple files", func() {

			paths := getPaths("archive_files_multiple_files")
			_ = newTempMocksDir(path.Join("large_files_test_mock", "archive_files_multiple_files", "dir1"), true)

			for _, v := range archiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing packing", v.title), func() {

					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesPacking(t, packingOutputFullFilepath, lfiArr, paths, v, v.zipEncryptionMethod, true, false)
					if err != nil {
						log.Panic(err)
					}
				})
			}

			for _, v := range archiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, true)
					if err != nil {
						log.Panic(err)
					}

				})
			}

			encryptedZipArchiveFilesLargeFileTestsArr := []largeFileTests{
				{
					title:    "STD encryption - Wrong password - Zip - 1",
					filename: "std_enc_pack.zip",
					pwd: largeFileTestsPassword{
						[]string{"", "demo", "1234567"},
					},
					deflectiveProgress:  true,
					zipEncryptionMethod: zip.StandardEncryption,
				},

				{
					title:    "STD encryption - Wrong password - Zip",
					filename: "std_enc_pack.zip",
					pwd: largeFileTestsPassword{
						[]string{"1234", "demo", "1234567"},
					},
					deflectiveProgress:  true,
					zipEncryptionMethod: zip.StandardEncryption,
				},

				{
					title:    "AES192 encryption - Wrong password - Zip",
					filename: "AES192_enc_pack.zip",
					pwd: largeFileTestsPassword{
						[]string{"1234", "demo", "1234567"},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				}, {
					title:    "AES256 encryption - Wrong password - Zip",
					filename: "AES256_enc_pack.zip",
					pwd: largeFileTestsPassword{
						[]string{"1234", "demo", "1234567"},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				},
			}

			for _, v := range encryptedZipArchiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking with multiple wrong and one correct password", v.title), func() {

					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, true)
					if err != nil {
						log.Panic(err)
					}
				})
			}

		})

		Convey("archive files single files", func() {

			paths := getPaths("archive_files_single_files")

			for _, v := range archiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing packing", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)

					err := _testLargeFilesPacking(t, packingOutputFullFilepath, []largeFileTestInfo{{
						fileSize: 40 * 1000,
						filename: "40kb",
						fileType: largeFileTestInfoFileTypeFile},
					}, paths, v, v.zipEncryptionMethod, true, false)

					if err != nil {
						log.Panic(err)
					}
				})
			}
			for _, v := range archiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)

					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, true)
					if err != nil {
						log.Panic(err)
					}
				})
			}
		})

		Convey("archive files 0 bytes", func() {

			paths := getPaths("archive_files_0_bytes")

			for _, v := range archiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing packing", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesPacking(t, packingOutputFullFilepath, []largeFileTestInfo{
						{
							fileSize: 0,
							filename: "0b",
							fileType: largeFileTestInfoFileTypeFile,
						},
					}, paths, v, v.zipEncryptionMethod, true, false)
					if err != nil {
						log.Panic(err)
					}
				})
			}
			for _, v := range archiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, true)
					if err != nil {
						log.Panic(err)
					}
				})
			}
		})

		Convey("compressed files greater than 0 bytes", func() {
			paths := getPaths("compressed_files_less_greater_than_bytes")

			for _, v := range compressedFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing packing", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesPacking(t, packingOutputFullFilepath, []largeFileTestInfo{
						{
							fileSize: 60 * 1000,
							filename: "60kb",
							fileType: largeFileTestInfoFileTypeFile,
						},
					}, paths, v, v.zipEncryptionMethod, false, true)
					if err != nil {
						log.Panic(err)
					}
				})
			}
			for _, v := range compressedFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, false)
					if err != nil {
						log.Panic(err)
					}
				})
			}
		})

		Convey("compressed files = 0 bytes", func() {

			paths := getPaths("compressed_files_0_bytes")

			for _, v := range compressedFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing packing", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)

					err := _testLargeFilesPacking(t, packingOutputFullFilepath, []largeFileTestInfo{
						{
							fileSize: 0,
							filename: "0b",
							fileType: largeFileTestInfoFileTypeFile,
						},
					}, paths, v, v.zipEncryptionMethod, false, true)
					if err != nil {
						log.Panic(err)
					}
				})
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)

					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, false)
					if err != nil {
						log.Panic(err)
					}
				})
			}
		})
	})

	Convey("Large files testing - Rar", t, func() {
		Convey("archive files multiple files", func() {

			encryptedRarArchiveFilesLargeFileTestsArr := []largeFileTests{
				{
					title:    "Rar4 - non encrypted",
					filename: "archive_files_multiple_files_rar4.rar",
					pwd: largeFileTestsPassword{
						[]string{},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				},

				{
					title:    "Rar5 - non encrypted",
					filename: "archive_files_multiple_files_rar5.rar",
					pwd: largeFileTestsPassword{
						[]string{},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				},

				{
					title:    "Rar5 - encrypted",
					filename: "archive_files_multiple_files_rar5_enc.rar",
					pwd: largeFileTestsPassword{
						[]string{"1234567"},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				},

				{
					title:    "Rar5 - encrypted file names and locked",
					filename: "archive_files_multiple_files_rar5_enc_filename_enc_locked_archive.rar",
					pwd: largeFileTestsPassword{
						[]string{"1234567"},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				},
			}

			paths := getRarPaths("archive_files_multiple_files")
			for _, v := range encryptedRarArchiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking", v.title), func() {

					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, true)
					if err != nil {
						log.Panic(err)
					}
				})
			}
		})

		Convey("archive files single files", func() {

			encryptedRarArchiveFilesLargeFileTestsArr := []largeFileTests{
				{
					title:    "Rar4 - non encrypted",
					filename: "archive_files_single_files_rar4.rar",
					pwd: largeFileTestsPassword{
						[]string{},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				}, {
					title:    "Rar5 - non encrypted",
					filename: "archive_files_single_files_rar5.rar",
					pwd: largeFileTestsPassword{
						[]string{},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				},
			}

			paths := getRarPaths("archive_files_single_files")
			for _, v := range encryptedRarArchiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)

					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, true)
					if err != nil {
						log.Panic(err)
					}
				})
			}
		})

		Convey("archive files 0 bytes", func() {
			encryptedRarArchiveFilesLargeFileTestsArr := []largeFileTests{
				{
					title:    "Rar5 - non encrypted",
					filename: "archive_files_0_bytes_rar5.rar",
					pwd: largeFileTestsPassword{
						[]string{},
					},
					zipEncryptionMethod: zip.StandardEncryption,
				},
			}

			paths := getRarPaths("archive_files_single_files")
			for _, v := range encryptedRarArchiveFilesLargeFileTestsArr {
				Convey(fmt.Sprintf("%s - %s", "Testing unpacking", v.title), func() {
					packingOutputFullFilepath := v.withDestinationPath(paths.packingOutputDirPath)
					err := _testLargeFilesUnpacking(t, packingOutputFullFilepath, paths, v, v.zipEncryptionMethod, true)
					if err != nil {
						log.Panic(err)
					}
				})
			}
		})
	})

}
