package token

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var publicKey *PublicKey

func init() {
	publicKey = &PublicKey{publickey: nil, flag: false}
}

type PublicKey struct {
	publickey *rsa.PublicKey
	flag      bool
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func (p *PublicKey) readFromPubFile() {
	str1 := getCurrentDirectory()
	//str2 := getParentDirectory(str1)
	keyData, _ := ioutil.ReadFile(str1 + "/conf/publickey.pub")
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		fmt.Println("parse public key error!")
		return
	}
	p.publickey = key
	p.flag = true
}

func (p *PublicKey) showPubKey() *rsa.PublicKey {
	if p.flag == false {
		fmt.Println("public key will be read from file")
		p.readFromPubFile()
	}
	return p.publickey
}

func CheckToken(token string) error {
	parts := strings.Split(token, ".")
	//tmp := ""
	//if strings.Contains(parts[0]," ") {
	//	p := strings.Split(parts[0]," ")
	//	tmp = p[1]
	//}else {
	//	tmp = parts[0]
	//}
	np := strings.Join(parts[0:2], ".")
	//fmt.Printf("np is %s\n",np)
	err := jwt.SigningMethodRS256.Verify(np, parts[2], publicKey.showPubKey())
	return err
}

func GetUserFromToken(token string) (string, float64, error) {
	parts := strings.Split(token, ".")
	a, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return "", 0.0, err
	}
	tempStru := map[string]interface{}{}
	if err := json.Unmarshal(a, &tempStru); err != nil {
		return "", 0.0, err
	}
	username := tempStru["user_name"].(string)
	exptime := tempStru["exp"].(float64)
	return username, exptime, nil
}
