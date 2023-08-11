package onearchiver

import (
	"fmt"
	"github.com/ganeshrvel/archiver"
	"github.com/yeka/zip"
	"io"
	"os"
	"path/filepath"
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

func (arc zipArchive) isEncrypted() (EncryptedArchiveInfo, error) {
	filename := arc.meta.Filename
	password := arc.meta.Password

	ai := EncryptedArchiveInfo{
		IsEncrypted:     false,
		IsValidPassword: false,
	}

	reader, err := zip.OpenReader(filename)
	if err != nil {
		return ai, err
	}
	defer func() {
		if err = reader.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	for _, file := range reader.File {
		if file.IsEncrypted() {
			ai.IsEncrypted = true

			file.SetPassword(password)

			r, err := file.Open()
			if err != nil {
				return ai, err
			}
			defer func() {
				if err = r.Close(); err != nil {
					fmt.Printf("%v\n", err)
				}
			}()

			_, err = io.ReadAll(r)
			if err != nil {
				return ai, nil
			}

			ai.IsValidPassword = true

			return ai, err
		}
	}

	return ai, err
}

func (arc commonArchive) isEncrypted() (EncryptedArchiveInfo, error) {
	filename := arc.meta.Filename
	password := arc.meta.Password

	ai := EncryptedArchiveInfo{
		IsEncrypted:     false,
		IsValidPassword: false,
	}

	arcFileObj, err := archiver.ByExtension(filename)

	if err != nil {
		return ai, err
	}

	err = archiveFormat(&arcFileObj, password, OverwriteExisting)

	if err != nil {
		return ai, err
	}

	switch arcValues := arcFileObj.(type) {
	case *archiver.Rar:
		// check if the rar file is encrypted
		r1, err := isRarArchiveEncrypted(arcValues, filename, "")
		if err != nil {
			return ai, err
		}

		// check if the password is correct
		if r1 {
			ai.IsEncrypted = true

			r2, err := isRarArchiveEncrypted(arcValues, filename, password)
			ai.IsValidPassword = !r2

			if err != nil {
				return ai, err
			}
		}

		return ai, err

	default:
		return ai, nil
	}
}

func IsArchiveEncrypted(meta *ArchiveMeta) (EncryptedArchiveInfo, error) {
	_meta := *meta

	var utilsObj ArchiveUtils

	ext := filepath.Ext(_meta.Filename)

	switch ext {
	case ".zip":
		utilsObj = zipArchive{meta: _meta}

		break

	case ".rar":
		utilsObj = commonArchive{meta: _meta}

		break
	// todo add 7 zip here
	default:
		ai := EncryptedArchiveInfo{
			IsEncrypted:     false,
			IsValidPassword: false,
		}

		return ai, nil
	}

	return utilsObj.isEncrypted()
}
