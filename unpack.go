package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	"path/filepath"
)

func (arc zipArchive) doUnpack(ph *ProgressHandler) error {
	return startUnpackingZip(arc, ph)
}

func (arc commonArchive) doUnpack(ph *ProgressHandler) error {
	_filename := arc.meta.Filename
	_password := arc.meta.Password

	arcFileObj, err := archiver.ByExtension(_filename)

	if err != nil {
		return err
	}

	err = archiveFormat(&arcFileObj, _password, OverwriteExisting)

	if err != nil {
		return err
	}

	var arcWalker, ok = arcFileObj.(archiver.Walker)
	if !ok {
		return fmt.Errorf("some error occured while reading the archive")
	}

	return startUnpackingCommonArchives(arc, arcWalker, ph)
}

func StartUnpacking(meta *ArchiveMeta, pack *ArchiveUnpack, ph *ProgressHandler) error {
	_meta := *meta
	_pack := *pack

	var arcUnpackObj ArchiveUnpacker

	// check whether the archive is encrypted
	// if yes, check whether the password is valid
	iae, err := IsArchiveEncrypted(meta)

	if err != nil {
		return err
	}

	if iae.IsEncrypted && !iae.IsValidPassword {
		return fmt.Errorf("invalid password")
	}

	ext := filepath.Ext(_meta.Filename)

	switch ext {
	case ".zip":
		arcUnpackObj = zipArchive{meta: _meta, unpack: _pack}

		break

	default:
		arcUnpackObj = commonArchive{meta: _meta, unpack: _pack}

		break
	}

	return arcUnpackObj.doUnpack(ph)
}
