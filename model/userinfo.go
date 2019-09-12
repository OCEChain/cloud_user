package model

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/henrylee2cn/faygo"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	"time"
	"user/config"
)

//用户详细信息
type Userinfo struct {
	Uid            string `xorm:"not null default('') varchar(20) comment(用户uid)"`
	Nickname       string `xorm:"not null default('') varchar(50) comment(昵称)"`
	Face           string `xorm:"not null default('') varchar(255) comment(头像)"`
	Sex            int    `xorm:"not null default(0) tinyint(4) comment(性别1为男0为保密2为女)"`
	Audit_status   int    `xorm:"not null default(0) tinyint(4) comment('审核状态 0未审核 1正在审核 2审核通过 3审核未通过')"`
	Info_status    int    `xorm:"not null default(0) tinyint(4) comment('信息完善状态 0未完善 1以完善')"`
	Invite_num     int    `xorm:"not null default(0) int(11) comment('邀请码')"`
	Invite_man_num int    `xorm:"not null default(0) int(11) comment('邀请人数')"`
	Invite_url     string `xorm:"-"`
}

const (
	Userinfo_TABLE = "userinfo"
)

var DefaultUserinfo *Userinfo = new(Userinfo)

func init() {
	err := xorm.MustDB().Table(Userinfo_TABLE).Sync2(DefaultUserinfo)
	if err != nil {
		faygo.Error(err.Error())
	}
}

//通过uid获取所有的详细信息
func (u *Userinfo) GetInfo(uid string) (userInfo Userinfo, err error) {
	engine := xorm.MustDB()
	userInfo = Userinfo{}
	has, err := engine.Where("uid=?", uid).Get(&userInfo)
	if err != nil {
		faygo.Info("通过uid查询用户出现一个意外错误，错误信息为:", err.Error())
		err = SystemFail
		return
	}
	if !has {
		err = errors.New("不存在的用户")
	}
	userInfo.Invite_url = fmt.Sprintf("/invite?num=%v", userInfo.Invite_num)
	return
}

//修改头像信息
func (u *Userinfo) EditFace(token string, userData UserData, uid string, face string) (err error) {
	engine := xorm.MustDB()
	sess := engine.NewSession()
	defer sess.Close()
	err = sess.Begin()
	if err != nil {
		err = SystemFail
		return
	}
	userInfo := new(Userinfo)
	userInfo.Face = face + "?time=" + fmt.Sprintf("%v", time.Now().Unix())
	_, err = sess.Where("uid=?", uid).Cols("face").Update(userInfo)
	if err != nil {
		faygo.Info(err)
		err = SystemFail
		sess.Rollback()
		return
	}

	userData.UserInfo.Face = userInfo.Face
	err = EditToken(token, userData)
	if err != nil {
		faygo.Info(err)
		err = SystemFail
		sess.Rollback()
		return
	}
	sess.Commit()
	return
}

func (u *Userinfo) EditInfo(token string, userData UserData, uid, nickname string, sex int, edit_info_status bool) (err error) {
	engine := xorm.MustDB()
	sess := engine.NewSession()
	defer sess.Close()
	err = sess.Begin()
	if err != nil {
		faygo.Info(err)
		err = SystemFail
		return
	}
	userInfo := new(Userinfo)
	userInfo.Nickname = nickname
	userInfo.Sex = sex
	cols := []string{"nickname", "sex"}
	if edit_info_status {
		userInfo.Info_status = 1
		cols = append(cols, "info_status")
	}
	_, err = sess.Where("uid=?", uid).Cols(cols...).Update(userInfo)
	if err != nil {
		faygo.Info(err)
		err = SystemFail
		sess.Rollback()
		return
	}

	if edit_info_status {
		content := "完善用户信息增加算力"
		t := time.Now().Unix()
		//添加算力提升记录
		url := config.GetConfig("wallet", "url").String() + "/admin/add_cal_info"
		b, err := RsaEncrypt([]byte(Md5(fmt.Sprintf("%v%v%v", uid, content, t))))
		if err != nil {
			sess.Rollback()
			return err
		}
		param := make(map[string]string)
		param["uid"] = uid
		param["content"] = content
		param["num"] = fmt.Sprintf("%v", 10)
		param["time"] = fmt.Sprintf("%v", t)
		param["type"] = "1"
		param["sign"] = string(b)
		data, err := curl_post(url, param, time.Second*2)
		if err != nil {
			sess.Rollback()
			return err
		}
		if data.Code != 200 {
			err = errors.New(data.Info)
			sess.Rollback()
			return err
		}
	}
	userData.UserInfo.Nickname = nickname
	userData.UserInfo.Sex = sex
	if edit_info_status {
		userData.UserInfo.Info_status = 1
	}
	err = EditToken(token, userData)
	if err != nil {
		err = SystemFail
		sess.Rollback()
		return
	}

	sess.Commit()
	return
}

