package mfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"../config"
)

type ContentType int8

var KB = uint64(1024)

type ContentInformation struct {
	ContentName,
	ContentFullRoot,
	ContentData string
	ContentSize int32
	Type        ContentType
}

type ContentFolderInformation struct {
	FName string
	FSize int64
}

const rootMsfDir string = "fcloud"
const (
	Directory ContentType = iota
	File      ContentType = iota
)

func init() {
	log.Println("Load MSF module...")
	if !checkDir(rootMsfDir) {
		log.Fatalf("Cannot create main root folder %s%s", config.GetRootLocation(), rootMsfDir)
	}
}

func OpenFileStream(currentDir, fileName string, userID int) (*os.File, error) {
	fullDir := fmt.Sprintf("%s/%s/%s", getUserFullDir(userID, false), currentDir, fileName)

	if currentDir == "/" {
		fullDir = fmt.Sprintf("%s%s/%s", getUserFullDir(userID, false), currentDir, fileName)
	}

	return os.OpenFile(fullDir, os.O_RDWR|os.O_CREATE, 0755)
}

func PrepareUserCloud(userID int) bool {
	return checkDir(fmt.Sprintf("%s/%d", rootMsfDir, userID))
}

func CreateContentOnUserCloud(userID int, content ContentInformation) bool {
	fullDir := fmt.Sprintf("%s/%s/%s", getUserFullDir(userID, false), content.ContentFullRoot, content.ContentName)

	if content.ContentFullRoot == "/" {
		fullDir = fmt.Sprintf("%s/%s", getUserFullDir(userID, false), content.ContentName)
	}

	var result bool = false
	if content.Type == Directory {
		result = checkDir(fullDir)
	} else {
		result = checkFile(fullDir, content.ContentData)
	}

	return result
}

func GetContentOnUserCloud(userID int, content ContentInformation) (bool, ContentInformation) {
	fullDir := fmt.Sprintf("%s/%s/%s", getUserFullDir(userID, false), content.ContentFullRoot, content.ContentName)
	var result ContentInformation = content
	var resultSuccess bool = false

	if result.Type == File {
		result.ContentData = readFile(fullDir)
		resultSuccess = (result.ContentData != "")
	} else {
		if content.ContentFullRoot == "/" {
			fullDir = fmt.Sprintf("%s", getUserFullDir(userID, false))
		} else {
			fullDir = fmt.Sprintf("%s/%s", getUserFullDir(userID, false), content.ContentFullRoot)
		}

		var resultList []ContentFolderInformation = []ContentFolderInformation{}
		if dirList, dirListErr := ioutil.ReadDir(fullDir); dirListErr == nil {
			var c int32 = 0
			for _, dir := range dirList {
				c++
				resultList = append(resultList, ContentFolderInformation{
					FName: dir.Name(),
					FSize: dir.Size() * 1024,
				})
			}

			if jsonList, jsonListErr := json.Marshal(resultList); jsonListErr == nil {
				resultSuccess = true
				result.ContentData = string(jsonList)
				result.ContentSize = c
			}
		}
	}

	return resultSuccess, result
}

func RemoveContentOnUserCloud(userID int, content ContentInformation) bool {
	fullDir := fmt.Sprintf("%s/%s/%s", getUserFullDir(userID, false), content.ContentFullRoot, content.ContentName)
	removeErr := os.RemoveAll(fullDir)

	return (removeErr == nil)
}

func readFile(filename string) string {
	var result string

	if newFile, newFileErr := ioutil.ReadFile(filename); newFileErr == nil {
		result = string(newFile)
	}

	return result
}

func checkFile(filename, content string) bool {
	var result bool = false

	if newFile, newFileErr := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755); newFileErr == nil {
		result = true
		newFile.WriteString(content)
		newFile.Close()
	}

	return result
}

func checkDir(dir string) bool {
	var result bool = false
	fullDir := fmt.Sprintf("%s%s", config.GetRootLocation(), dir)
	if _, err := os.Stat(fullDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(fullDir, 0755); err == nil {
				result = true
			}
		}
	} else {
		result = true
	}

	return result
}

func getUserFullDir(userID int, includeRoot bool) string {
	fullDir := fmt.Sprintf("%s%s/%d", config.GetRootLocation(), rootMsfDir, userID)

	if !includeRoot {
		fullDir = fmt.Sprintf("%s/%d", rootMsfDir, userID)
	}

	return fullDir
}
