package model

import (
	"github.com/go-errors/errors"
	"github.com/henrylee2cn/faygo"
	"github.com/henrylee2cn/faygo/ext/db/xorm"
)

//用于存放用户刚注册的信息的表
type UserRegisterInfo struct {
	Uid                  string `xorm:"not null default('') varchar(20) comment(用户uid)"`
	Register_ip          string `xorm:"not null default('') char(15) comment(注册ip)"`
	Register_device      int    `xorm:"not null default(0) tinyint(4) comment(注册设备,1安卓，2ios)"`
	Register_device_code string `xorm:"not null default('') varchar(50) comment(注册设备的设备编码)"`
	Id_name              string `xorm:"not null default('') varchar(255) comment('身份证名字')"`
	Id_num               string `xorm:"not null default('') varchar(255) comment('身份证号码')"`
	Id_cart              string `xorm:"not null default('') varchar(255) comment('身份证正面')"`
	Id_cart_back         string `xorm:"not null default('') varchar(255) comment('身份证反面')"`
}

const (
	UserRegisterInfo_TABLE = "user_register_info"
)

var DefaultRigisterInfo *UserRegisterInfo = new(UserRegisterInfo)

func init() {
	err := xorm.MustDB().Table(UserRegisterInfo_TABLE).Sync2(DefaultRigisterInfo)
	if err != nil {
		faygo.Error(err.Error())
	}
}

//修改身份信息
func (u *UserRegisterInfo) EditIdCart(uid, id_name, id_num, id_cart, id_cart_back string) (err error) {
	engine := xorm.MustDB()
	userRegisterInfo := new(UserRegisterInfo)
	userRegisterInfo.Id_cart = id_cart
	userRegisterInfo.Id_cart_back = id_cart_back
	userRegisterInfo.Id_name = id_name
	userRegisterInfo.Id_num = id_num
	sess := engine.NewSession()
	err = sess.Begin()
	if err != nil {
		err = SystemFail
		return
	}
	_, err = sess.Where("uid=?", uid).Cols("id_name", "id_num", "id_cart", "id_cart_back").Update(userRegisterInfo)
	if err != nil {
		err = SystemFail
		sess.Rollback()
		return
	}
	//修改用户信息里面的状态
	userinfo := new(Userinfo)
	userinfo.Audit_status = 1
	_, err = sess.Where("uid=?", uid).Cols("audit_status").Update(userinfo)
	if err != nil {
		err = SystemFail
		sess.Rollback()
		return
	}
	sess.Commit()
	return
}

func (u *UserRegisterInfo) GetAudit(uid string) (info *UserRegisterInfo, err error) {
	engine := xorm.MustDB()
	info = new(UserRegisterInfo)
	has, err := engine.Where("uid=?", uid).Get(info)
	if err != nil {
		err = SystemFail
		return
	}
	if !has {
		err = errors.New("不存在的用户")
	}
	return
}
