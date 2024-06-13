package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/lewisy158/youdao-translate-server/logging"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	gClient = g.Client()
	keyHash []byte
	ivHash  []byte
)

const (
	dataString = "i=%s&from=AUTO&to=AUTO&domain=0&dictResult=true&keyid=webfanyi&sign=%s&client=%s&product=%s&appVersion=%s&vendor=%s&pointParam=%s&mysticTime=%d&keyfrom=%s"
)

func generateData(translateWords string) string {
	client := "fanyideskweb"
	product := "webfanyi"
	key := "fsdsogkndfokasodnaso"
	pointParam := "client,mysticTime,product"
	appVersion := "1.0.0"
	vendor := "web"
	keyfrom := "fanyi.web"
	mysticTime := time.Now().Unix() * 1000

	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("client=%s&mysticTime=%d&product=%s&key=%s", client, mysticTime, product, key)))
	sign := fmt.Sprintf("%x", hash.Sum(nil))

	return fmt.Sprintf(dataString,
		translateWords, sign, client, product, appVersion, vendor, pointParam, mysticTime, keyfrom,
	)
}

func PKCS7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, fmt.Errorf("unpadding size is larger than data size")
	}
	return data[:(length - unpadding)], nil
}

func Decrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
	}
	plaintext := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, data)
	plaintext, err = PKCS7UnPadding(plaintext)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func hashCal(str string) []byte {
	// 创建 MD5 哈希
	hash := md5.New()
	hash.Write([]byte(str)) // 编码为 UTF-8 并写入哈希对象

	// 计算哈希值并获取结果
	return hash.Sum(nil)
}

type TranslateWords struct {
	Text string `form:"text"`
}

func youdaoTranslate(context *gin.Context) {
	var translateWords TranslateWords
	if err := context.ShouldBind(&translateWords); err != nil {
		// 处理错误请求
		return
	}
	logging.Infof("需要翻译文本: %s", translateWords.Text)

	data := generateData(translateWords.Text)

	response, err := gClient.Post(context, "https://dict.youdao.com/webtranslate", data)
	if err != nil {
		logging.Errorf("youdaoTranslate error: %v", err)
		return
	}
	responseString := response.ReadAllString()

	decodedBytes, err := base64.URLEncoding.DecodeString(responseString)
	if err != nil {
		logging.Errorf("youdaoTranslate error: %v", err)
		return
	}

	decrypted, err := Decrypt(decodedBytes, keyHash, ivHash)
	if err != nil {
		logging.Errorf("youdaoTranslate error: %v", err)
		return
	}
	logging.Infof("翻译结果: %s", string(decrypted))

	context.String(http.StatusOK, string(decrypted))
}

func init() {
	// 获取当前程序的绝对路径
	executablePath, _ := os.Executable()
	executableDir := filepath.Dir(executablePath)

	// 初始化日志
	logging.Init(executableDir, "run.log")

	gClient.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54")
	gClient.SetHeader("Cookie", "OUTFOX_SEARCH_USER_ID_NCOO=976405377.6815147; OUTFOX_SEARCH_USER_ID=-198948307@211.83.126.235; _ga=GA1.2.1162596953.1667349221; search-popup-show=12-2")
	gClient.SetHeader("Referer", "https://fanyi.youdao.com/")

	keyHash = hashCal("ydsecret://query/key/B*RGygVywfNBwpmBaZg*WT7SIOUP2T0C9WHMZN39j^DAdaZhAnxvGcCY6VYFwnHl")
	ivHash = hashCal("ydsecret://query/iv/C@lZe2YzHtZ2CYgaXKSVfsb7Y4QWHjITPPZ0nQp87fBeJ!Iv6v^6fvi2WN@bYpJ4")
}

func main() {
	r := gin.Default()
	r.POST("/youdaoTranslate", youdaoTranslate)
	err := r.Run(":9527")
	if err != nil {
		logging.Panicf("无法启动服务, 原因: %v\n", err)
	}
}
