package onearchiver

import (
	"github.com/yeka/zip"
	"os"
	"time"
)

type allowedSecondExtMap map[string]string

type ArchiveFileInfo struct {
	Mode       os.FileMode
	Size       int64
	IsDir      bool
	ModTime    time.Time
	Name       string
	FullPath   string
	ParentPath string
	Extension  string
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
	ParentPath    string
	Extension     string
}

type zipArchive struct {
	meta   ArchiveMeta   // required
	read   ArchiveRead   // required for listing files
	pack   ArchivePack   // required for archiving files
	unpack ArchiveUnpack // required for unarchiving files
}

type compressedFile struct {
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
	IsEncrypted     bool
	IsValidPassword bool
}

type ProgressInfo struct {
	StartTime          time.Time
	TotalFiles         int
	ProgressCount      int
	CurrentFilename    string
	ProgressPercentage float32
	lastSentTime       time.Time
}

type ProgressHandler struct {
	OnReceived  func(*ProgressInfo)
	OnError     func(error, *ProgressInfo)
	OnCompleted func(*ProgressInfo)
}
