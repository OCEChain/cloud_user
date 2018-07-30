package model

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"user/redis"
)

var LoginExpire = errors.New("用户登陆失效")

//通过token获取存放在缓存中的所有用户信息
func GetInfoByToken(token string) (userData UserData, err error) {
	//根据token获取用户信息
	data, err := redis.NewRedis().Get(token)
	if err != nil {
		err = SystemFail
		return
	}
	if data == "" {
		err = LoginExpire
		return
	}
	//如果数据不为空,则返回用户数据
	err = json.Unmarshal([]byte(data), &userData)
	if err != nil {
		err = SystemFail
		return
	}
	return
}

type UserData struct {
	User     User
	UserInfo Userinfo
}

func SetToken(token string, user User, userInfo Userinfo) (err error) {
	userData := UserData{}
	userData.User = user
	userData.UserInfo = userInfo

	//放到缓存中
	_, err = redis.NewRedis().Set(token, userData, "21600") //设置成6小时后才失效
	if err != nil {
		err = SystemFail
	}
	return
}

//重新设置usertoken
func EditToken(token string, userData UserData) (err error) {
	_, err = redis.NewRedis().Set(token, userData, "21600") //设置成6小时后才失效
	if err != nil {
		err = SystemFail
	}
	return
}

//清空token
func ClearToken(token string) (err error) {
	err = redis.NewRedis().Del(token)
	if err != nil {
		err = SystemFail
	}
	return
}
