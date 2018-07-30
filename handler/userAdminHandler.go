package handler

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/henrylee2cn/faygo"
	"user/model"
)

//管理后台获取用户信息
type UserList struct {
	Offset int    `param:"<in:formData><required>"`
	Limit  int    `param:"<in:formData><required>"`
	Time   int64  `param:"<in:formData><required>"`
	Sign   string `param:"<in:formData><required>"`
}

func (u *UserList) Serve(ctx *faygo.Context) error {
	err := u.Check(ctx)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}

	//获取用户的总数
	count, err := model.DefaultUser.Count()
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	var list []model.UserGroup
	if count > 0 {
		//获取用户列表
		list, err = model.DefaultUser.List(u.Offset, u.Limit)
		if err != nil {
			return jsonReturn(ctx, 0, err.Error())
		}
	}

	return jsonReturn(ctx, 200, list, count)
}

//检查是否合法
func (u *UserList) Check(ctx *faygo.Context) (err error) {
	err = ctx.BindForm(u)
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	b, err := RsaDecrypt([]byte(u.Sign))
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	if string(b) != Md5(fmt.Sprintf("%v%v%v", u.Offset, u.Limit, u.Time)) {
		err = errors.New("非法参数")
	}
	return
}

//获取用户详情
type UserInfo struct {
	Account string `param:"<in:formData><required><name:account>"`
	Time    int64  `param:"<in:formData><required><name:time>"`
	Sign    string `param:"<in:formData><required><name:sign>"`
}

func (u *UserInfo) Serve(ctx *faygo.Context) error {
	err := u.Check(ctx)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	//通过用户账号获取用户详细信息
	account, err := model.DefaultUser.GetUserInfoByAccount(u.Account)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, account)
}

//检查是否合法
func (u *UserInfo) Check(ctx *faygo.Context) (err error) {
	err = ctx.BindForm(u)
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	b, err := RsaDecrypt([]byte(u.Sign))
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	if string(b) != Md5(fmt.Sprintf("%v%v", u.Account, u.Time)) {
		err = errors.New("非法参数")
	}
	return
}

//修改用户状态
type EditAccountStatus struct {
	Account string `param:"<in:formData><required><name:account>"`
	Typeid  int    `param:"<in:formData><required><name:type>"`
	Time    int64  `param:"<in:formData><required><name:time>"`
	Sign    string `param:"<in:formData><required><name:sign>"`
}

func (e *EditAccountStatus) Serve(ctx *faygo.Context) error {
	err := e.Check(ctx)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	//根据用户账号修改用户信息
	var status int
	switch e.Typeid {
	case 1:
		status = 1
	case 2:
		status = 0
	default:
		return jsonReturn(ctx, 0, "非法参数")
	}
	err = model.DefaultUser.EditStatus(e.Account, status)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, "操作成功")
}

//检查是否合法
func (e *EditAccountStatus) Check(ctx *faygo.Context) (err error) {
	err = ctx.BindForm(e)
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	b, err := RsaDecrypt([]byte(e.Sign))
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	if string(b) != Md5(fmt.Sprintf("%v%v%v", e.Account, e.Typeid, e.Time)) {
		err = errors.New("非法参数")
	}
	return
}

//获取身份审核列表
type IdAuditList struct {
	Offset int    `param:"<in:formData><required>"`
	Limit  int    `param:"<in:formData><required>"`
	Time   int64  `param:"<in:formData><required>"`
	Sign   string `param:"<in:formData><required>"`
}

func (i *IdAuditList) Serve(ctx *faygo.Context) error {
	err := i.Check(ctx)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}

	//获取需要身份审核的总数
	count, err := model.DefaultUserinfo.AuditCount()
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	var list []*model.UserAudit
	if count > 0 {
		list, err = model.DefaultUserinfo.AuditList(i.Offset, i.Limit)
		if err != nil {
			return jsonReturn(ctx, 0, err.Error())
		}
	}
	return jsonReturn(ctx, 200, list, count)
}

//检查是否合法
func (u *IdAuditList) Check(ctx *faygo.Context) (err error) {
	err = ctx.BindForm(u)
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	b, err := RsaDecrypt([]byte(u.Sign))
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	if string(b) != Md5(fmt.Sprintf("%v%v%v", u.Offset, u.Limit, u.Time)) {
		err = errors.New("非法参数")
	}
	return
}

