package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

// 全局配置变量
var (
	configFile string
	endpoint   string
	username   string
	password   string
	srcDir     string
	dstDir     string

	lookupProcess = 0
)

func init() {
	// 定义命令行参数
	flag.StringVar(&configFile, "config", "", "配置文件路径")
	flag.StringVar(&endpoint, "endpoint", endpoint, "Alist服务地址")
	flag.StringVar(&username, "username", username, "Alist用户名")
	flag.StringVar(&password, "password", password, "Alist密码")
	flag.StringVar(&srcDir, "src-dir", srcDir, "本地数据目录")
	flag.StringVar(&dstDir, "dst-dir", dstDir, "远程数据目录")
}

func main() {
	flag.Parse()

	if configFile != "" {
		loadConfig(configFile)
	}

	sync()
}

func loadConfig(filePath string) {
	log.Printf("Loading config from file: %s", filePath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		return
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Error parsing config file: %v", err)
		return
	}

	log.Printf("Loaded config: %+v", config)
	// 只有当配置文件中存在值时才覆盖默认值
	if config.Endpoint != "" {
		endpoint = config.Endpoint
	}
	if config.Username != "" {
		username = config.Username
	}
	if config.Password != "" {
		password = config.Password
	}
	if config.SrcDir != "" {
		srcDir = config.SrcDir
	}
	if config.DstDri != "" {
		dstDir = config.DstDri
	}

	if srcDir == "" || dstDir == "" {
		log.Printf("Error: src-dir or dst-dir are empty")
	}

	log.Printf("Config loaded successfully")
}

func sync() {
	alistClient := NewAlistClient(context.TODO(), endpoint, username, password)
	alistClient.Login()

	lookupProcess = 0
	log.Printf("Start to lookup dst files")
	remoteFilesMap, err := lookupFiles(dstDir, alistClient)
	if err != nil {
		log.Fatalf("Error listing remote filesystems: %v", err)
	}
	log.Writer().Write([]byte{'\n'})
	log.Printf("Total %d dst files found", len(remoteFilesMap))

	lookupProcess = 0
	log.Printf("Start to lookup src files")
	localFilesMap, err := lookupFiles(srcDir, alistClient)
	if err != nil {
		log.Fatalf("Failed to lookup local files: %v", err)
	}
	log.Writer().Write([]byte{'\n'})
	log.Printf("Total %d src files found", len(localFilesMap))

	deleteRemoteFiles(alistClient, remoteFilesMap, localFilesMap)
	uploadLocalFiles(alistClient, remoteFilesMap, localFilesMap)
}

func uploadLocalFiles(alistClient *AlistClient, remoteFilesMap, localFilesMap map[string]*FSListContentItem) {
	needUploadFiles := map[string][]string{}
	for rp := range localFilesMap {
		pureFilePath := strings.TrimPrefix(rp, srcDir)
		remoteFilePath := path.Join(dstDir, pureFilePath)

		if _, ok := remoteFilesMap[remoteFilePath]; !ok {
			dir, name := path.Split(pureFilePath)
			if needUploadFiles[dir] == nil {
				needUploadFiles[dir] = []string{}
			}
			needUploadFiles[dir] = append(needUploadFiles[dir], name)
			log.Printf("Local file %s%s not found in remote, add it to upload list. ", dir, name)
		}
	}

	log.Printf("Total %d src files need to upload", len(needUploadFiles))
	for dir, names := range needUploadFiles {
		src := path.Join(srcDir, dir)
		dst := path.Join(dstDir, dir)
		if err := alistClient.FSMkdir(dst); err != nil {
			log.Printf("Failed to mkdir %s: %v", dstDir, err)
		}
		if err := alistClient.FSCopy(src, dst, names); err != nil {
			log.Printf("Error uploading file %s in dir %s: %v", names, dir, err)
		}
	}

	log.Printf("============ Successfully uploaded %d files ============= ", len(needUploadFiles))
}

func deleteRemoteFiles(alistClient *AlistClient, remoteFilesMap, localFilesMap map[string]*FSListContentItem) {
	needDeleteFiles := map[string][]string{}
	for rp := range remoteFilesMap {
		localFilePath := path.Join(srcDir, strings.TrimPrefix(rp, dstDir))

		if _, ok := localFilesMap[localFilePath]; !ok {
			dir, name := path.Split(rp)
			if needDeleteFiles[dir] == nil {
				needDeleteFiles[dir] = []string{}
			}
			needDeleteFiles[dir] = append(needDeleteFiles[dir], name)
			log.Printf("Remote file %s%s not found in local, add it to delete list. ", dir, name)
		}
	}
	log.Printf("Total %d dst files need to delete", len(needDeleteFiles))

	for dir, names := range needDeleteFiles {
		if err := alistClient.FSRemove(dir, names); err != nil {
			log.Printf("Error deleting file %s in dir %s: %v", names, dir, err)
		}
	}
	log.Printf("============ Successfully deleted %d files ============= ", len(needDeleteFiles))
}

func lookupFiles(dir string, alistClient *AlistClient) (map[string]*FSListContentItem, error) {
	filesMap := make(map[string]*FSListContentItem)

	rResp, err := alistClient.FSList(dir)
	if err != nil {
		log.Println("Error listing remote filesystems:", err)
		return filesMap, err
	}

	for _, item := range rResp.Data.Content {
		if item.IsDir {
			subFilesMap, err := lookupFiles(path.Join(dir, item.Name), alistClient)
			if err != nil {
				log.Println("Error listing remote filesystems:", err)
				return filesMap, err
			}
			for k, v := range subFilesMap {
				lookupProcess++
				filesMap[k] = v
			}
			log.Writer().Write([]byte(fmt.Sprintf("\rLookup files: %d", lookupProcess)))
			continue
		}

		filesMap[path.Join(dir, item.Name)] = item
	}

	return filesMap, nil
}
