package pcloud

import (
	"errors"
	"net/url"
	"strconv"

	"encoding/json"
	"fmt"
	"net/http"

	"bytes"
	"io"
	"mime/multipart"
)

var is_debug bool = false

type CloudContent struct {
	Id             string
	Name           string
	IsFolder       bool
	ParentFolderId int64
	FolderId       int64
	FileId         int64
	Hash           int64
	Size           int64
}



type ResultFolderContents struct {
	ParentFolderId int64  `json:"parentfolderid"` //: 0,
	Id             string `json:"id"`             //: "d230807",
	Modified       string `json:"modified"`       //: "Wed, 02 Oct 2013 13:23:35 +0000",
	Path           string `json:"path"`           //: "/Simple Folder",
	Thumb          bool   `json:"thumb"`          //: false,
	Created        string `json:"created"`        //: "Wed, 02 Oct 2013 13:11:53 +0000",
	FolderId       int64  `json:"folderid"`       //: 230807,
	FileId         int64  `json:"fileid"`         //: 5948299758,
	IsMine         bool   `json:"ismine"`         //: true,
	IsShared       bool   `json:"isshared"`       //: false,
	IsFolder       bool   `json:"isfolder"`       //: true,
	Name           string `json:"name"`           //: "Simple Folder",
	Icon           string `json:"icon"`           //: "folder"
	ContentType    string `json:"contenttype"`    //: "audio/mpeg",
	Hash           int64  `json:"hash"`           //: 5380817599554757000,
	Category       int64  `json:"category"`       //: 3,
	Size           int64  `json:"size"`           //: 11252576
	Comments       int64  `json:"comments"`       //: 0,
	Contents []ResultFolderContents `json:"contents"` //: [{
}


/*
type ResultFolderMetadata struct {
	Icon     string                 `json:"icon"`     //: "folder"
	Id       string                 `json:"id"`       //: "d0",
	Modified string                 `json:"modified"` //: "Wed, 02 Oct 2013 13:23:35 +0000",
	Path     string                 `json:"path"`     //: "/",
	Thumb    bool                   `json:"thumb"`    //: false,
	Created  string                 `json:"created"`  //: "Wed, 02 Oct 2013 13:11:53 +0000",
	FolderId int64                  `json:"folderid"` //: 0,
	IsShared bool                   `json:"isshared"` //: false,
	IsFolder bool                   `json:"isfolder"` //: true,
	IsMine   bool                   `json:"ismine"`   //: true,
	Name     string                 `json:"name"`     //: "/",
	Contents []ResultFolderContents `json:"contents"` //: [{
}

type ResultFolderContents struct {
	ParentFolderId int64  `json:"parentfolderid"` //: 0,
	Id             string `json:"id"`             //: "d230807",
	Modified       string `json:"modified"`       //: "Wed, 02 Oct 2013 13:23:35 +0000",
	Path           string `json:"path"`           //: "/Simple Folder",
	Thumb          bool   `json:"thumb"`          //: false,
	Created        string `json:"created"`        //: "Wed, 02 Oct 2013 13:11:53 +0000",
	FolderId       int64  `json:"folderid"`       //: 230807,
	FileId         int64  `json:"fileid"`         //: 5948299758,
	IsMine         bool   `json:"ismine"`         //: true,
	IsShared       bool   `json:"isshared"`       //: false,
	IsFolder       bool   `json:"isfolder"`       //: true,
	Name           string `json:"name"`           //: "Simple Folder",
	Icon           string `json:"icon"`           //: "folder"
	ContentType    string `json:"contenttype"`    //: "audio/mpeg",
	Hash           int64  `json:"hash"`           //: 5380817599554757000,
	Category       int64  `json:"category"`       //: 3,
	Size           int64  `json:"size"`           //: 11252576
	Comments       int64  `json:"comments"`       //: 0,
}

*/


// checkResult; returned error if request is failed or server returned error
func checkResultMod(resp *http.Response, err error) ([]CloudContent, error) {
	var ContentList []CloudContent
	buf, err := convertToBuffer(resp, err)
	if err != nil {
		return nil, err
	}
	if is_debug {
		fmt.Println(buf)
	}
	result := struct {
		Result   int                  `json:"result"`
		Error    string               `json:"error"`
		Metadata ResultFolderContents `json:"metadata"`
	}{}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, err
	}
	if is_debug {
		fmt.Println(result.Metadata)
	}
	for _, cn := range result.Metadata.Contents {
		rez, _ := parseContent(&cn)
		ContentList = append(ContentList, rez)
		if is_debug {
			fmt.Println(rez)
		}
	}

	if result.Result != 0 {
		return nil, errors.New(result.Error)
	}

	return ContentList, nil
}

func parseContent(cn *ResultFolderContents) (rez CloudContent, err error) {
	rez.IsFolder = cn.IsFolder
	if cn.IsFolder {
		rez.FolderId = cn.FolderId
		rez.Id = cn.Id
		rez.Name = cn.Name
		rez.ParentFolderId = cn.ParentFolderId
	} else {
		rez.FileId = cn.FileId
		rez.Id = cn.Id
		rez.Name = cn.Name
		rez.ParentFolderId = cn.ParentFolderId
		rez.Hash = cn.Hash
		rez.Size = cn.Size
	}
	return
}

// ListFolder; https://docs.pcloud.com/methods/folder/listfolder.html
func (c *pCloudClient) ListFolder(path string, folderID int) ([]CloudContent, error) {
	values := url.Values{
		"auth": {*c.Auth},
	}

	switch {
	case path != "":
		values.Add("path", path)
	case folderID >= 0:
		values.Add("folderid", strconv.Itoa(folderID))
	default:
		return nil, errors.New("bad params")
	}
	return checkResultMod(c.Client.Get(urlBuilder("listfolder", values)))

}

// UploadFile; https://docs.pcloud.com/methods/file/uploadfile.html
func (c *pCloudClient) UploadFileMod(reader io.Reader, path string, folderID int, filename string, noPartial int, progressHash string, renameIfExists int) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	values := url.Values{
		"auth": {*c.Auth},
	}
	fmt.Println("auth", *c.Auth)
	switch {
	case path != "":
		values.Add("path", path)
	case folderID >= 0:
		values.Add("folderid", strconv.Itoa(folderID))
	default:
		return errors.New("bad params")
	}

	if filename == "" {
		return errors.New("bad params")
	}

	if noPartial > 0 {
		values.Add("nopartial", strconv.Itoa(noPartial))
	}
	if progressHash != "" {
		values.Add("progresshash", progressHash)
	}
	if renameIfExists > 0 {
		values.Add("renameifexists", strconv.Itoa(renameIfExists))
	}

	fw, err := w.CreateFormFile(filename, filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, reader); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", urlBuilder("uploadfile", values), &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	return checkResult(c.Client.Do(req))
}
