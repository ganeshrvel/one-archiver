package onearchiver_test

import (
	. "github.com/ganeshrvel/one-archiver"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestUtils(t *testing.T) {
	Convey("Test extension", t, func() {
		type s struct {
			filename, ext string
		}

		sl := []s{
			s{
				filename: "",
				ext:      "",
			}, s{
				filename: "abc.xyz.tar.gz",
				ext:      "tar.gz",
			}, s{
				filename: "abc.xyz.tar.tar",
				ext:      "tar.tar",
			}, s{
				filename: "xyz.tar.gz",
				ext:      "tar.gz",
			}, s{
				filename: "tar.gz",
				ext:      "gz",
			}, s{
				filename: "abc.gz",
				ext:      "gz",
			}, s{
				filename: ".gz",
				ext:      "gz",
			}, s{
				filename: ".tar",
				ext:      "tar",
			}, s{
				filename: ".tar.gz",
				ext:      "tar.gz",
			}, s{
				filename: "tar.tar.gz",
				ext:      "tar.gz",
			}, s{
				filename: ".htaccess",
				ext:      "htaccess",
			}, s{
				filename: "abc.txt",
				ext:      "txt",
			}, s{
				filename: "abc",
				ext:      "",
			}, s{
				filename: "github.com/ganeshrvel/one-archiver/e2e_list_test.go",
				ext:      "go",
			}, s{
				filename: "one-archiver/e2e_list_test.go",
				ext:      "go",
			}, s{
				filename: "e2e_list_test.go/.go.psd",
				ext:      "psd",
			},
		}

		for _, f := range sl {
			ext := Extension(f.filename)

			So(ext, ShouldEqual, f.ext)
		}
	})

	Convey("Test parent directory", t, func() {
		type s struct {
			filename, parentPath string
		}

		sl := []s{
			s{

				filename:   "",
				parentPath: "",
			}, s{

				filename:   "/",
				parentPath: "/",
			}, s{

				filename:   "/1234",
				parentPath: "/",
			}, s{

				filename:   "/1234/",
				parentPath: "/",
			}, s{

				filename:   "/1234//",
				parentPath: "/",
			}, s{

				filename:   "/1234/abc/",
				parentPath: "/1234/",
			}, s{

				filename:   "/1234/abc",
				parentPath: "/1234/",
			}, s{

				filename:   "Hey.zip",
				parentPath: "",
			}, s{

				filename:   "/Hey.zip",
				parentPath: "/",
			}, s{

				filename:   "/Hey.zip/",
				parentPath: "/",
			}, s{

				filename:   "/Hey.zip/1234",
				parentPath: "/Hey.zip/",
			}, s{

				filename:   "/Hey.zip/123/",
				parentPath: "/Hey.zip/",
			}, s{

				filename:   "/Hey.zip/123//",
				parentPath: "/Hey.zip/",
			},
		}

		for _, f := range sl {
			p := GetParentDirectory(f.filename)

			So(p, ShouldEqual, f.parentPath)
		}
	})
}
