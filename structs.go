package onearchiver

import (
	"github.com/ganeshrvel/yeka_zip"
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

func (fi ArchiveFileInfo) Kind() string {
	return getFileKind(fi.Extension, fi.Mode, fi.IsDir)
}

type ArchiveMeta struct {
	Filename         string
	GitIgnorePattern []string
	EncryptionMethod zip.EncryptionMethod
}

type ArchiveRead struct {
	Passwords         []string
	ListDirectoryPath string
	OrderBy           ArchiveOrderBy
	OrderDir          ArchiveOrderDir
	Recursive         bool
}

func (ar *ArchiveRead) passwordContext() *PasswordContext {
	return &PasswordContext{passwords: ar.Passwords}
}

type ArchivePack struct {
	Password string
	FileList []string
}

type ArchiveUnpack struct {
	Passwords   []string
	FileList    []string
	Destination string
}

func (au *ArchiveUnpack) passwordContext() *PasswordContext {
	return &PasswordContext{passwords: au.Passwords}
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
	Kind          string
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
	prepare() (PrepareArchiveInfo, error)
}

type ArchivePacker interface {
	doPack(session *Session) error
}

type ArchiveUnpacker interface {
	doUnpack(session *Session) error
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

	// Formatted values
	fileInfo *ArchiveFileInfo

	// File provides methods for accessing information about
	// or contents of a file within an archive.
	sourceArchiveFileInfo *os.FileInfo
}

type PrepareArchiveInfo struct {
	IsValidPassword      bool
	IsSinglePasswordMode bool
	IsPasswordRequired   bool
}
