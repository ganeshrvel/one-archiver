package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
)

func archiveFormat(arcFileObj *interface{}, singlePassword string, overwriteExisting bool) error {
	const (
		mkdirAll               = true
		implicitTopLevelFolder = false
		continueOnError        = false
		compressionLevel       = 9
		selectiveCompression   = false
	)

	tarObj := &archiver.Tar{
		OverwriteExisting:      overwriteExisting,
		MkdirAll:               mkdirAll,
		ImplicitTopLevelFolder: implicitTopLevelFolder,
		ContinueOnError:        continueOnError,
	}

	_arcFileObj := *arcFileObj

	// refer https://github.com/ganeshrvel/archiver/blob/master/cmd/arc/main.go for more
	switch arcValues := _arcFileObj.(type) {
	case *archiver.Rar:
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.MkdirAll = mkdirAll
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError
		arcValues.Password = singlePassword
		break

	case *archiver.Tar:
		arcValues = tarObj
		arcValues.MkdirAll = mkdirAll
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError

		break

	case *archiver.TarBrotli:
		arcValues.Tar = tarObj
		arcValues.MkdirAll = mkdirAll
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError
		arcValues.Quality = compressionLevel

		break

	case *archiver.TarBz2:
		arcValues.Tar = tarObj
		arcValues.MkdirAll = mkdirAll
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError
		arcValues.CompressionLevel = compressionLevel

		break

	case *archiver.TarGz:
		arcValues.Tar = tarObj
		arcValues.MkdirAll = mkdirAll
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError
		arcValues.CompressionLevel = compressionLevel

		break

	case *archiver.TarLz4:
		arcValues.Tar = tarObj
		arcValues.MkdirAll = mkdirAll
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError
		arcValues.CompressionLevel = compressionLevel

		break

	case *archiver.TarSz:
		arcValues.Tar = tarObj
		arcValues.MkdirAll = mkdirAll
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError

		break

	case *archiver.TarXz:
		arcValues.Tar = tarObj
		arcValues.MkdirAll = mkdirAll
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError

		break

	case *archiver.TarZstd:
		arcValues.Tar = tarObj
		arcValues.MkdirAll = mkdirAll
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError

		break

	case *archiver.Zip:
		arcValues.CompressionLevel = compressionLevel
		arcValues.OverwriteExisting = overwriteExisting
		arcValues.MkdirAll = mkdirAll
		arcValues.SelectiveCompression = selectiveCompression
		arcValues.ImplicitTopLevelFolder = implicitTopLevelFolder
		arcValues.ContinueOnError = continueOnError

		break

	case *archiver.Brotli:
		arcValues.Quality = compressionLevel

		break

	case *archiver.Bz2:
		arcValues.CompressionLevel = compressionLevel

		break

	case *archiver.Lz4:
		arcValues.CompressionLevel = compressionLevel

		break

	case *archiver.Gz:
		arcValues.CompressionLevel = compressionLevel

		break

	case *archiver.Snappy:
		break

	case *archiver.Xz:
		break

	case *archiver.Zstd:
		break

	default:
		return fmt.Errorf(string(ErrorFormatUnSupported))
	}

	return nil
}
