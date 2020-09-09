package one_archiver

import (
	"github.com/yeka/zip"
	"os"
	"time"
)

type ArchiveFileInfo struct {
	Mode     os.FileMode
	Size     int64
	IsDir    bool
	ModTime  time.Time
	Name     string
	FullPath string
}

type ArchiveMeta struct {
	Filename         string
	Password         string
	GitIgnorePattern []string
	EncryptionMethod zip.EncryptionMethod
}

type ArchiveRead struct {
	ListDirectoryPath string
	OrderBy           ArchiveOrderBy
	OrderDir          ArchiveOrderDir
	Recursive         bool
}

type ArchivePack struct {
	FileList []string
}

type ArchiveUnpack struct {
	FileList    []string
	Destination string
}

type filePathListSortInfo struct {
	splittedPaths [2]string
	IsDir         bool
	Mode          os.FileMode
	Size          int64
	ModTime       time.Time
	Name          string
	FullPath      string
}

type zipArchive struct {
	meta   ArchiveMeta   // required
	read   ArchiveRead   // required for listing files
	pack   ArchivePack   // required for archiving files
	unpack ArchiveUnpack // required for unarchiving files
}

type commonArchive struct {
	meta   ArchiveMeta   // required
	read   ArchiveRead   // required for listing files
	pack   ArchivePack   // required for archiving files
	unpack ArchiveUnpack // required for unarchiving files
}

type ArchiveReader interface {
	list() ([]ArchiveFileInfo, error)
}

type ArchiveUtils interface {
	isEncrypted() (EncryptedArchiveInfo, error)
}

type ArchivePacker interface {
	doPack(ph *ProgressHandler) error
}

type ArchiveUnpacker interface {
	doUnpack(ph *ProgressHandler) error
}

type createArchiveFileInfo struct {
	absFilepath, relativeFilePath string
	isDir                         bool
	fileInfo                      *os.FileInfo
}

type extractZipFileInfo struct {
	absFilepath, name string
	fileInfo          *os.FileInfo
	zipFileInfo       *zip.File
}

type extractCommonArchiveFileInfo struct {
	absFilepath, name string
	fileInfo          *ArchiveFileInfo
	fileBytes         *[]byte
}

type EncryptedArchiveInfo struct {
	isEncrypted     bool
	isValidPassword bool
}

type ProgressInfo struct {
	startTime          time.Time
	totalFiles         int
	progressCount      int
	currentFilename    string
	progressPercentage float32
}

type ProgressHandler struct {
	onReceived  func(*ProgressInfo)
	onError     func(error, *ProgressInfo)
	onCompleted func(*ProgressInfo)
}