type UserAudit struct {
	User             `xorm:"extends"`
	Userinfo         `xorm:"extends"`
	UserRegisterInfo `xorm:"extends"`
	Uid              string
}

func (UserAudit) TableName() string {
	return "userinfo"
}

//获取需要身份认证的用户信息
func (u *Userinfo) AuditList(offset, limit int) (list []*UserAudit, err error) {
	engine := xorm.MustDB()
	userAudit := new(UserAudit)
	cols := []string{"user.id", "userinfo.uid", "userinfo.nickname", "user.account", "userinfo.audit_status", "user_register_info.*"}
	rows, err := engine.Cols(cols...).Where("userinfo.audit_status=?", 1).OrderBy("user.id").Limit(limit, offset).Join("inner", User_TABLE, "user.uid=userinfo.uid").Join("inner", UserRegisterInfo_TABLE, "user_register_info.uid=userinfo.uid").Rows(userAudit)
	if err != nil {
		err = SystemFail
		return
	}
	defer rows.Close()
	for rows.Next() {
		ua := new(UserAudit)
		err = rows.Scan(ua)
		if err != nil {
			err = SystemFail
			return
		}
		faygo.Debug(ua)
		ua.Uid = ua.User.Uid
		list = append(list, ua)
	}

	return
}

//获取需要身份认证的用户的总数
func (u *Userinfo) AuditCount() (count int64, err error) {
	engine := xorm.MustDB()
	userInfo := new(Userinfo)
	count, err = engine.Where("audit_status=?", 1).Count(userInfo)
	if err != nil {
		err = SystemFail
	}
	return
}

//判断用户是否已经审核过了
func (u *Userinfo) CheckAudit(uid string) (ok bool, err error) {
	engine := xorm.MustDB()
	var isok int
	has, err := engine.Table(Userinfo_TABLE).Where("uid=?", uid).Cols("audit_status").Get(&isok)
	if err != nil {
		err = SystemFail
		return
	}
	if !has {
		err = errors.New("不存在的记录")
		return
	}
	if isok == 1 {
		ok = true
	}

	return
}

//修改用户的审核状态
func (u *Userinfo) EditAudit(uid string, status int) (err error) {
	engine := xorm.MustDB()

	//通过uid获取该用户的token
	token, err := DefaultUser.GetTokenByUid(uid)
	if err != nil {
		return
	}

	sess := engine.NewSession()
	defer sess.Close()
	err = sess.Begin()
	if err != nil {
		err = SystemFail
		return
	}
	userInfo := new(Userinfo)
	userInfo.Audit_status = status
	faygo.Debug("status:", status)
	_, err = sess.Where("uid=?", uid).Update(userInfo)
	if err != nil {
		err = SystemFail
		sess.Rollback()
		return
	}

	//如果是审核通过，则增加一条算力增加记录
	if status == 2 {
		content := "身份信息审核通过增加算力"
		t := time.Now().Unix()
		//添加算力提升记录
		url := config.GetConfig("wallet", "url").String() + "/admin/add_cal_info"
		b, err := RsaEncrypt([]byte(Md5(fmt.Sprintf("%v%v%v", uid, content, t))))
		if err != nil {
			sess.Rollback()
			return err
		}
		param := make(map[string]string)
		param["uid"] = uid
		param["content"] = content
		param["num"] = fmt.Sprintf("%v", 20)
		param["time"] = fmt.Sprintf("%v", t)
		param["type"] = "2"
		param["sign"] = string(b)
		data, err := curl_post(url, param, time.Second*2)
		if err != nil {
			sess.Rollback()
			return err
		}
		if data.Code != 200 {
			err = errors.New(data.Info)
			sess.Rollback()
			return err
		}
	}

	//修改用户保存在缓存中的数据
	userData, err := GetInfoByToken(token)
	if err == LoginExpire {
		err = nil
		sess.Commit()
		return
	}
	if err != nil {
		sess.Rollback()
		return
	}
	userData.UserInfo.Audit_status = status
	err = EditToken(token, userData)
	if err != nil {
		sess.Rollback()
		return
	}
	sess.Commit()
	return
}
