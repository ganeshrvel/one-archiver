package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	"os"
	"strings"
)

func isRarArchiveEncrypted(arcValues *archiver.Rar, filename, password string) (bool, error) {
	arcValues.Password = password

	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	err = arcValues.Open(file, 0)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := arcValues.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	r, err := arcValues.Read()
	defer func() {
		if r == (archiver.File{}) {
			return
		}

		if err := r.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	if err != nil {
		if strings.Contains(err.Error(), "incorrect password") {
			return true, nil
		}

		return false, err
	}

	return false, nil
}

func (arc commonArchive) prepare() (PrepareArchiveInfo, error) {
	filename := arc.meta.Filename
	pctx := arc.read.passwordContext()

	prepArcInf := PrepareArchiveInfo{
		IsValidPassword:      false,
		IsSinglePasswordMode: false,
		IsPasswordRequired:   false,
	}

	arcFileObj, err := archiver.ByExtension(filename)

	if err != nil {
		return prepArcInf, err
	}

	err = archiveFormat(&arcFileObj, pctx, OverwriteExisting)

	if err != nil {
		return prepArcInf, err
	}

	switch arcValues := arcFileObj.(type) {
	case *archiver.Rar:
		// check if the rar file is encrypted
		r1, err := isRarArchiveEncrypted(arcValues, filename, "")
		if err != nil {
			return prepArcInf, err
		}

		// check if the password is correct
		if r1 {
			prepArcInf.IsPasswordRequired = true
			prepArcInf.IsSinglePasswordMode = true

			r2, err := isRarArchiveEncrypted(arcValues, filename, pctx.getSinglePassword())
			prepArcInf.IsValidPassword = !r2

			if err != nil {
				return prepArcInf, err
			}
		}

		return prepArcInf, err

	default:
		return prepArcInf, nil
	}
}

func PrepareArchive(meta *ArchiveMeta, passwords []string) (PrepareArchiveInfo, error) {
	_meta := *meta

	var utilsObj ArchiveUtils

	ext := extension(_meta.Filename)

	switch ext {
	case "rar":
		utilsObj = commonArchive{meta: _meta, read: ArchiveRead{Passwords: passwords}}

		break
	default:
		ai := PrepareArchiveInfo{
			IsSinglePasswordMode: false,
			IsPasswordRequired:   false,
			IsValidPassword:      false,
		}

		return ai, nil
	}

	return utilsObj.prepare()
}
