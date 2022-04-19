package commands

import (
	"errors"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/auth"
)

var LoginCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "login in idp",
		ShortDescription: `
Login in Idp server, get an avaibale Token
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("username", true, false, "imput username ").EnableStdin(),
		cmds.StringArg("password", true, false, "imput password").EnableStdin(),
		cmds.StringArg("appid", true, false, "imput appid").EnableStdin(),
	},
	Options: []cmds.Option{},
	PreRun: func(req *cmds.Request, env cmds.Environment) error {
		return nil
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		if err := req.ParseBodyArgs(); err != nil {
			return err
		}
		log.Infof("test %+v", req.Arguments)
		if len(req.Arguments) != 3 {
			return errors.New("bad argument")
		}
		if err := auth.Login(req.Arguments[0], req.Arguments[1], req.Arguments[2]); err != nil {
			return err
		}
		res.Emit("login in success")
		return nil
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			return nil
		},
	},
}
