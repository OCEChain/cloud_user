package router

import (
	"github.com/henrylee2cn/faygo"
	"user/handler"
)

// Route register router in a tree style.
func Route(frame *faygo.Framework) {
	frame.Route(
		frame.NewNamedAPI("注册接口", "POST", "/register", &handler.Register{}),
		frame.NewNamedAPI("发送短信验证码", "POST", "/sendsms", &handler.Send_SMS{}),
		frame.NewNamedAPI("登陆", "POST", "/login", &handler.Login{}),
		frame.NewNamedAPI("修改头像", "POST", "/edit_face", &handler.FaceEdit{}),
		frame.NewNamedAPI("token获取用户信息", "POST", "/getinfo", &handler.GetUserInfo{}),
		frame.NewNamedAPI("上传身份证信息", "POST", "/certification", &handler.Certification{}),
		frame.NewNamedAPI("修改密码发送短信", "POST", "/findpwdsms", &handler.FIndPwdSendSMS{}),
		frame.NewNamedAPI("找回密码", "POST", "/findpwd", &handler.FindPwd{}),
		frame.NewNamedAPI("修改用户信息", "POST", "/editinfo", &handler.EditInfo{}),
		frame.NewNamedAPI("获取用户完成状况", "POST", "/task_status", &handler.TaskStatus{}),
		frame.NewNamedAPI("修改用户账号", "POST", "/edit_account", &handler.EditAccount{}),
		frame.NewNamedAPI("修改声纹信息", "POST", "/edit_gtspwd", &handler.EditGstpwd{}),

		//以下是用于后台获取数据
		frame.NewGroup("admin",
			frame.NewNamedAPI("获取用户列表", "POST", "/userlist", &handler.UserList{}),
			frame.NewNamedAPI("修改账号的状态", "POST", "/editstatus", &handler.EditAccountStatus{}),
			frame.NewNamedAPI("获取审核列表", "POST", "/audit_list", &handler.IdAuditList{}),
			frame.NewNamedAPI("获取用户详情", "POST", "/account_info", &handler.UserInfo{}),
			frame.NewNamedAPI("获取用户审核信息", "POST", "/account_audit", &handler.GetAccountAudit{}),
			frame.NewNamedAPI("修改用户审核状态", "POST", "/edit_audit", &handler.EditAudit{}),
			frame.NewNamedAPI("通过uid获取账号", "POST", "/getaccount", &handler.GetAccountByUids{}),
			frame.NewNamedAPI("通过账号获取uid", "POST", "/getuid", &handler.GetUidByAccount{}),
		),
		frame.NewNamedStaticFS("upload", "/upload", faygo.MarkdownFS(
			"./upload/",
		)),
	)
}
