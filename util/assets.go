// Code generated by go-bindata.
// sources:
// assets/app.js
// assets/index.html
// assets/picodom.js
// assets/styles.css
// DO NOT EDIT!

package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() any {
	return nil
}

var _appJs = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x55\xcd\x6e\xdb\x3c\x10\x3c\x47\x4f\x31\x39\x51\x82\x65\x21\x67\xeb\x13\x3e\xa0\x48\x0f\xe9\x21\x3d\xb4\x3d\x05\x41\xc1\x50\x6b\x9b\xb0\x44\x0a\x14\xe5\x24\x30\xf4\xee\x05\x29\xea\xc7\x89\x5d\xd4\x17\x8b\xdc\xdd\x59\xee\x70\x76\xc9\xba\x96\xd0\x5a\x23\x85\x65\x79\x14\x6d\x3b\x25\xac\xd4\x0a\xbf\x1e\x62\x69\xa9\x6e\x13\x9c\x22\xe0\xc8\x0d\xf6\x28\xd0\x48\xa1\x4b\x5d\x67\xfb\x3c\x02\x26\xdf\xb6\x7b\xa9\xa5\x8d\x69\xf0\x1d\xbc\xa9\x42\x81\x52\x8b\xae\x26\x65\xb3\x1d\xd9\xaf\x15\xb9\xcf\x2f\xef\x0f\x65\xcc\x2c\x6f\x0f\x6b\xc5\x6b\x5a\x4b\xd5\x74\x96\x25\xb9\x0f\x34\x8d\xc8\x78\x59\xfe\xe4\xed\x21\xa6\x2a\x3b\xf2\xaa\xa3\x60\x1a\x97\x28\xc0\x98\xdb\xea\x97\x27\x10\x15\x71\xe3\xc2\xda\x38\xc1\xc9\xe3\xf8\xad\x7b\xad\x28\x6c\xe7\xe7\x11\x35\x37\x07\x9f\x47\x26\x27\x43\xb6\x33\x6a\xb2\x4d\x10\xb3\x4f\x8a\x5b\xcf\xc6\x93\x7c\xce\x4a\xad\xc8\xa1\xf5\x79\x14\x98\x71\xd5\x3c\x38\x33\x0a\x3c\x3d\x7b\x6a\xb4\x41\xec\x4c\x12\x05\xee\x72\x48\xfc\x07\x0f\x90\x55\xa4\x76\x76\x9f\x43\xae\x56\x4b\xba\xc4\x9e\xc4\x81\x4a\x14\x88\xcf\x12\xe1\x7f\xb0\x60\x63\xd8\x80\x75\x6a\x5c\x05\x5e\xa6\xdc\x59\xd3\xb5\xfb\x38\xba\xd9\xc7\xac\x94\x47\x96\xe2\x24\x2a\xde\xb6\x8f\xbc\x26\x17\xe8\x19\x77\xd0\x60\x58\x8d\xe9\x52\x68\x25\x2a\x29\x0e\xd8\x2c\x09\xe9\xd3\xe8\x06\x98\xce\xe1\xee\x29\x49\x06\xce\x23\x20\xb0\x75\x25\x8f\xd0\xca\x72\xa9\xc8\xb0\x01\xc5\xf9\x6d\xb5\xa9\x3f\x1f\x88\xde\xec\x70\xfb\xeb\x57\xc3\x9b\x86\x0c\x73\xc7\x19\xb4\x84\x4d\x10\x55\x00\xf1\x30\x83\x54\x52\x9c\xc2\x16\x20\xcb\xa9\xb4\x85\x98\xd2\xc9\x7e\x25\xe1\xc2\xc3\xbe\x37\x93\x71\xb1\xcd\x3b\xab\xb7\x5a\x74\x2d\x36\xb0\xa6\xa3\x60\xe8\x93\x64\x2a\xea\x2a\xc9\x95\x6c\x2d\xeb\xd3\xf9\x66\xfe\x1e\xb3\xd5\xda\xce\x6c\x5d\x75\x7b\xb1\x6a\xed\x25\xbd\x76\xb8\x2d\x5b\x5e\xdd\xac\xfe\x7e\x2e\x81\xdd\x53\x45\x96\x20\x74\xdd\xb8\x8f\x92\x25\xee\x12\xfb\x28\x1a\xda\xd3\x77\x63\xee\x17\xba\x2a\x1f\x75\x49\xc3\xc2\x34\x02\x85\x57\xa6\x54\x47\x7d\x70\xc9\xa7\xc6\xe0\x66\xe7\x7a\xe3\x55\xaa\x52\xbf\x66\xf4\x66\xc9\x28\x5e\x65\x83\xe3\xef\xf8\xdb\x8f\xef\x8f\x99\x9b\x25\x6a\x27\xb7\xef\xde\xdb\x35\x4a\xea\xb1\xfc\xa5\x7e\x6a\xb1\x21\x34\x3e\x89\xda\x5f\xa5\x73\x63\xfd\x18\x54\xe9\xdd\x87\x98\xa9\x5f\xda\x69\x0e\x5c\x6e\x36\x6e\x76\x7e\xf2\x5c\x6c\x38\x40\x6e\x11\x4b\xdc\x16\xb8\x73\x7b\x37\x0e\xad\xc5\x0a\x0c\x01\x72\x18\x2f\xee\x37\x9a\x3e\xd7\x36\xe0\x3f\xc9\xe7\xd0\x89\xfd\x34\xc2\x3e\x14\x55\xe9\x1d\x4b\xe1\x14\xe6\x64\xdd\x0f\x9d\xe4\xea\x0b\xa3\x6e\x59\xa3\xef\xb5\xcb\xdc\x04\x6f\x96\x42\x0d\x9a\x70\x7f\x13\x57\xe7\xe3\xee\x1f\xa8\x3e\x0f\x98\x49\x1f\xe7\xc0\x12\x42\xaa\x92\xde\x52\xf8\xc9\x17\x38\xbc\x80\x38\x46\xb2\x14\x3e\x00\x1b\x2c\x02\xb1\xf1\x7f\x8b\xf2\x0d\xa9\x92\xcc\x59\x9e\xf9\xc5\x99\xe6\x4c\x50\xea\xe2\xf1\x69\xb8\x15\xfb\x38\x88\x36\xc5\xf8\x85\x62\x7e\xb4\x92\x74\x8c\x1b\xd3\xb9\x69\x1d\x74\xab\x55\xa5\xb9\x1b\xb5\x17\x28\x92\xd6\xbf\x14\x79\xf4\x27\x00\x00\xff\xff\x4b\x30\xde\x77\x18\x07\x00\x00")

