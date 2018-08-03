package handler

import (
	"fmt"
	"github.com/henrylee2cn/faygo"
	"mime/multipart"
	"strings"
	"time"
	"user/model"
)

//注册的接口
type Register struct {
	Account     string `param:"<in:formData><required><name:account><desc:用户手机号,格式为08618022407520,前面三位为区号>"`
	Passwd      string `param:"<in:formData><required><name:passwd><desc:用户密码>"`
	Gstpwd      string `param:"<in:formData><name:gstpwd><desc:手势密码>"`
	Tradepwd    string `param:"<in:formData><name:tradepwd><desc:交易密码>"`
	Ip          string `param:"<in:formData><name:ip><desc:注册地点ip>"`
	Device_type string `param:"<in:formData><name:device_type><desc:注册设备类型:ios或android>"`
	Device_id   string `param:"<in:formData><name:device_id><desc:注册设备编码>"`
	VerifyCode  string `param:"<in:formData><required><name:verifycode><desc:验证码(暂时不验证)>"`
	Invite_num  int    `param:"<in:formData><name:invite_num><desc:邀请码>"`
}

func (r *Register) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(r)
	if err != nil {
		return jsonReturn(ctx, 0, "参数解析出错")
	}
	if r.Account == "" {
		return jsonReturn(ctx, 0, "手机号不能为空")
	}
	if !CheckPwd(r.Passwd) {
		return jsonReturn(ctx, 0, "请输入6～16位数字，字母组成的密码")
	}
	if r.Tradepwd != "" && !CheckTradepwd(r.Tradepwd) {
		return jsonReturn(ctx, 0, "请输入6位数组组成的交易密码")
	}

	if r.Gstpwd != "" && !CheckGstpwd(r.Gstpwd) {
		return jsonReturn(ctx, 0, "请输入1～24位数字组成的的手势密码")
	}

	if r.Ip != "" && !IsIP(r.Ip) {
		return jsonReturn(ctx, 0, "请输入正确的ip")
	}

	if r.Device_type != "" && r.Device_type != "android" && r.Device_type != "ios" {
		return jsonReturn(ctx, 0, "请输入正确的设备类型")
	}
	if len(r.Device_id) > 50 {
		return jsonReturn(ctx, 0, "设备编码最大不可超过50")
	}

	//检验邀请码
	if r.Invite_num > 0 && r.Invite_num < 363636 {
		return jsonReturn(ctx, 0, "不存在的邀请码")
	}

	//检验验证码
	ret, err := CheckSMS(r.Account, r.VerifyCode)
	if err != nil {
		return jsonReturn(ctx, 0, "验证短信失败，请重试")
	}
	retcode := int(ret["code"].(float64))
	if retcode == 413 {
		return jsonReturn(ctx, 0, "请输入正确的校验码")
	}
	if retcode != 200 {
		return jsonReturn(ctx, 0, "验证短信失败，请重试")
	}

	//将用户名分成两部分
	phone := strings.Split(r.Account, "-")
	if len(phone) < 2 {
		return jsonReturn(ctx, 0, "用户名格式不对")
	}
	prefix := phone[0]
	r.Account = phone[1]
	uid := CreateId(r.Account)

	r.Passwd = encrypt(r.Passwd)
	r.Gstpwd = encrypt(r.Gstpwd)
	r.Tradepwd = encrypt(r.Tradepwd)

	//入库
	err = model.DefaultUser.Register(prefix, r.Account, uid, r.Passwd, r.Gstpwd, r.Tradepwd, r.Ip, r.Device_type, r.Device_id, r.Invite_num)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, "注册成功")
}

func (r *Register) Doc() faygo.Doc {
	err := return_jonData(0, "失败")
	success := return_jonData(200, "注册成功")
	return_param := []interface{}{err, success}

	return faygo.Doc{
		Note:   "注册接口哦",
		Return: return_param,
	}
}

type FaceEdit struct {
	Token string                `param:"<in:formData><required><name:token>"`
	File  *multipart.FileHeader `param:"<in:formData><required><name:file>"`
}

func (f *FaceEdit) Serve(ctx *faygo.Context) error {
	token := ctx.FormParam("token")
	if token == "" {
		return jsonReturn(ctx, 0, "token不能为空")
	}
	//根据token获取用户信息
	userData, err := model.GetInfoByToken(token)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}

	saveInfo, err := ctx.SaveFile("file", true, "/user/"+userData.User.Account+"/"+Md5(userData.User.Uid+userData.User.Account))
	if err != nil {
		return jsonReturn(ctx, 0, "上传失败")
	}

	//判断用户的头像信息是否为空
	if userData.UserInfo.Face == "" {
		//修改用户的头像
		err = model.DefaultUserinfo.EditFace(token, userData, userData.User.Uid, saveInfo.Url)
		if err != nil {
			return jsonReturn(ctx, 0, err.Error())
		}
		//更新token中的信息

	}
	return jsonReturn(ctx, 200, saveInfo.Url+"?time="+fmt.Sprintf("%v", time.Now().Unix()))
}

