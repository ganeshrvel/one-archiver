package onearchiver

type Errors string

const (
	ErrorInvalidPassword               Errors = "invalid password"
	ErrorPasswordRequired              Errors = "password is required"
	ErrorArchiverList                  Errors = "some error occured while reading the archive"
	ErrorFormatSupported               Errors = "format unrecognized by filename"
	ErrorNoPathToFilter                Errors = "path not found to filter"
	ErrorCompressedFileInvalidSize     Errors = "only a single file can be packed to a compressed file"
	ErrorCompressedFileNoFileFound     Errors = "atleast a single file is required for creating a compress file, check if files are getting ignored by filters"
	ErrorCompressedFileOnlyFileAllowed Errors = "only a single file be packed to a compressed file, no directories are allowed"
)