//获取单个用户的审核信息
type GetAccountAudit struct {
	Uid  string `param:"<in:formData><required><name:uid>"`
	Time int64  `param:"<in:formData><required><name:time>"`
	Sign string `param:"<in:formData><required><name:sign>"`
}

func (g *GetAccountAudit) Serve(ctx *faygo.Context) error {
	err := g.Check(ctx)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}

	//判断用户是否已经审核过
	ok, err := model.DefaultUserinfo.CheckAudit(g.Uid)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	if !ok {
		return jsonReturn(ctx, 0, "该用户已经审核通过")
	}
	//获取用户的身份信息
	info, err := model.DefaultRigisterInfo.GetAudit(g.Uid)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, info)
}

//检查是否合法
func (g *GetAccountAudit) Check(ctx *faygo.Context) (err error) {
	err = ctx.BindForm(g)
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	b, err := RsaDecrypt([]byte(g.Sign))
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	if string(b) != Md5(fmt.Sprintf("%v%v", g.Uid, g.Time)) {
		err = errors.New("非法参数")
	}
	return
}

//修改用户的审核状态
type EditAudit struct {
	Uid    string `param:"<in:formData><required><name:uid>"`
	Action int    `param:"<in:formData><name:action>"`
	Time   int64  `param:"<in:formData><required><name:time>"`
	Sign   string `param:"<in:formData><required><name:sign>"`
}

func (e *EditAudit) Serve(ctx *faygo.Context) error {
	err := e.Check(ctx)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	var audit_stauts int
	switch e.Action {
	case 1:
		audit_stauts = 2
	case 2:
		audit_stauts = 3
	default:
		return jsonReturn(ctx, 0, "非法参数")
	}

	//判断用户是否已经审核过
	ok, err := model.DefaultUserinfo.CheckAudit(e.Uid)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	if !ok {
		return jsonReturn(ctx, 0, "该用户已经审核通过")
	}

	//修改用户的审核状态
	err = model.DefaultUserinfo.EditAudit(e.Uid, audit_stauts)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}

	return jsonReturn(ctx, 200, "操作成功")
}

//检查是否合法
func (e *EditAudit) Check(ctx *faygo.Context) (err error) {
	err = ctx.BindForm(e)
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	b, err := RsaDecrypt([]byte(e.Sign))
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	if string(b) != Md5(fmt.Sprintf("%v%v%v", e.Uid, e.Action, e.Time)) {
		err = errors.New("非法参数")
	}
	return
}

//获取用户的uid
type GetUidByAccount struct {
	Account string `param:"<in:formData><required><name:account>"`
	Time    int64  `param:"<in:formData><required><name:time>"`
	Sign    string `param:"<in:formData><required><name:sign>"`
}

func (g *GetUidByAccount) Serve(ctx *faygo.Context) error {
	err := g.Check(ctx)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	//根据用户账户获取用户的uid
	uid, err := model.DefaultUser.GetUidByAccount(g.Account)
	if err == model.NotAccount {
		return jsonReturn(ctx, -1, err.Error())
	}
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, uid)
}

//检查是否合法
func (g *GetUidByAccount) Check(ctx *faygo.Context) (err error) {
	err = ctx.BindForm(g)
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	b, err := RsaDecrypt([]byte(g.Sign))
	if err != nil {
		err = errors.New("参数解析出错")
		return
	}
	if string(b) != Md5(fmt.Sprintf("%v%v", g.Account, g.Time)) {
		err = errors.New("非法参数")
	}
	return
}

//根据uid字符串（以逗号分隔）获取用户账号
type GetAccountByUids struct {
	Uids string `param:"<in:formData><required><name:uids>"`
}

func (g *GetAccountByUids) Serve(ctx *faygo.Context) error {
	err := ctx.BindForm(g)
	if err != nil {
		return jsonReturn(ctx, 0, "参数解析出错")
	}
	account, err := model.DefaultUser.GetAccountByUids(g.Uids)
	if err != nil {
		return jsonReturn(ctx, 0, err.Error())
	}
	return jsonReturn(ctx, 200, account)
}
