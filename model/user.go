package model

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/henrylee2cn/faygo"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
	"strings"
	"time"
	"user/config"
)

type User struct {
	Id       int    `xorm:"not null INT(11) pk autoincr"`
	Uid      string `xorm:"not null default('') varchar(20) comment(uid)"`
	Account  string `xorm:"not null default('') char(20) comment(账号)"`
	Prefix   string `xorm:"not null default('') char(6) comment('前缀')"`
	Passwd   string `xorm:"not null default('') char(32) comment(密码)"`
	Gstpwd   string `xorm:"not null default('') char(32) comment(手势密码)"`
	Tradepwd string `xorm:"not null default('') char(32) comment(交易密码)"`
	Token    string `xorm:"not null default('') char(32) comment('用户登陆的token')"`
	Status   int    `xorm:"not null default(0) tinyint(4) comment('状态，0为正常，1为封号')"`
}

const (
	User_TABLE = "user"
)

var DefaultUser *User = new(User)

func init() {
	err := xorm.MustDB().Table(User_TABLE).Sync2(DefaultUser)
	if err != nil {
		faygo.Error(err.Error())
	}
}

func (u *User) GetUserByAccount(account string) (user User, err error) {
	engine := xorm.MustDB()
	user = User{}
	//查询是否存在
	has, err := engine.Where("account=?", account).Get(&user)
	if err != nil {
		err = SystemFail
		return
	}
	if !has {
		err = errors.New("用户不存在")
		return
	}
	if user.Status == 1 {
		err = errors.New("该账号已经被封停")
	}
	return
}

//注册用户
func (u *User) Register(prefix, account, uid, passwd, gstpwd, tradepwd, ip, Device_type, Device_id string, invite_num int) (err error) {
	engine := xorm.MustDB()
	//查询是否存在
	has, err := engine.Table(User_TABLE).Where("account=?", account).Exist()
	if err != nil {
		err = SystemFail
		return
	}
	if has {
		err = errors.New("用户已存在")
		return
	}
	sess := engine.NewSession()
	defer sess.Close()
	err = sess.Begin()
	if err != nil {
		err = SystemFail
		return
	}
	user := new(User)
	user.Prefix = prefix
	user.Account = account
	user.Passwd = passwd
	user.Gstpwd = gstpwd
	user.Tradepwd = tradepwd
	user.Uid = uid
	n, err := sess.Insert(user)
	if err != nil || n == 0 {
		err = SystemFail
		sess.Rollback()
		return
	}
	//添加用户详细信息
	userinfo := new(Userinfo)
	userinfo.Uid = uid
	userinfo.Invite_num = user.Id + 363636
	n, err = sess.Insert(userinfo)
	if err != nil || n == 0 {
		faygo.Debug(err)
		err = SystemFail
		sess.Rollback()
		return
	}

	//添加用户注册信息
	userRegisterInfo := new(UserRegisterInfo)
	userRegisterInfo.Uid = uid
	userRegisterInfo.Register_ip = ip
	userRegisterInfo.Register_device_code = Device_id
	switch Device_type {
	case "android":
		userRegisterInfo.Register_device = 1
	case "ios":
		userRegisterInfo.Register_device = 2
	default:
		userRegisterInfo.Register_device = 0
	}
	n, err = sess.Insert(userRegisterInfo)
	if err != nil || n == 0 {
		err = SystemFail
		sess.Rollback()
		return
	}

	if invite_num > 0 {
		//根据邀请码获取邀请人的id
		if invite_num < 363636 {
			err = errors.New("不存在的邀请码")
			return
		}
		id := invite_num - 363636
		var uid string
		user := new(User)
		has, err := sess.Table(User_TABLE).Where("id=?", id).Cols("uid", "token").Get(user)
		if err != nil {
			err = SystemFail
			sess.Rollback()
			return err
		}
		if !has {
			err = errors.New("不存在的邀请码")
			sess.Rollback()
			return err
		}
		uid = user.Uid
		_, err = sess.Exec("update userinfo set invite_man_num=invite_man_num+1 where uid=?", uid)
		if err != nil {
			err = SystemFail
			sess.Rollback()
			return err
		}
		//增加一条算力增加的记录
		content := "邀请好友增加算力"
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
		param["type"] = "3"
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
		//修改该用户的token中的信息
		userData, err := GetInfoByToken(user.Token)
		if err == LoginExpire {
			err = nil
			sess.Commit()
			return err
		}
		if err != nil {
			err = SystemFail
			sess.Rollback()
			return err
		}
		userData.UserInfo.Invite_man_num = userData.UserInfo.Invite_man_num + 1
		err = EditToken(user.Token, userData)
		if err != nil {
			err = SystemFail
			sess.Rollback()
			return err
		}
	}
	sess.Commit()
	return
}

