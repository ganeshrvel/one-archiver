package onearchiver

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestUtils(t *testing.T) {
	Convey("Test extensions", t, func() {
		filenameList := []string{"abc.txt", "xyz.gz", "123", "123.tar.gz", ".ssh", ".gitignore"}
		extList := []string{"txt", "gz", "", "tar.gz", "ssh", "gitignore"}

		for i, f := range filenameList {
			ext := extension(f)

			So(extList[i], ShouldEqual, ext)
		}
	})
}
