package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
)

func (arc zipArchive) doUnpack(ph *ProgressHandler) error {
	return startUnpackingZip(arc, ph)
}

func (arc commonArchive) doUnpack(ph *ProgressHandler) error {
	filename := arc.meta.Filename
	password := arc.meta.Password

	arcFileObj, err := archiver.ByExtension(filename)

	if err != nil {
		return err
	}

	err = archiveFormat(&arcFileObj, password, OverwriteExisting)

	if err != nil {
		return err
	}

	var arcWalker, ok = arcFileObj.(archiver.Walker)
	if !ok {
		return fmt.Errorf(string(ErrorArchiverList))
	}

	return startUnpackingCommonArchives(arc, arcWalker, ph)
}

func (arc compressedFile) doUnpack(ph *ProgressHandler) error {
	filename := arc.meta.Filename
	password := arc.meta.Password

	arcFileObj, err := archiver.ByExtension(filename)
	if err != nil {
		return err
	}

	err = archiveFormat(&arcFileObj, password, OverwriteExisting)
	if err != nil {
		return err
	}

	switch arcFileDecompressor := arcFileObj.(type) {
	case *archiver.Gz:
		err = startUnpackingCompressedFiles(arc, arcFileDecompressor, ph)
	case *archiver.Brotli:
		err = startUnpackingCompressedFiles(arc, arcFileDecompressor, ph)
	case *archiver.Bz2:
		err = startUnpackingCompressedFiles(arc, arcFileDecompressor, ph)
	case *archiver.Lz4:
		err = startUnpackingCompressedFiles(arc, arcFileDecompressor, ph)
	case *archiver.Snappy:
		err = startUnpackingCompressedFiles(arc, arcFileDecompressor, ph)
	case *archiver.Xz:
		err = startUnpackingCompressedFiles(arc, arcFileDecompressor, ph)
	case *archiver.Zstd:
		err = startUnpackingCompressedFiles(arc, arcFileDecompressor, ph)

	default:
		return fmt.Errorf(string(ErrorFormatSupported))
	}

	return nil
}

func StartUnpacking(meta *ArchiveMeta, pack *ArchiveUnpack, ph *ProgressHandler) error {
	_meta := *meta
	_pack := *pack

	var arcUnpackObj ArchiveUnpacker

	// check whether the archive is encrypted
	// if yes, check whether the password is valid
	prepareArchive, err := PrepareArchive(meta)

	if err != nil {
		return err
	}

	if prepareArchive.IsPasswordRequired && !prepareArchive.IsValidPassword {
		return fmt.Errorf(string(ErrorInvalidPassword))
	}

	ext := extension(_meta.Filename)

	switch ext {
	case "zip":
		arcUnpackObj = zipArchive{meta: _meta, unpack: _pack}
	case "zst":
		fallthrough
	case "xz":
		fallthrough
	case "sz":
		fallthrough
	case "lz4":
		fallthrough
	case "bz2":
		fallthrough
	case "br":
		fallthrough
	case "gz":
		arcUnpackObj = compressedFile{meta: _meta, unpack: _pack}
	case "tar.zst":
		fallthrough
	case "tar.xz":
		fallthrough
	case "tar.sz":
		fallthrough
	case "tar.lz4":
		fallthrough
	case "tar.bz2":
		fallthrough
	case "tar.br":
		fallthrough
	case "tar.gz":
		fallthrough
	default:
		arcUnpackObj = commonArchive{meta: _meta, unpack: _pack}
	}

	return arcUnpackObj.doUnpack(ph)
}
