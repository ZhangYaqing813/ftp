package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

type SWinfo struct {
	userIonf     string
	passWd       string
	ipAddr       string
	backFileName string
}

// 读取交换机相关信息
func readSWadd(fileName string) (swinfo []SWinfo) {
	swinfo = make([]SWinfo, 64)

	f, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		log.Printf("open file %s err %s ", fileName, err)
	}
	defer f.Close()
	buf := make([]byte, 0, 1024)

	buf, err = io.ReadAll(f)
	if err != nil {
		log.Printf("io reader err  %s  ", err)
	}
	fmt.Println("read config ")
	//fmt.Println(string(buf))
	str := strings.Split(string(buf), "\n")
	fmt.Println(len(str))
	fmt.Println(str)
	for i := 0; i < len(str)-1; i++ {
		fmt.Printf("str = %s\n", str[i])
		swinfo[i].ipAddr = strings.Split(str[i], ":")[0]
		swinfo[i].userIonf = strings.Split(str[i], ":")[1]
		swinfo[i].passWd = strings.Split(str[i], ":")[2]
		swinfo[i].backFileName = strings.Split(str[i], ":")[0] + ".cfg"

	}

	return swinfo
}

//

func backupCfg(username, password, host, backfile string) {

	conn, err := ftp.Dial(host + ":21")
	if err != nil {
		log.Printf("", err)
	}
	defer conn.Quit()

	// 登录
	err = conn.Login(username, password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("成功连sys/接并登录 FTP 服务器")

	// 下载远程文件
	resp, err := conn.Retr("/startup.cfg")
	if err != nil {
		log.Printf("", err)
	}
	defer resp.Close()

	// 创建本地文件
	outFile, err := os.Create(backfile)
	if err != nil {
		log.Printf("", err)
	}
	defer outFile.Close()

	// 写入本地
	_, err = io.Copy(outFile, resp)
	if err != nil {
		log.Printf("", err)
	}

	fmt.Println("文件下载成功")

}

func main() {

	minfo := readSWadd("/home/zhangyq/Documents/work/config/config")
	fmt.Println(minfo)
	// 连接到 FTP 服务器

	filePath := "/home/zhangyq/" + time.Now().Format("20060102") + "/"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 目录不存在，创建它（包括父目录）
		err := os.MkdirAll(filePath, 0755)
		if err != nil {
			log.Fatalf("创建目录失败: %v", err)
		}
		log.Printf("目录已创建: %s", filePath)
	}

	for k, v := range minfo {

		log.Printf("正在备份第 %d 个交换机,地址是 %s ", k+1, v.ipAddr)
		file := filePath + v.backFileName
		backupCfg(v.userIonf, v.passWd, v.ipAddr, file)

	}

}
