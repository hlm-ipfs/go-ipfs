package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	AK = "017194e9718f07feefc4b03422d8be5df654bafc623251480f7d760d1209b4ca39"
	SK = "02595d553697305c7670dfd92628e5ff68080335265edf804aea4e6e8df5112464"
)

func getSecKec(ak string) string {
	if strings.Compare(AK, ak) == 0 {
		return SK
	}
	return ""
}
func AKSKAuth(w http.ResponseWriter, r *http.Request) error {
	//os.Setenv("AK_SK_AUTH","false")
	if  strings.Contains(os.Getenv("AK_SK_AUTH"),"false") {
		return nil
	}
	var (
		ak, sk, sign, timeStamp, serverSign string
		iTime, timeDiff                     int64
		err                                 error
		requestBody                         []byte
	)
	ak = r.Header.Get("AccessKey")
	sign = r.Header.Get("Signature")
	timeStamp = r.Header.Get("TimeStamp")
	if ak == "" || sign == "" || timeStamp == "" {
		return errors.New("header missed: AccessKey|Signature|TimeStamp")
	}
	log.Infof("client:  AccessKey: %+v", ak)
	log.Infof("client:  Signature: %+v", sign)
	log.Infof("client:  TimeStamp: %+v", timeStamp)
	//check time
	iTime, err = strconv.ParseInt(timeStamp, 10, 64)
	if err != nil {
		return errors.New(fmt.Sprintf(`TimeStamp Error %s`, err.Error()))
	}
	timeDiff = time.Now().Unix() - iTime
	if timeDiff >= 60 || timeDiff <= -60 {
		return errors.New("timestamp error")
	}
	//check signature
	sk = getSecKec(ak)
	if sk == "" {
		return errors.New("User not exist")
	}
	requestBody, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

	serverSign = generateSign(r.Method, formatURLPath(r.URL.Path), r.URL.RawQuery, ak, timeStamp, sk, requestBody)
	log.Infof("server Signature: %+v ", serverSign)
	if serverSign != sign {
		return errors.New("signature error")
	}
	return nil
}
func formatURLPath(in string) string {
	in = strings.TrimSpace(in)
	if strings.HasSuffix(in, "/") {
		return in[:len(in)-1]
	}
	return in
}

func sha256byteArr(in []byte) string {
	if in == nil || len(in) == 0 {
		return ""
	}
	h := sha256.New()
	h.Write(in)
	return hex.EncodeToString(h.Sum(nil))
}

func generateSign(method, url, query, ak, timestamp, sk string, requestBody []byte) string {
	return hmacSha256(fmt.Sprintf(`%s\n%s\n%s\n%s\n%s\n%s`, method, url, query, ak, timestamp, sha256byteArr(requestBody)), sk)
}

func hmacSha256(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
