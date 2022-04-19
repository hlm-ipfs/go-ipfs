package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	logging "github.com/ipfs/go-log"
	"hlm-ipfs/ipfs-probe/host"
	"io/ioutil"
	"net/http"
)

var log = logging.Logger("auth")

//检索文件(input: cid)
//rpc Create (basic.String) returns (RetrievalCreateResponse)
//验证检索token
//rpc Verify (basic.String) returns (basic.Empty)
func CreateRetrievalOrder(cid string, headers http.Header, order *ResponseCreateRetrievalOrder) error {
	client := &http.Client{}
	reqBody := RequestCreateRetrievalOrder{
		Value: cid,
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "http://103.44.247.16:31686/market/Retrieval/Create", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.Error(err)
		return err
	}
	req.Header.Set("Authorization", headers.Get("Authorization"))
	req.Header.Set("Content-Type", "application/json")
	log.Infof("headers: %+v", req.Header)
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return err
	}
	type Response struct {
		Status  int32  `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Value   string `json:"value"`
	}
	respBody := Response{}
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Infof("response: %+v", respBody)
	if respBody.Status >= 400 {
		log.Warnf("bad response %+v", respBody)
	}
	err = json.Unmarshal(bytes.NewBufferString(respBody.Value).Bytes(), order)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

type RequestCreateRetrievalOrder struct {
	Value string //cid
}

type ResponseCreateRetrievalOrder struct {
	OrderNo string
	Token   string
}

func VerifyRetrievalToken(token string) error {
	client := &http.Client{}
	reqBody := RequestVerifyRetrievalToken{
		Value: token,
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "http://103.44.247.16:31686/market/Retrieval/Verify", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.Error(err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return err
	}
	type Response struct {
		Status  int32  `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Value   string `json:"value"`
	}
	respBody := Response{}
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		log.Error(err)
		return err
	}
	if respBody.Status != 200 {
		log.Warnf("bad response %+v", respBody)
		return errors.New(respBody.Message)
	}
	return nil
}

type RequestVerifyRetrievalToken struct {
	Value string //token
}

type DeviceInfo struct {
	IMEI       string `protobuf:"bytes,1,opt,name=IMEI,proto3" json:"IMEI,omitempty"`             //设备唯一标识
	Brand      string `protobuf:"bytes,2,opt,name=Brand,proto3" json:"Brand,omitempty"`           //品牌
	Model      string `protobuf:"bytes,3,opt,name=Model,proto3" json:"Model,omitempty"`           //型号
	OSName     string `protobuf:"bytes,4,opt,name=OSName,proto3" json:"OSName,omitempty"`         //操作系统:android/ios等
	OSVersion  string `protobuf:"bytes,5,opt,name=OSVersion,proto3" json:"OSVersion,omitempty"`   //操作系统版本
	AppName    string `protobuf:"bytes,6,opt,name=AppName,proto3" json:"AppName,omitempty"`       //当前使用的应用的名称
	AppVersion string `protobuf:"bytes,7,opt,name=AppVersion,proto3" json:"AppVersion,omitempty"` //当前使用的应用的版本
}
type LoginByUserNameRequest struct {
	UserName   string `protobuf:"bytes,1,opt,name=UserName,proto3" json:"UserName,omitempty"`
	Password   string `protobuf:"bytes,2,opt,name=Password,proto3" json:"Password,omitempty"`
	VerifyCode string `protobuf:"bytes,3,opt,name=VerifyCode,proto3" json:"VerifyCode,omitempty"` //图片验证码
}
type LoginRequest struct {
	AppID        string                  `protobuf:"bytes,1,opt,name=AppID,proto3" json:"AppID,omitempty"`               //所属的app
	Device       *DeviceInfo             `protobuf:"bytes,2,opt,name=Device,proto3" json:"Device,omitempty"`             //设备信息
	UserPassword *LoginByUserNameRequest `protobuf:"bytes,3,opt,name=UserPassword,proto3" json:"UserPassword,omitempty"` //用户名+密码登录
}
type LoginResponse struct {
	MFAToken     string `protobuf:"bytes,1,opt,name=MFAToken,proto3" json:"MFAToken,omitempty"`         //不为空时代表需要两步认证(MFALogin时需要的参数)
	UserID       string `protobuf:"varint,2,opt,name=UserID,proto3" json:"UserID,omitempty"`            //用户id（用于前端(移动端)缓存）
	AccessToken  string `protobuf:"bytes,3,opt,name=AccessToken,proto3" json:"AccessToken,omitempty"`   //登录成功后返回的jwt token
	RefreshToken string `protobuf:"bytes,4,opt,name=RefreshToken,proto3" json:"RefreshToken,omitempty"` //保留
	Newcomer     bool   `protobuf:"varint,10,opt,name=Newcomer,proto3" json:"Newcomer,omitempty"`       //是否新注册的用户
}

//http://103.44.247.16:18001/api/idp/idp/login
/*
{
    "AppID": "174af3b0f840f962abc6792921dc17b2",
    "UserPassword": {
        "UserName": "xiaoming",
        "Password": "zdz1234561"
    },
    "Device": {
        "IMEI": "huangdong"
    }
}
*/

var DefaultAuthorization string

func Login(username string, password string, appid string) error {
	pid, _ := host.ReadPlatformMachineID()
	client := &http.Client{}
	reqBody := &LoginRequest{
		AppID: appid,
		Device: &DeviceInfo{
			IMEI: pid,
		},
		UserPassword: &LoginByUserNameRequest{
			UserName: username,
			Password: password,
		},
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "http://103.44.247.16:18001/api/idp/idp/login", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.Error(err)
		return err
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return err
	}
	type Response struct {
		Status  int32  `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Value   string `json:"value"`
	}
	respBody := Response{}
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Infof("response: %+v", respBody)
	if respBody.Status >= 400 {
		log.Warnf("bad response %+v", respBody)
	}
	loginRespons := LoginResponse{}
	err = json.Unmarshal(bytes.NewBufferString(respBody.Value).Bytes(), &loginRespons)
	if err != nil {
		log.Error(err)
		return err
	}
	DefaultAuthorization = fmt.Sprintf("Idp %v", loginRespons.AccessToken)
	return nil
}