func appJsBytes() ([]byte, error) {
	return bindataRead(
		_appJs,
		"app.js",
	)
}

func appJs() (*asset, error) {
	bytes, err := appJsBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "app.js", size: 1816, mode: os.FileMode(436), modTime: time.Unix(1505974565, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _indexHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x92\x41\x8b\xdb\x30\x10\x85\xcf\xde\x5f\x31\xd1\xb9\xb2\xce\x2d\xb2\xa1\x94\x1c\x72\xeb\x65\xa1\xb0\x84\xa2\x48\x13\x5b\xad\x2c\xa9\x9a\x71\x9c\x40\x7f\x7c\x91\xed\x42\xcb\xe6\x64\x3c\xf3\x34\xdf\x9b\x27\xe9\x83\x4b\x96\x1f\x19\x61\xe4\x29\xf4\x2f\x7a\xfb\x34\x7a\x44\xe3\xfa\x97\xa6\xd1\x13\xb2\x81\x91\x39\x4b\xfc\x35\xfb\x5b\x27\xbe\xc9\xd7\xcf\xf2\x4b\x9a\xb2\x61\x7f\x09\x28\xc0\xa6\xc8\x18\xb9\x13\xa7\x63\x87\x6e\x40\xd1\x43\x3d\x18\x7c\xfc\x09\x05\x43\x27\x88\x1f\x01\x69\x44\x64\x01\x95\xd5\x09\xc6\x3b\x2b\x4b\x24\x60\x2c\x78\xfd\xab\x68\x6b\xa5\xc2\xd5\x4e\xd7\x97\xe4\x1e\xab\x8b\x83\x94\x6f\xfe\x0a\x81\xe1\x74\x84\x8f\xe7\xb5\xe6\xfc\x0d\x6c\x30\x44\x9d\xf0\x28\xe7\x3c\x14\xe3\x50\x56\x37\xc6\x47\x2c\x75\x52\xd3\xe8\xfc\x44\x33\x21\x91\xa9\x46\xbf\x06\x34\x84\x1f\x60\x6f\xc0\x29\x32\x96\x88\x0c\xc7\x7b\x0e\xa9\x60\x01\x4e\xeb\x7e\x3e\xce\x08\x33\xf9\x38\x00\x8f\x9e\x80\xd2\x95\x17\x53\xb0\xd5\x2a\x6f\x1c\xf3\x84\x53\x23\x10\xc0\xa6\x0c\xc8\x9d\xf8\x7e\x09\xa6\xfe\x6f\x2b\xd7\x48\xe9\x93\x52\xcb\xb2\xb4\x93\xb7\x25\xd5\x91\xad\x4d\x93\xc2\x28\x67\x52\x2e\x2d\x31\x24\xe3\x94\xdf\x3d\x49\xdc\x3d\xb5\x86\xf2\x5d\xf4\xaf\x1b\x44\x2b\xb3\xa6\xa1\x9c\xbf\x6d\x51\xbd\x61\x74\xfe\x7a\x96\xf2\xdf\xe4\x06\xc6\x35\x3a\xf8\x0d\x87\xd3\x11\xce\x3d\xd4\xce\xaa\x20\x5b\x7c\x66\xa0\x62\x3b\x91\xbd\x4d\x2e\x4d\xed\x0f\x12\xbd\x56\x5b\xe7\x9d\xc8\xe4\xfc\x5e\xf0\x1f\x56\xab\xed\xe6\xb4\xda\xde\xd3\x9f\x00\x00\x00\xff\xff\x0b\x97\x9a\x0c\x67\x02\x00\x00")

func indexHtmlBytes() ([]byte, error) {
	return bindataRead(
		_indexHtml,
		"index.html",
	)
}

func indexHtml() (*asset, error) {
	bytes, err := indexHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "index.html", size: 615, mode: os.FileMode(436), modTime: time.Unix(1505973140, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _picodomJs = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x56\x4d\x93\xa3\x36\x10\xbd\xe7\x57\xd8\xa4\x8a\x92\xca\xbd\x8c\x77\x73\x33\xdb\x71\xcd\xa6\x72\x4b\xf6\x90\xaf\x0b\xc5\x01\x43\x83\x59\x63\x89\x12\xcd\xd8\x2e\xc3\x7f\x4f\x49\xc2\xc6\x33\x93\xcd\xc5\x0d\xea\xee\xc7\xd3\xeb\x56\xcb\xcb\xb2\x57\x39\xd7\x5a\x09\x02\x96\xd7\x40\xef\xbe\x51\xce\x01\x22\x5f\x5a\xd2\xe5\x82\xce\xad\x36\xdc\x85\x61\xd0\xab\x82\xca\x5a\x51\x11\x2c\x6f\xce\xa3\x2e\xfa\x86\xb6\x2c\xa6\x28\xb9\x09\x6e\x70\x33\x82\xcf\x0a\x43\x6f\xa3\xec\x58\x6c\xfd\xa3\x48\x82\x29\x2f\x48\x81\xe5\x86\x05\x45\x6d\x9d\xeb\x42\x1f\xf1\x3a\xca\x51\xf0\xbe\xee\x60\xe6\x27\xaf\x41\xdf\xd1\xa2\x63\x53\xe7\x1c\xc4\x37\xc7\x82\x3d\xf5\x97\xcc\x2c\x14\x18\x4c\xd2\xb8\xd4\x46\xf4\x98\x99\xaa\x3f\x92\xe2\x2e\x6a\x48\x55\xbc\x8f\xfb\x0f\x1f\x16\x3f\x7f\x8a\x65\x11\xb5\x7d\xb7\x17\x77\x7f\xd2\xa7\xd2\xe5\xc4\xc5\x2d\x54\xd6\xa5\x78\x36\x26\xbb\x44\x75\xe7\xac\x50\x58\x44\xad\x6e\x85\x94\xd2\xc3\xab\x07\xd8\x3b\xa6\x72\x58\xd4\x74\xb4\x50\x7d\xd3\x2c\x51\x85\xe1\x72\xbd\x44\x67\x3f\x7a\x2b\x02\xd5\x1f\x77\x64\x66\x89\xec\xa2\x5a\x61\x10\x48\x30\x13\x8e\x94\xb1\x21\xee\x8d\x0a\xec\x7e\x55\xf5\x50\x91\xed\x95\xb3\x6a\x43\x50\x64\x9c\x6d\x78\x18\xae\x23\xe4\xfb\xba\x29\x0c\xa9\x8d\x19\x37\x24\x18\x8c\x1c\xef\xf2\xb8\xca\x42\x06\x0d\xf4\x50\xc8\x6b\x5d\x0a\xcb\x0c\x31\x93\x8c\x14\xd5\xaa\x23\xc3\x5f\xa8\xd4\x86\x84\x16\x0d\xf4\x12\x78\xda\xc1\x14\xba\xc4\x26\xe2\xac\x0a\x43\x67\x10\x31\xb3\x56\x5e\x6b\xc1\x90\x45\x96\x05\x34\xce\x48\xe8\xb1\x1f\x86\xa0\x7b\xb1\x74\x7d\x96\xd3\xd5\x96\x26\xc7\x26\xba\xd1\x9c\x94\x83\x0e\xb3\x77\x6b\x7b\xbc\x8e\xf0\x82\x49\x0a\xad\x7d\xaa\x70\x1d\x57\x9f\xbb\xb8\x5a\xad\x7c\x89\x2f\xf8\x92\x54\x29\xb2\x4f\xfc\xaa\x0b\xea\x92\x2a\x85\xe3\x03\x94\x7d\xdf\xa1\x11\x47\x19\x7b\xfe\xbb\x30\x14\xfb\x64\x97\x62\x72\x81\x63\x2a\xc7\x1b\xa7\x0a\xd7\x70\xc0\x75\x7c\xf8\x9c\xc7\x8f\xe8\xef\xd0\x4e\x0f\xe4\x93\xc3\x1d\xbd\x2e\x45\x9b\xec\x52\x59\xad\x56\x4e\x31\x07\x71\x46\x23\x4e\x12\x9e\x71\x9f\x9c\xd3\x61\x48\xd2\xd8\xeb\x7d\xde\x4e\xc2\x5b\x36\x4a\x30\x5c\xe0\x08\x27\xab\xf7\x61\xb5\x92\x60\x37\xb8\x11\x3b\xf4\x81\x82\xe1\x39\x59\xa7\xf0\x9c\x7c\x4c\x7d\x90\xf3\xdb\xb5\xad\xe0\xd7\x55\x73\x81\x17\x09\xef\x93\xe4\xc6\x7f\xc7\x7e\xf7\xfe\x29\x68\x93\x73\x8a\x27\x39\x3a\x19\x9c\xb8\x7e\xef\xff\x23\xa1\x23\x5d\x0a\x06\xaf\xce\x54\xed\x6a\xb5\x9a\xa5\x5c\xd4\x6a\xb1\xf7\x40\x76\xeb\x55\x0a\x5f\xd0\x12\x89\xdb\xe4\x8b\x8b\x8f\x0e\x74\x49\x87\xa1\xbc\xb1\xf4\xab\x72\x1c\x5d\xaf\x71\x18\x36\x4b\x44\x8e\x94\x2e\xe8\x9f\xac\xe9\x29\x0c\xc5\xf7\x1a\xb4\x40\x96\x50\x0a\x82\x62\x6a\xc0\xfb\x69\x59\xf0\xdc\xf8\xc6\x8e\x8c\xba\x14\x16\x89\x90\xa6\xc0\x29\x8e\x2c\x9d\x39\x36\x7b\x98\x21\x78\x1d\xef\x5d\x6b\xec\xb6\x48\xaa\xc4\xa4\x48\x89\x49\x5f\x3b\xd8\x3b\xd8\x3a\x26\x5c\x35\x63\x6a\x8f\x59\x97\xe2\xfd\x19\x96\xfe\x4b\x85\xce\xdd\x08\x8a\x72\x43\x19\xd3\x5f\x74\x66\xdb\xd1\x82\xe4\xdc\x4f\x0a\x05\x23\xcf\xc7\x8a\xdc\xe9\xdb\xbe\x49\xfd\xb5\x21\xfb\xf6\xf5\x4f\x11\xec\x99\xdb\xcd\xd3\xd3\xe9\x74\x8a\x4e\x3f\x45\xda\x54\x4f\x9f\xd6\xeb\xf5\x93\x4d\x07\x9f\xbc\xf9\xef\x64\xe1\xbd\xb1\x97\x2a\x0c\xbd\x8d\xb4\xf2\x51\x61\x98\xfb\xe1\x74\x1f\xc9\xf2\xfa\x26\x44\x28\x39\xca\x37\xe2\x79\xdd\x1b\xa1\xc0\x80\x7f\x49\x4c\xfa\x10\x84\xeb\xd8\x7c\xa6\xb7\x73\x20\x96\x2a\xca\xda\x96\x54\xf1\x8b\x75\x08\x2d\xe6\x90\xc4\xac\x56\xf6\xc6\x90\xe3\x7b\xd1\x6b\x37\xed\x94\xbc\xbe\x22\x91\x09\xbb\xe6\xeb\xab\xd1\x16\x0d\x6a\x0c\x5e\x6c\x9b\x59\x49\xcd\x30\x04\xf9\x9e\xf2\x03\x15\xee\x75\x6b\x4b\xbd\x71\x65\xd5\x4b\xc4\x3a\x0c\x1b\x41\x60\x40\x43\x2d\x47\x15\x86\x2a\xd2\xaa\x6f\x8b\xef\x68\x32\x7b\x5d\x03\x8c\x0f\xb3\xb8\xbc\xb1\x9b\x40\x0c\x1d\xf5\x0b\x6d\xe7\x47\xc1\x72\x43\x91\x7f\xf6\x1b\xe7\x87\xf4\xc6\xa7\x83\xf1\x5d\x75\xa0\x8b\xa5\xfb\x30\xae\x83\x8e\x2f\x0d\xf9\xc5\x9b\x00\xda\x0b\x60\x40\xa1\xb2\x97\x85\x94\x14\xb9\xb0\x44\xa7\xa8\x12\x9d\x0e\x43\x10\xf8\x76\x63\x73\xb9\x52\xc2\x29\xaa\x31\xcf\x38\xdf\xdb\x03\x34\xce\x57\xfa\xf2\xd5\x7d\xb5\xa5\xa8\x23\x7e\x66\x36\xf5\xae\x67\x72\x0a\xdf\xb9\x3f\x2c\x4b\x39\x8e\x96\x47\x0f\x85\x9d\xea\xb9\xfd\xe9\xf0\xf1\x6f\x07\x18\xc8\x40\xfb\x92\x19\x54\x22\x1b\x86\x7b\x83\xee\x74\x71\xb1\x8d\x63\x2f\x25\x8d\xb9\xbf\x82\x63\xa9\xc5\xfd\xc4\x9b\x31\xa6\x68\x8f\x0c\x14\xb5\x96\x34\x76\xa3\x8c\x7f\x78\x7a\xfa\x71\xd1\xe9\xde\xe4\xf4\x7b\xd6\xb6\xb5\xaa\xfe\xfe\xe3\x37\x9c\xfe\x5d\x44\xdf\xba\xe8\x98\xb5\xff\x06\x00\x00\xff\xff\x75\xeb\xaa\x7d\xfb\x08\x00\x00")

func picodomJsBytes() ([]byte, error) {
	return bindataRead(
		_picodomJs,
		"picodom.js",
	)
}

func picodomJs() (*asset, error) {
	bytes, err := picodomJsBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "picodom.js", size: 2299, mode: os.FileMode(436), modTime: time.Unix(1505973119, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _stylesCss = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x54\xef\x6e\x9b\x30\x10\xff\x0c\x4f\x61\x69\x9a\xb4\x4d\x75\x44\x68\xab\x66\xe6\xd3\x1e\xc5\x36\x07\x58\x31\x3e\x64\x9b\x25\xd9\xd4\x77\x9f\x6c\x93\x00\x25\x6d\x17\x29\x08\x1b\xdf\xfd\xfe\xdc\x9d\x7f\x90\xbf\x79\xd6\x73\xdb\x2a\xc3\x48\x51\xe5\xd9\xc0\xeb\x5a\x99\x36\x2d\x04\x9e\xa9\x53\x7f\xe2\x5a\xa0\xad\xc1\x52\x81\xe7\x2a\xcf\x1a\x34\x3e\x7c\x01\x46\xca\xc3\x70\xdb\x69\x78\xaf\xf4\x85\x11\xc7\x8d\xa3\x0e\xac\x6a\xaa\xfc\x35\xcf\x3b\xdf\xeb\x07\x22\xb0\xbe\x04\xb4\x0e\x54\xdb\x79\x46\xf6\x45\xf1\xb5\xca\x33\xfc\x0d\xb6\xd1\x78\x62\xc4\xa0\x81\x78\x7e\xa7\x80\x8e\x43\x6b\x79\x0d\x54\xa2\xf1\x5c\x19\xb0\x21\xf4\xa4\x6a\xdf\xdd\x22\xdf\x24\x5a\x31\xf8\x65\x15\xd7\x0f\x2b\x22\x4b\xd2\x8f\x65\x24\x2d\x51\xa3\x65\xe4\x4b\x13\x7f\x41\x30\x97\xc7\xd6\xe2\x68\x6a\x7a\xfd\xb6\x07\x21\x00\x96\xce\x3c\x42\x4f\xf6\xf3\x3f\x50\x5e\x32\xd6\xca\x1c\x97\xae\x96\xd0\x93\x22\x1c\x5b\x78\x9b\xe2\xb2\x0d\xc6\x1d\xfc\x1b\xb7\x48\xff\x34\x69\x16\xa8\xeb\x2a\xcf\x3c\x9c\x3d\xe5\x5a\xb5\x86\x11\x09\xc6\x83\xad\xf2\xac\x56\x6e\xd0\xfc\xc2\x88\xd0\x28\x8f\xd5\x7b\xb6\x95\x91\x82\x56\x06\xe8\x7a\x2b\x26\xf5\x96\x1b\xd7\xa0\xed\x19\x19\x87\x01\xac\xe4\x6e\x2a\xce\xff\x57\xe4\x8e\x9a\x9f\xb2\x7c\x11\x45\x4a\x14\x71\x94\x19\x46\x4f\x4f\x96\x07\x90\x90\x71\x36\x69\xf7\x9c\x5c\x43\xa7\xbc\x42\xc3\x48\xa3\xce\x10\x55\xe3\x90\xfa\x53\x43\xe3\xd3\x9b\x55\x6d\x17\xdf\xd6\x89\xb7\x14\x57\x7a\xf7\x57\x88\xb9\x30\xc5\x2e\x99\xf0\xf6\x08\x8e\x3e\x84\x5e\xbb\x34\x4b\xe3\x70\x5b\x5e\xf5\x3d\xf1\xfd\xd3\x41\xde\xd5\x6e\x5b\xc1\xbf\x95\xcf\xcf\x0f\x64\x7e\x14\xbb\xc3\xcb\xf7\xd8\x40\x33\x67\xd6\xa0\x1c\x5d\x60\xfe\x41\x37\x44\x99\xdc\x1d\xa9\x56\x2e\xaa\xbc\x8e\x11\xbd\x30\xc2\x47\x8f\xef\x1b\x57\x4e\x8a\x04\x7a\x8f\x7d\x50\x38\x75\xc2\xca\x4c\x3f\xbb\x19\x60\x94\x87\x7e\x35\xbc\x53\x92\xcf\xb5\x6d\x0b\x2a\x47\xeb\x42\xd0\x80\x2a\x75\xec\xeb\x02\x64\x27\x3b\x90\x47\xa8\x03\x58\xf4\xa4\x06\x89\x96\x27\x1d\xb1\x78\xbe\xb3\x38\xb6\xdd\x67\xe0\x8f\x87\x64\x6c\x83\xe8\xa7\xce\xda\xf8\x31\x4b\xbe\x9a\xb1\x96\xff\xe1\x3c\x6e\x5b\x7a\x27\xbc\xa1\x52\x03\xb7\x34\xe8\x71\xdb\xee\xbb\x3b\xaf\x8b\x7b\x69\x9f\x2e\xd3\xdb\x3c\x4e\x8e\xad\x87\x74\xda\xfc\x60\x4c\xef\x59\xfc\x2f\x00\x00\xff\xff\x7f\xd5\x66\xf8\xea\x05\x00\x00")

func stylesCssBytes() ([]byte, error) {
	return bindataRead(
		_stylesCss,
		"styles.css",
	)
}

func stylesCss() (*asset, error) {
	bytes, err := stylesCssBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "styles.css", size: 1514, mode: os.FileMode(436), modTime: time.Unix(1505976763, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"app.js":     appJs,
	"index.html": indexHtml,
	"picodom.js": picodomJs,
	"styles.css": stylesCss,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"app.js":     {appJs, map[string]*bintree{}},
	"index.html": {indexHtml, map[string]*bintree{}},
	"picodom.js": {picodomJs, map[string]*bintree{}},
	"styles.css": {stylesCss, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
