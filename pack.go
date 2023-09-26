package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	"os"
	"strings"
)

func (arc zipArchive) doPack(session *Session) error {
	fileList := arc.pack.FileList

	commonParentPath := GetCommonParentPath(os.PathSeparator, fileList...)

	if indexExists(&fileList, 0) && commonParentPath == fileList[0] {
		commonParentPathSplitted := strings.Split(fileList[0], PathSep)

		commonParentPath = strings.Join(commonParentPathSplitted[:len(commonParentPathSplitted)-1], PathSep)
	}

	if err := createZipFile(session, &arc, fileList, commonParentPath); err != nil {
		return err
	}

	return nil
}

func (arc commonArchive) doPack(session *Session) error {
	filename := arc.meta.Filename
	fileList := arc.pack.FileList
	password := arc.pack.Password

	arcFileObj, err := archiver.ByExtension(filename)

	if err != nil {
		return err
	}

	err = archiveFormat(&arcFileObj, password, OverwriteExisting)

	if err != nil {
		return err
	}

	commonParentPath := GetCommonParentPath(os.PathSeparator, fileList...)

	if indexExists(&fileList, 0) && commonParentPath == fileList[0] {
		commonParentPathSplitted := strings.Split(fileList[0], PathSep)

		commonParentPath = strings.Join(commonParentPathSplitted[:len(commonParentPathSplitted)-1], PathSep)
	}

	switch archFileWrite := arcFileObj.(type) {
	case *archiver.Tar:
		err = packTarballs(session, &arc, archFileWrite, &fileList, commonParentPath)
	case *archiver.TarGz:
		err = packTarballs(session, &arc, archFileWrite, &fileList, commonParentPath)
	case *archiver.TarBz2:
		err = packTarballs(session, &arc, archFileWrite, &fileList, commonParentPath)
	case *archiver.TarBrotli:
		err = packTarballs(session, &arc, archFileWrite, &fileList, commonParentPath)
	case *archiver.TarLz4:
		err = packTarballs(session, &arc, archFileWrite, &fileList, commonParentPath)
	case *archiver.TarSz:
		err = packTarballs(session, &arc, archFileWrite, &fileList, commonParentPath)
	case *archiver.TarXz:
		err = packTarballs(session, &arc, archFileWrite, &fileList, commonParentPath)
	case *archiver.TarZstd:
		err = packTarballs(session, &arc, archFileWrite, &fileList, commonParentPath)
	case *archiver.Gz:
		err = packCompressedFile(session, &arc, archFileWrite, &fileList)
	case *archiver.Brotli:
		err = packCompressedFile(session, &arc, archFileWrite, &fileList)
	case *archiver.Bz2:
		err = packCompressedFile(session, &arc, archFileWrite, &fileList)
	case *archiver.Lz4:
		err = packCompressedFile(session, &arc, archFileWrite, &fileList)
	case *archiver.Snappy:
		err = packCompressedFile(session, &arc, archFileWrite, &fileList)
	case *archiver.Xz:
		err = packCompressedFile(session, &arc, archFileWrite, &fileList)
	case *archiver.Zstd:
		err = packCompressedFile(session, &arc, archFileWrite, &fileList)

	default:
		return fmt.Errorf(string(ErrorFormatUnSupported))
	}

	if err != nil {
		return err
	}

	return nil
}

func StartPacking(meta *ArchiveMeta, pack *ArchivePack, session *Session) error {
	_meta := *meta
	_pack := *pack

	var arcPackObj ArchivePacker

	ext := extension(_meta.Filename)

	if OverwriteExisting && FileExists(_meta.Filename) {
		if err := os.Remove(_meta.Filename); err != nil {
			return err
		}
	}

	switch ext {
	case "zip":
		arcPackObj = zipArchive{meta: _meta, pack: _pack}

		break

	default:
		arcPackObj = commonArchive{meta: _meta, pack: _pack}

		break
	}

	return arcPackObj.doPack(session)
}
