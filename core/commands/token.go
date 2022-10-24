package commands

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-datastore"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/kubo/auth"
	"github.com/ipfs/kubo/core/commands/cmdenv"
	"github.com/ipfs/kubo/repo/fsrepo"
	"github.com/o1egl/paseto"
	"io"
	"time"
)

const (
	ErrBadArguments int = 10000+iota
	ErrTokenExists
	ErrTokenNotExists
	ErrTokenNotVerify
)
// TokenCmd  小盒子访问令牌管理功能
var TokenCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Manager ipfs token.",
		ShortDescription: `'ipfs token' is a set of commands to help manager client token
for your IPFS node.
`,
		LongDescription: `'ipfs token' is a set of commands to help manager client token
for your IPFS node.
`,
	},

	Subcommands: map[string]*cmds.Command{
		"create":  createTokenCmd,
		"refresh": refreshTokenCmd,
		"revoke":  revokeTokenCmd,
	},
}

var createTokenCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "create ipfs token.",
		ShortDescription: `
		ipfs token create
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("client_id", true, false, "server side client id"),
		cmds.StringArg("public_key", true, false, "server side token public key"),
		cmds.StringArg("token", true, false, "server side token "),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		//参数解析, 持久化，返回
		if len(req.Arguments) < 3 {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: "expecting three arguments: client_id, public_key and token",
				Code:    ErrBadArguments,
			})
		}
		client_id := req.Arguments[0]
		public_key := req.Arguments[1]
		token := req.Arguments[2]

		var newJsonToken paseto.JSONToken
		var newFooter string
		if client_id == "" {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: "illegal parameter client id need non empty string",
				Code:    ErrBadArguments,
			})
		}
		tokenBytes, err := hex.DecodeString(token)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		publicKeyBytes, err := hex.DecodeString(public_key)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		publicKey := ed25519.PublicKey(publicKeyBytes)
		err = paseto.NewV2().Verify(string(tokenBytes), publicKey, &newJsonToken, &newFooter)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrTokenNotVerify,
			})
		}
		//返回给客户端的code码
		code := fmt.Sprintf("%s:%s", GetRandomString(6), client_id)
		key := datastore.NewKey(fmt.Sprintf("/remote/client/keys/%s", client_id))

		val := auth.ClientDevice{
			ClientID:   client_id,
			Token:      token,
			PublicKey:  public_key,
			CreateTime: time.Now().String(),
			Code:       code,
		}

		cfgRoot, err := cmdenv.GetConfigRoot(env)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		r, err := fsrepo.Open(cfgRoot)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		defer r.Close()

		if err := req.ParseBodyArgs(); err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		valBytes, err := json.Marshal(val)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		exists, err := r.Datastore().Has(req.Context, key)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		if exists {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: "token exists,please refresh",
				Code:    ErrTokenExists,
			})
		}
		err = r.Datastore().Put(req.Context, key, valBytes)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		//code 进行非对称加密
		PublicKey, err := auth.DecodePublicKey([]byte(auth.PublicPem))
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		//GenRsaKeyFile("box")
		encryptedBytes, err := rsa.EncryptOAEP(
			sha256.New(),
			rand.Reader,
			PublicKey,
			[]byte(code),
			nil)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		return cmds.EmitOnce(res, &TokenOutput{
			Data: hex.EncodeToString(encryptedBytes),
		})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *IdOutput) error {
			return nil
		}),
	},
	Type: TokenOutput{},
}

var refreshTokenCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "refresh ipfs token.",
		ShortDescription: `
		ipfs token refresh
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("client_id", true, false, "server side client id"),
		cmds.StringArg("public_key", true, false, "server side token public key"),
		cmds.StringArg("token", true, false, "server side token "),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		//参数解析, 持久化，返回
		if len(req.Arguments) < 3 {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message:"expecting three arguments: client_id, public_key and token",
				Code:    ErrBadArguments,
			})
		}
		client_id := req.Arguments[0]
		public_key := req.Arguments[1]
		token := req.Arguments[2]

		var newJsonToken paseto.JSONToken
		var newFooter string
		if client_id == "" {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: "illegal parameter client id need non empty string",
				Code:    ErrBadArguments,
			})
		}
		tokenBytes, err := hex.DecodeString(token)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		publicKeyBytes, err := hex.DecodeString(public_key)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		publicKey := ed25519.PublicKey(publicKeyBytes)
		err = paseto.NewV2().Verify(string(tokenBytes), publicKey, &newJsonToken, &newFooter)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrTokenNotVerify,
			})
		}
		//返回给客户端的code码
		code := fmt.Sprintf("%s:%s", GetRandomString(6), client_id)
		key := datastore.NewKey(fmt.Sprintf("/remote/client/keys/%s", client_id))

		val := auth.ClientDevice{
			ClientID:   client_id,
			Token:      token,
			PublicKey:  public_key,
			CreateTime: time.Now().String(),
			Code:       code,
		}

		cfgRoot, err := cmdenv.GetConfigRoot(env)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		r, err := fsrepo.Open(cfgRoot)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		defer r.Close()

		if err := req.ParseBodyArgs(); err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		valBytes, err := json.Marshal(val)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		exists, err := r.Datastore().Has(req.Context, key)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		if !exists {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: "token not exists,please create",
				Code:    ErrTokenNotExists,
			})
		}
		err = r.Datastore().Put(req.Context, key, valBytes)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		//code 进行非对称加密
		PublicKey, err := auth.DecodePublicKey([]byte(auth.PublicPem))
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		//GenRsaKeyFile("box")
		encryptedBytes, err := rsa.EncryptOAEP(
			sha256.New(),
			rand.Reader,
			PublicKey,
			[]byte(code),
			nil)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		return cmds.EmitOnce(res, &TokenOutput{
			Data: hex.EncodeToString(encryptedBytes),
		})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *IdOutput) error {
			return nil
		}),
	},
	Type: TokenOutput{},
}

var revokeTokenCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "revoke ipfs token.",
		ShortDescription: `
		ipfs token revoke
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("client_id", true, false, "server side client id"),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		//参数解析, 持久化，返回
		if len(req.Arguments) < 1 {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: "expecting one arguments: client_id",
				Code:    ErrBadArguments,
			})
		}
		client_id := req.Arguments[0]

		//返回给客户端的code码
		key := datastore.NewKey(fmt.Sprintf("/remote/client/keys/%s", client_id))

		cfgRoot, err := cmdenv.GetConfigRoot(env)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		r, err := fsrepo.Open(cfgRoot)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		defer r.Close()

		if err := req.ParseBodyArgs(); err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		err = r.Datastore().Delete(req.Context, key)
		if err != nil {
			return cmds.EmitOnce(res, &TokenOutput{
				Data:    "",
				Message: err.Error(),
				Code:    ErrBadArguments,
			})
		}
		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *IdOutput) error {
			return nil
		}),
	},
	Type: TokenOutput{},
}

// GetRandomString 生成一个随机salt
func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}

// TokenOutput 根据状态码 400就是错误，200就是正常返回
type TokenOutput struct {
	Data    string `json:"data"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