type Login struct {
	Account string `param:"<in:formData><required><name:account><desc:用户手机号,格式为18022407520,不需要加前缀>"`
	Passwd  string `param:"<in:formData><required><name:passwd><desc:用户密码>"`
}

func (l *Login) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(l)
	if err != nil {
		return jsonReturn(ctx, 0, "参数解析出错")
	}
	if l.Account == "" {
		return jsonReturn(ctx, 0, "手机号不能为空")
	}
	if !CheckPwd(l.Passwd) {
		return jsonReturn(ctx, 0, "请输入6～16位数字，字母组成的密码")
	}

	//查询出该用户
	user, err := model.DefaultUser.GetUserByAccount(l.Account)
	if err != nil {
		faygo.Debug(324)
		return jsonReturn(ctx, 0, err.Error())
	}

	if user.Passwd != encrypt(l.Passwd) {
		return jsonReturn(ctx, 0, "密码不正确")
	}
	//查询出用户的详细信息
	userInfo, err := model.DefaultUserinfo.GetInfo(user.Uid)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	//生成token
	token := Md5(fmt.Sprintf("%v%v", time.Now().Unix(), l.Account))
	//将token写如数据库中
	err = model.DefaultUser.EditToken(l.Account, token)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	//将用户信息保存到缓存中
	err = model.SetToken(token, user, userInfo)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, token)
}

func (l *Login) Doc() faygo.Doc {
	err := return_jonData(0, "失败")
	success := return_jonData(200, "注册成功")
	return_param := []interface{}{err, success}

	return faygo.Doc{
		Note:   "登陆接口",
		Return: return_param,
	}
}

//发送短信验证码
type Send_SMS struct {
	Phone string `param:"<in:formData><required><name:phone><desc:phone>"`
}

func (s *Send_SMS) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(s)
	if err != nil {
		return jsonReturn(ctx, 0, "参数解析出错")
	}
	if s.Phone == "" {
		return jsonReturn(ctx, 0, "请输入手机号码")
	}
	ret, err := SendSMS(s.Phone)
	if err != nil {
		return jsonReturn(ctx, 0, "发送短信失败，请重试")
	}
	retcode := int(ret["code"].(float64))
	if retcode == 416 {
		return jsonReturn(ctx, 1, "发送失败，每天最多发送10次短信")
	}
	if retcode != 200 {
		return jsonReturn(ctx, 0, "发送短信失败，请重试")
	}
	return jsonReturn(ctx, 200, "发送成功")
}

//找回密码
type FindPwd struct {
	Account    string `param:"<in:formData><required><name:account>"`
	VerifyCode string `param:"<in:formData><required><name:verifycode><desc:验证码(暂时不验证，随便传个参数)>"`
	NewPwd     string `param:"<in:formData><required><name:passwd><desc:新密码密码>"`
}

func (f *FindPwd) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(f)
	if err != nil {
		return jsonReturn(ctx, 0, "解析参数出错")
	}
	if f.Account == "" {
		return jsonReturn(ctx, 0, "token不能为空")
	}

	if !CheckPwd(f.NewPwd) {
		return jsonReturn(ctx, 0, "请输入6～16位数字，字母组成的密码")
	}
	//根据用户token获取用户信息
	userData, err := model.DefaultUser.GetUserByAccount(f.Account)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}

	phone := userData.Prefix + "-" + userData.Account
	//检验验证码
	ret, err := CheckSMS(phone, f.VerifyCode)
	if err != nil {
		return jsonReturn(ctx, 0, "验证短信失败，请重试")
	}
	retcode := int(ret["code"].(float64))
	if retcode == 413 {
		return jsonReturn(ctx, 0, "请输入正确的校验码")
	}
	if retcode != 200 {
		return jsonReturn(ctx, 0, "验证短信失败，请重试")
	}

	//修改密码
	err = model.DefaultUser.EditPwd(userData, encrypt(f.NewPwd))
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, "操作成功")
}

//找回密码发送短信验证码
type FIndPwdSendSMS struct {
	Token string `param:"<in:formData><required><name:token>"`
}

func (f *FIndPwdSendSMS) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(f)
	if err != nil {
		return jsonReturn(ctx, 0, "解析参数出错")
	}
	if f.Token == "" {
		return jsonReturn(ctx, 0, "token不能为空")
	}
	//根据用户token获取用户信息
	userData, err := model.GetInfoByToken(f.Token)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	phone := userData.User.Prefix + "-" + userData.User.Account
	//发送短信
	ret, err := SendSMS(phone)
	if err != nil {
		return jsonReturn(ctx, 0, "发送短信失败，请重试")
	}
	retcode := int(ret["code"].(float64))
	if retcode == 416 {
		return jsonReturn(ctx, 1, "发送失败，每天最多发送10次短信")
	}
	if retcode != 200 {
		return jsonReturn(ctx, 0, "发送短信失败，请重试")
	}
	return jsonReturn(ctx, 0, "发送成功")
}

type GetUserInfo struct {
	Token string `param:"<in:formData><required><name:token>"`
}

