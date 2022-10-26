package auth

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/kubo/core"
	"github.com/o1egl/paseto"
	"net/http"
	"os"
	"strings"
)

var (
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
)

const PublicPem = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCkVCtfR1yZ4RmqF4fV+RD8EizX
a811s5JSHQDD0pyq5FNynPafkgwIUgK6ve4jwGH1IYfjdo71YVwBZNGRrqBpfOLP
3jV3jtNf02uyySqskes2cI0xFog04XK6DyMm5EGTMbKIh1C5xcpsOi21nGSHhFTC
RjdZUJd9Iel+BoXkVwIDAQAB
-----END PUBLIC KEY-----
`
const PrivatePem = `
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCkVCtfR1yZ4RmqF4fV+RD8EizXa811s5JSHQDD0pyq5FNynPaf
kgwIUgK6ve4jwGH1IYfjdo71YVwBZNGRrqBpfOLP3jV3jtNf02uyySqskes2cI0x
Fog04XK6DyMm5EGTMbKIh1C5xcpsOi21nGSHhFTCRjdZUJd9Iel+BoXkVwIDAQAB
AoGAcOqpVvIpTk+gHAHJRB2+LweqKmiYKN24mJX3VZfeMYttT99NlD596CW6XGlw
Pr7OUOu2fXWVLEW3O/n0C1/sNxXm3f5Sk7u4tFdlNOVl4jGohZshdfGUDQTdZvkX
VSXTNyqwcVntYij4KxJ1cDN5un2OCmkTg7QTecCXinK0YHkCQQDCnDcVeXeVvJT6
gIWDwFIhXfx0FnNErcqnJ6NYC1uCkuXTKqV9JPzUT0bqtiaR5IGhV8sBBMzbGEgf
ur3w4DtDAkEA2CqSXXAOsdnv5OM3cYBvX5cPQByTYhGiB4bAM8wClcILjG50BU5l
tZFiVoe9SWVvjXm+5mNb4qdksQr+QzvfXQJAOIknW365Je/uG8QAhaD2eykRsSvg
esA+UkcqpeZOxGUmjS0jfHKlXIxXXtJh7BpFKnkVyYr6dF/efBFFItgRfwJAfShA
8FqmF3GAxTyNaGkdkvIKgWtKJjjhYIwzkonnYUAloZx+Dmhy2SqhDkqvXOSfLwGz
dv9GPzf+Q8eCqfjrfQJBALUlQpp8ovhyaTZTuyA6vDGsn5w6qVmca2D52p69w3do
OyYy5hTSOJSKyakEFegeLTTqMHx0xr1m/VKOeNfQgWY=
-----END RSA PRIVATE KEY-----
`

func GenRsaKeyFile(prefix string) error {
	size := 1024
	privateKey, err := rsa.GenerateKey(rand.Reader, size)
	if err != nil {
		return err

	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}

	file, err := os.Create(prefix + "_private.pem")
	if err != nil {
		return err
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	publicKey := &privateKey.PublicKey

	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	file, err = os.Create(prefix + "_public.pem")
	if err != nil {
		return err
	}

	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}
func DecodePublicKey(pemContent []byte) (publicKey *rsa.PublicKey, err error) {
	block, _ := pem.Decode(pemContent)
	if block == nil {
		return nil, fmt.Errorf("pem.Decode(%s)：pemContent decode error", pemContent)
	}
	switch block.Type {
	case "CERTIFICATE":
		pubKeyCert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("x509.ParseCertificate(%s)：%w", pemContent, err)
		}
		pubKey, ok := pubKeyCert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("公钥证书提取公钥出错 [%s]", pemContent)
		}
		publicKey = pubKey
	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("x509.ParsePKIXPublicKey(%s),err:%w", pemContent, err)
		}
		pubKey, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("公钥解析出错 [%s]", pemContent)
		}
		publicKey = pubKey
	case "RSA PUBLIC KEY":
		pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("x509.ParsePKCS1PublicKey(%s)：%w", pemContent, err)
		}
		publicKey = pubKey
	}
	return publicKey, nil
}

func DecodePrivateKey(pemContent []byte) (privateKey *rsa.PrivateKey, err error) {
	block, _ := pem.Decode(pemContent)
	if block == nil {
		return nil, fmt.Errorf("pem.Decode(%s)：pemContent decode error", pemContent)
	}
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		pk8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("私钥解析出错 [%s]", pemContent)
		}
		var ok bool
		privateKey, ok = pk8.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("私钥解析出错 [%s]", pemContent)
		}
	}
	return privateKey, nil
}

// Authorization 前端收到错误中有token has expired需要向服务端重新申请
func Authorization(n *core.IpfsNode, r *http.Request) error {
	urlCode:=r.URL.Query().Get("x-device-code")
	headCode:=r.Header.Get("x-device-code")
	var deviceCode string
	if len(urlCode)==0&&len(headCode)==0{
		return nil
	}else {
		if len(urlCode)>0{
			deviceCode=urlCode
		}else if len(headCode)>0 {
			deviceCode=headCode
		}
	}
	if strings.Contains(r.RemoteAddr,"127.0.0.1") {
		return nil
	}
	encryptedBytes, err := hex.DecodeString(deviceCode)
	if err != nil {
		return err
	}
	PrivateKey, err := DecodePrivateKey([]byte(PrivatePem))
	if err != nil {
		return err
	}
	decryptedBytes, err := PrivateKey.Decrypt(nil, encryptedBytes, &rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		return err
	}
	code := strings.Split(string(decryptedBytes), ":")
	if len(code) != 2 {
		return fmt.Errorf("%s", "invalid authorization token format")
	}
	client_id := code[1]
	key := datastore.NewKey(fmt.Sprintf("/remote/client/keys/%s", client_id))
	valBytes, err := n.Repo.Datastore().Get(r.Context(), key)
	if err != nil {
		return err
	}
	val := ClientDevice{}
	err = json.Unmarshal(valBytes, &val)
	if err != nil {
		return err
	}
	token := val.Token
	public_key := val.PublicKey

	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return err
	}
	publicKeyBytes, err := hex.DecodeString(public_key)
	if err != nil {
		return err
	}
	publicKey := ed25519.PublicKey(publicKeyBytes)
	var newJsonToken paseto.JSONToken
	var newFooter string
	err = paseto.NewV2().Verify(string(tokenBytes), publicKey, &newJsonToken, &newFooter)
	if err != nil {
		return err
	}
	return nil
}

type ClientDevice struct {
	ClientID   string `json:"client_id"`   //主键，服务端根据规则生成
	Token      string `json:"token"`       //设备用户信息等签发的一个token
	PublicKey  string `json:"public_key"`  //服务端token加密的公钥
	Code       string `json:"code"`        //小盒子生成的MD5摘要，返回给客户端
	CreateTime string `json:"create_time"` //创建时间，或者更新时间
}
