package onearchiver

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestUtils(t *testing.T) {
	Convey("Test extensions", t, func() {
		filenameList := []string{"abc.txt", "xyz.gz", "123", "123.tar.gz", ".ssh", ".gitignore", "github.com/ganeshrvel/one-archiver/e2e_list_test.go", "one-archiver/e2e_list_test.go", "e2e_list_test.go/.go.psd"}
		extList := []string{"txt", "gz", "", "tar.gz", "ssh", "gitignore", "go", "go", "go.psd"}

		for i, f := range filenameList {
			ext := extension(f)

			So(extList[i], ShouldEqual, ext)
		}
	})
}