func (g *GetUserInfo) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(g)
	if err != nil {
		return jsonReturn(ctx, 0, "解析参数出错")
	}
	if g.Token == "" {
		return jsonReturn(ctx, 0, "token不能为空")
	}
	//根据用户token获取用户信息
	userData, err := model.GetInfoByToken(g.Token)
	if err == model.LoginExpire {
		return jsonReturn(ctx, 1, err.Error())
	}
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	if userData.UserInfo.Face != "" {
		userData.UserInfo.Face = userData.UserInfo.Face + "?time=" + fmt.Sprintf("%v", time.Now().Unix())
	}

	return jsonReturn(ctx, 200, userData.UserInfo)
}

//实名认证
type Certification struct {
	Token        string                `param:"<in:formData><required><name:token>"`
	Id_name      string                `param:"<in:formData><required><name:id_name><desc:身份证姓名>"`
	Id_num       string                `param:"<in:formData><required><name:id_num><desc:身份证号码>"`
	Id_cart      *multipart.FileHeader `param:"<in:formData><required><name:id_cart><desc:身份证正面>"`
	Id_cart_back *multipart.FileHeader `param:"<in:formData><required><name:id_cart_back><desc:身份证反面>"`
}

func (c *Certification) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(c)
	if err != nil {

		return jsonReturn(ctx, 0, "解析参数出错")
	}
	if c.Token == "" {
		return jsonReturn(ctx, 0, "token不能为空")
	}
	//根据用户token获取用户信息
	userData, err := model.GetInfoByToken(c.Token)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	audit_status := userData.UserInfo.Audit_status
	//检查是否已经认证了
	if audit_status == 1 {
		return jsonReturn(ctx, 0, "正在审核中，请耐心等候")
	}
	if audit_status == 2 {
		return jsonReturn(ctx, 0, "已经审核通过")
	}
	saveInfo, err := ctx.SaveFile("id_cart", true, "/user/"+userData.User.Account+"/"+"id_cart")
	if err != nil {
		return jsonReturn(ctx, 0, "上传失败")
	}

	back_saveInfo, err := ctx.SaveFile("id_cart_back", true, "/user/"+userData.User.Account+"/"+"id_cart_back")
	if err != nil {
		return jsonReturn(ctx, 0, "上传失败")
	}
	//数据入库
	err = model.DefaultRigisterInfo.EditIdCart(userData.User.Uid, c.Id_name, c.Id_num, saveInfo.Url, back_saveInfo.Url)
	if err != nil {
		faygo.Debug(34343)
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, "提交成功")
}

//修改用户信息
type EditInfo struct {
	Nickname string `param:"<in:formData><required><name:nickname>"`
	Sex      int    `param:"<in:formData><required><name:sex>"`
	Token    string `param:"<in:formData><required><name:token>"`
}

func (e *EditInfo) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(e)
	if err != nil {
		return jsonReturn(ctx, 0, "参数解析出错")
	}

	if len([]rune(e.Nickname)) > 10 {
		return jsonReturn(ctx, 0, "昵称不能超过十个字符")
	}
	if e.Sex < 0 || e.Sex > 2 {
		return jsonReturn(ctx, 0, "非法参数")
	}
	if e.Token == "" {
		return jsonReturn(ctx, 0, "token不能为空")
	}
	//根据token获取用户信息
	userData, err := model.GetInfoByToken(e.Token)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	edit_info_status := false
	if userData.UserInfo.Info_status == 0 && userData.UserInfo.Face != "" {
		edit_info_status = true
	}
	err = model.DefaultUserinfo.EditInfo(e.Token, userData, userData.User.Uid, e.Nickname, e.Sex, edit_info_status)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, "修改成功")
}

//获取用户完成状况
type TaskStatus struct {
	Token string `param:"<in:formData><required><name:token>"`
}

func (t *TaskStatus) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(t)
	if err != nil {
		return jsonReturn(ctx, 0, "解析参数出错")
	}
	if t.Token == "" {
		return jsonReturn(ctx, 0, "token不能为空")
	}
	//根据用户token获取用户信息
	userData, err := model.GetInfoByToken(t.Token)
	if err == model.LoginExpire {
		return jsonReturn(ctx, 1, err.Error())
	}
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	data := make(map[string]interface{})
	data["audit_status"] = userData.UserInfo.Audit_status
	data["info_status"] = userData.UserInfo.Info_status
	return jsonReturn(ctx, 200, data)
}

//修改用户账号(用于测试)
type EditAccount struct {
	Account    string `param:"<in:formData><required><name:account><desc:用户手机号,格式为18022407520,不需要加前缀>"`
	NewAccount string `param:"<in:formData><required><name:new_account><desc:用户手机号,格式为18022407520,不需要加前缀>"`
}

func (e *EditAccount) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(e)
	if err != nil {
		return jsonReturn(ctx, 0, "参数解析出错")
	}
	err = model.DefaultUser.EditAccount(e.Account, e.NewAccount)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, "修改成功")
}
