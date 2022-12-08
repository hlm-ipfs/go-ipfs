package corehttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ipfs/kubo/core"
	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
)

var TusUploadPath string = "/data/ipfs/tus/uploads"

func CancelUpload(path string) ServeOption {
	return func(i *core.IpfsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			returnMap := make(map[string]interface{})
			if r.Method != http.MethodPost {
				returnMap["code"] = "500"
				returnMap["message"] = "only POST allowed"
				reByte, _ := json.Marshal(returnMap)
				http.Error(w, string(reByte), http.StatusMethodNotAllowed)
				return
			}
			if err := r.ParseForm(); err != nil {
				returnMap["code"] = "500"
				returnMap["message"] = err.Error()
				reByte, _ := json.Marshal(returnMap)
				http.Error(w, string(reByte), http.StatusMethodNotAllowed)
				return
			}
			s, _ := ioutil.ReadAll(r.Body) //把	body 内容读入字符串 s
			//
			type AddIpfsReq struct {
				UUID string `json:"uuid"`
			}
			var addIpfsReq AddIpfsReq

			json.Unmarshal(s, &addIpfsReq)
			filePath := TusUploadPath + "/" + addIpfsReq.UUID
			//删除文件
			os.Remove(filePath)
			os.Remove(filePath + ".info")
			returnMap["code"] = "200"
			returnMap["message"] = "success."
			reByte, _ := json.Marshal(returnMap)
			io.WriteString(w, string(reByte))

		})
		return mux, nil
	}
}

func AddIpfs(path string) ServeOption {
	return func(i *core.IpfsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			returnMap := make(map[string]interface{})
			if r.Method != http.MethodPost {
				returnMap["code"] = "500"
				returnMap["message"] = "only POST allowed"
				reByte, _ := json.Marshal(returnMap)
				http.Error(w, string(reByte), http.StatusMethodNotAllowed)
				return
			}
			if err := r.ParseForm(); err != nil {
				returnMap["code"] = "500"
				returnMap["message"] = err.Error()
				reByte, _ := json.Marshal(returnMap)
				http.Error(w, string(reByte), http.StatusMethodNotAllowed)
				return
			}
			s, _ := ioutil.ReadAll(r.Body) //把	body 内容读入字符串 s
			//
			type AddIpfsReq struct {
				UUID     string `json:"uuid"`
				Encrypt  string `json:"encrypt"`
				Password string `json:"passwd"`
			}
			var addIpfsReq AddIpfsReq

			json.Unmarshal(s, &addIpfsReq)
			filePath := TusUploadPath + "/" + addIpfsReq.UUID

			exist, err := PathExists(filePath)
			if err != nil {
				returnMap["code"] = "500"
				returnMap["message"] = err.Error()
				reByte, _ := json.Marshal(returnMap)
				http.Error(w, string(reByte), http.StatusMethodNotAllowed)
				return
			}
			if !exist {
				returnMap["code"] = "500"
				returnMap["message"] = "文件不存在."
				reByte, _ := json.Marshal(returnMap)
				http.Error(w, string(reByte), http.StatusMethodNotAllowed)
				return
			}

			//读取文件
			//input, err := ioutil.ReadFile(filePath + ".info")
			//if err != nil {
			//	http.Error(w, err.Error(), http.StatusBadRequest)
			//	return
			//}
			//var fileInfo tusd.FileInfo
			//err = json.Unmarshal(input, &fileInfo)
			//if err != nil {
			//	http.Error(w, "格式化文件json报错", http.StatusBadRequest)
			//	return
			//}
			filePath = TusUploadPath + "/" + addIpfsReq.UUID
			if len(addIpfsReq.Encrypt) == 0 {
				addIpfsReq.Encrypt = "false"
			}
			url := "http://127.0.0.1:5001/api/v0/add?stream-channels=true&pin=false&wrap-with-directory=false&progress=false&encrypt=" + addIpfsReq.Encrypt
			if len(addIpfsReq.Password) > 0 {
				url = url + "&passwd=" + addIpfsReq.Password
			}
			//addIpfs
			req, err := NewfileUploadRequest(url, nil, "file", filePath)
			if err != nil {
				returnMap["code"] = "500"
				returnMap["message"] = err.Error()
				reByte, _ := json.Marshal(returnMap)
				http.Error(w, string(reByte), http.StatusMethodNotAllowed)
				return
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				returnMap["code"] = "500"
				returnMap["message"] = err.Error()
				reByte, _ := json.Marshal(returnMap)
				http.Error(w, string(reByte), http.StatusMethodNotAllowed)
				return
			}
			returnMap["code"] = "200"
			returnMap["message"] = "success"
			type AddResp struct {
				Name     string `json:"Name"`
				Hash     string `json:"Hash"`
				Size     string `json:"Size"`
				Password string `json:"Password"`
			}
			var dataResp AddResp
			json.Unmarshal(body, &dataResp)
			returnMap["data"] = dataResp

			reByte, _ := json.Marshal(returnMap)
			io.WriteString(w, string(reByte))
			//删除文件
			os.Remove(filePath)
			os.Remove(filePath + ".info")

		})
		return mux, nil
	}
}

func TusFiles(path string) ServeOption {
	return func(node *core.IpfsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		storePath := TusUploadPath
		//if ex, err := os.Executable(); err == nil {
		//	storePath = filepath.Dir(ex) + "/uploads"
		//}
		//fmt.Println("===storePath=====", storePath)
		b, err := PathExists(storePath)
		if !b {
			os.Mkdir(storePath, 0777)
			// 再修改权限
			os.Chmod(storePath, 0777)
		}

		store := filestore.FileStore{
			Path: storePath,
		}

		composer := tusd.NewStoreComposer()
		store.UseIn(composer)
		handler, err := tusd.NewHandler(tusd.Config{
			BasePath:              path,
			StoreComposer:         composer,
			NotifyCompleteUploads: true,
		})
		if err != nil {
			fmt.Println("========", err)
		}
		go func() {
			for {
				event := <-handler.CompleteUploads
				fmt.Printf("Upload %s finished\n", event.Upload.ID)
			}
		}()

		//http.Handle(path,handler)
		mux.Handle(path, http.StripPrefix(path, handler))
		return mux, nil
	}
}

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}

//
func NewfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormField(path)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", uri, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request, err
}
