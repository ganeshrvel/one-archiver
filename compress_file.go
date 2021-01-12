package onearchiver

//import (
//	"github.com/ganeshrvel/archiver"
//	ignore "github.com/sabhiram/go-gitignore"
//	"os"
//)
//
//func compressFile(arc *commonArchive, newArchiveFile interface{ archiver.Compressor }, fileList *[]string) error {
//	_filename := arc.meta.filename
//	_gitIgnorePattern := arc.meta.gitIgnorePattern
//	_fileList := *fileList
//
//	var ignoreList []string
//	ignoreList = append(ignoreList, GlobalPatternDenylist...)
//	ignoreList = append(ignoreList, _gitIgnorePattern...)
//	ignoreListLength := len(ignoreList)
//
//	ignoreMatches, _ := ignore.CompileIgnoreLines(ignoreList...)
//
//	if !indexExists(_fileList, 0) {
//		return nil
//	}
//
//	_source := _fileList[0]
//
//	if ignoreListLength > 0 {
//	if ignoreMatches.MatchesPath(_source) {
//		return nil
//	}
//	}
//
//	in, err := os.Open(_source)
//	if err != nil {
//		return err
//	}
//	defer in.Close()
//
//	out, err := os.Create(_filename)
//	if err != nil {
//		return err
//	}
//	defer out.Close()
//
//	err = newArchiveFile.Compress(in, out)
//	if err != nil {
//		return err
//	}
//	defer out.Close()
//
//	return err
//}
