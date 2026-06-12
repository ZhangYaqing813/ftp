package main

import (
	"encoding/base64"
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
	//fmt.Println("read config ")
	//fmt.Println(string(buf))
	// 处理配置文件行

	lines := strings.Split(string(buf), "\n")
	//log.Printf("config file is : %s \n", lines)
	for _, line := range lines {

		line = strings.TrimSpace(line)
		log.Printf("line === %s \n", line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			log.Printf("跳过格式错误的行: %s", line)
			continue
		}
		swinfo = append(swinfo, SWinfo{
			ipAddr:       parts[0],
			userIonf:     parts[1],
			passWd:       parts[2],
			backFileName: parts[0] + ".cfg",
		})
	}

	return swinfo
}

// base 64 编码
func base64_t(password string) string {
	decode, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		log.Fatalf("解密失败: %s \n", err)
	}

	return string(decode)
}

func backupCfg(username, password, host, backfile string) {

	conn, err := ftp.Dial(host + ":21")
	if err != nil {
		log.Printf("", err)
	}
	defer conn.Quit()

	// 登录
	//fmt.Println("password == ", password)
	err = conn.Login(username, password)
	if err != nil {
		log.Printf("login failed ： %s \n", err)
	}

	//fmt.Println("成功连sys/接并登录 FTP 服务器")

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

	//fmt.Println("文件下载成功")

}

func main() {

	minfo := readSWadd("/data/backup/config")
	log.Printf("共读取到 %d 台交换机配置", len(minfo))
	//fmt.Println(minfo)

	// 创建日志文件
	logfile := "log/" + time.Now().Format("20060102") + ".log"

	log_f, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("日志文件创建失败：%v", err)
	}
	defer log_f.Close()
	log.SetOutput(log_f)

	// 连接到 FTP 服务器

	filePath := "/data/backup/" + time.Now().Format("20060102") + "/"

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 目录不存在，创建它（包括父目录）
		err := os.MkdirAll(filePath, 0755)
		if err != nil {
			log.Fatalf("创建目录失败: %v", err)
		}
		log.Printf("目录已创建: %s", filePath)
	}
	successCount := 0
	for k, v := range minfo {

		log.Printf("正在备份第 %d 个交换机,地址是 %s ", k+1, v.ipAddr)

		file := filePath + v.backFileName
		backupCfg(v.userIonf, base64_t(v.passWd), v.ipAddr, file)
		successCount++

	}
	log.Printf("========== 备份完成，成功 %d/%d 台 ==========", successCount, len(minfo))

}