//修改密码
func (u *User) EditPwd(userData User, pwd string) (err error) {
	engine := xorm.MustDB()
	sess := engine.NewSession()
	err = sess.Begin()
	if err != nil {
		err = SystemFail
		return
	}
	id := userData.Id
	user := new(User)
	user.Passwd = pwd
	_, err = sess.Where("id=?", id).Cols("passwd").Update(user)
	if err != nil {
		err = SystemFail
		sess.Rollback()
		return
	}

	sess.Commit()
	return
}

type UserGroup struct {
	User     `xorm:"extends"`
	Userinfo `xorm:"extends"`
}

func (UserGroup) TableName() string {
	return "user"
}

//获取用户列表
func (u *User) List(offset, limit int) (list []UserGroup, err error) {
	engine := xorm.MustDB()
	userGroup := new(UserGroup)
	cols := []string{"user.id", "user.status", "user.account", "user.uid", "userinfo.*"}
	rows, err := engine.Desc("id").Cols(cols...).Join("inner", "userinfo", "user.uid=userinfo.uid").Limit(limit, offset).Rows(userGroup)
	if err != nil {
		err = SystemFail
		return
	}
	defer rows.Close()
	for rows.Next() {
		userGroup := UserGroup{}
		err = rows.Scan(&userGroup)
		if err != nil {
			err = SystemFail
			return
		}
		list = append(list, userGroup)
	}
	return
}

//获取用户的总数
func (u *User) Count() (count int64, err error) {
	engine := xorm.MustDB()
	count, err = engine.Count(u)
	if err != nil {
		err = SystemFail
	}
	return
}

func (u *User) EditStatus(account string, status int) (err error) {
	engine := xorm.MustDB()
	sess := engine.NewSession()
	err = sess.Begin()
	if err != nil {
		err = SystemFail
		return
	}
	user := new(User)
	user.Status = status
	_, err = sess.Where("account=?", account).Cols("status").Update(user)
	if err != nil {
		err = SystemFail
		sess.Rollback()
		return
	}
	//清空session
	if status == 1 {
		//查询出token
		ue, err := u.GetUserByAccount(account)
		if err != nil {
			sess.Rollback()
			return err
		}
		err = ClearToken(ue.Token)
		if err != nil {
			sess.Rollback()
			return err
		}
	}
	sess.Commit()
	return
}

type AccountInfo struct {
	User             `xorm:"extends"`
	Userinfo         `xorm:"extends"`
	UserRegisterInfo `xorm:"extends"`
}

func (AccountInfo) TableName() string {
	return "user"
}

//获取用户详细信息
func (u *User) GetUserInfoByAccount(account string) (accountInfo *AccountInfo, err error) {
	engine := xorm.MustDB()
	accountInfo = new(AccountInfo)
	cols := []string{"user.id", "user.account", "userinfo.*", "user_register_info.*"}
	has, err := engine.Cols(cols...).Where("account=?", account).Join("inner", "userinfo", "user.uid=userinfo.uid").Join("inner", "user_register_info", "user_register_info.uid=user.uid").Get(accountInfo)
	if err != nil {
		faygo.Debug(err)
		err = SystemFail
		return
	}
	if !has {
		err = errors.New("不存在的用户")
	}
	return
}

//通过uid字符串（uid以逗号分隔）获取账号
func (u *User) GetAccountByUids(uids string) (account map[string]string, err error) {
	engine := xorm.MustDB()
	rows, err := engine.In("uid", strings.Split(uids, ",")).Cols("account", "uid").Rows(u)
	if err != nil {
		err = SystemFail
		return
	}
	defer rows.Close()
	account = make(map[string]string)
	for rows.Next() {
		user := new(User)
		err = rows.Scan(user)
		if err != nil {
			err = SystemFail
			return
		}
		account[user.Uid] = user.Account
	}
	return
}

var NotAccount = errors.New("不存在的账户")

//通过账户获取uid
func (u *User) GetUidByAccount(account string) (uid string, err error) {
	engine := xorm.MustDB()
	has, err := engine.Table(User_TABLE).Where("account=?", account).Cols("uid").Get(&uid)
	if err != nil {
		err = SystemFail
		return
	}
	if !has {
		err = NotAccount
	}
	return
}

//修改token
func (u *User) EditToken(account string, token string) (err error) {
	engine := xorm.MustDB()
	user := new(User)
	user.Token = token
	n, err := engine.Where("account=?", account).Cols("token").Update(user)
	if err != nil {
		err = SystemFail
		return
	}
	if n == 0 {
		err = errors.New("登陆失败")
	}
	return
}

//通过uid获取用户的token
func (u *User) GetTokenByUid(uid string) (token string, err error) {
	engine := xorm.MustDB()
	has, err := engine.Table(User_TABLE).Where("uid=?", uid).Cols("token").Get(&token)
	if err != nil {
		err = SystemFail
		return
	}
	if !has {
		err = errors.New("不存在的用户")
	}
	return
}
