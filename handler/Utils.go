package handler

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"io"
	r "math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"
	"github.com/lessgo/lessgoext/surfer"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"net/http"
	"net/url"
	"unsafe"
)

// JsonpToJson modify jsonp string to json string
// Example: forbar({a:"1",b:2}) to {"a":"1","b":2}
func JsonpToJson(json string) string {
	start := strings.Index(json, "{")
	end := strings.LastIndex(json, "}")
	start1 := strings.Index(json, "[")
	if start1 > 0 && start > start1 {
		start = start1
		end = strings.LastIndex(json, "]")
	}
	if end > start && end != -1 && start != -1 {
		json = json[start : end+1]
	}
	json = strings.Replace(json, "\\'", "", -1)
	regDetail, _ := regexp.Compile("([^\\s\\:\\{\\,\\d\"]+|[a-z][a-z\\d]*)\\s*\\:")
	return regDetail.ReplaceAllString(json, "\"$1\":")
}

// simple xml to string  support utf8
func XML2mapstr(xmldoc string) map[string]string {
	var t xml.Token
	var err error
	inputReader := strings.NewReader(xmldoc)
	decoder := xml.NewDecoder(inputReader)
	decoder.CharsetReader = func(s string, r io.Reader) (io.Reader, error) {
		return charset.NewReader(r, s)
	}
	m := make(map[string]string, 32)
	key := ""
	for t, err = decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		case xml.StartElement:
			key = token.Name.Local
		case xml.CharData:
			content := string([]byte(token))
			m[key] = content
		default:
			// ...
		}
	}

	return m
}

//string to hash
func MakeHash(s string) string {
	const IEEE = 0xedb88320
	var IEEETable = crc32.MakeTable(IEEE)
	hash := fmt.Sprintf("%x", crc32.Checksum([]byte(s), IEEETable))
	return hash
}

func HashString(encode string) uint64 {
	hash := fnv.New64()
	hash.Write([]byte(encode))
	return hash.Sum64()
}

//string to sha1
func MakeSha1(s string) string {
	//产生一个散列值得方式是 sha1.New()，sha1.Write(bytes)，然后 sha1.Sum([]byte{})。这里我们从一个新的散列开始。
	h := sha1.New()
	//写入要处理的字节。如果是一个字符串，需要使用[]byte(s) 来强制转换成字节数组。
	h.Write([]byte(s))
	//这个用来得到最终的散列值的字符切片。Sum 的参数可以用来都现有的字符切片追加额外的字节切片：一般不需要要。
	bs := h.Sum(nil)
	//SHA1 值经常以 16 进制输出，例如在 git commit 中。使用%x 来将散列结果格式化为 16 进制字符串。
	sha1 := fmt.Sprintf("%x", bs)
	return sha1
}

// 制作特征值方法一
func MakeUnique(obj interface{}) string {
	baseString, _ := json.Marshal(obj)
	return strconv.FormatUint(HashString(string(baseString)), 10)
}

// 制作特征值方法二
func MakeMd5(b []byte) string {
	h := md5.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

// 制作特征值方法三
func MakeMd5ForObj(obj interface{}) string {
	h := md5.New()
	baseString, _ := json.Marshal(obj)
	h.Write([]byte(baseString))
	return hex.EncodeToString(h.Sum(nil))
}

// 将对象转为json字符串
func JsonString(obj interface{}) string {
	b, _ := json.Marshal(obj)
	s := fmt.Sprintf("%+v", string(b))
	r := strings.Replace(s, `\u003c`, "<", -1)
	r = strings.Replace(r, `\u003e`, ">", -1)
	return r
}

// RandomCreateBytes generate random []byte by specify chars.
func RandomCreateBytes(n int, alphabets ...byte) []byte {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	var randby bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		r.Seed(time.Now().UnixNano())
		randby = true
	}
	for i, b := range bytes {
		if len(alphabets) == 0 {
			if randby {
				bytes[i] = alphanum[r.Intn(len(alphanum))]
			} else {
				bytes[i] = alphanum[b%byte(len(alphanum))]
			}
		} else {
			if randby {
				bytes[i] = alphabets[r.Intn(len(alphabets))]
			} else {
				bytes[i] = alphabets[b%byte(len(alphabets))]
			}
		}
	}
	return bytes
}

//生产MD5

func Md5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Sha(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	return mdStr
}

//从url上面获取Value，可以获取到特殊字符，原生方法不行
func UrlToMap(str string) map[string]string {
	urlMap := map[string]string{}
	if strings.Contains(str, "&") {
		urlParam := strings.Split(str, "&")
		for _, v := range urlParam {
			param := strings.Split(v, "=")
			urlMap[param[0]] = param[1]
		}
	} else if strings.Contains(str, "=") {

		param := strings.Split(str, "=")
		urlMap[param[0]] = param[1]

	}

	return urlMap
}

//判断字符是否在切片里面
func StringInSlice(needle string, haystack []string) bool {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}
	return false
}

//生产随机数字
func RandNum(n int) string {
	time.Sleep(time.Microsecond)
	tmp := r.New(r.NewSource(time.Now().UnixNano()))
	var letters = []byte("0123456789")
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[tmp.Intn(len(letters))]
	}
	return string(b)
}

//生产随机字母
func RandSeq(n int) string {
	time.Sleep(time.Microsecond)
	tmp := r.New(r.NewSource(time.Now().UnixNano()))
	var letters = []byte("abcdefghijklmnopqrstuvwxyz")
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[tmp.Intn(len(letters))]
	}
	return string(b)
}

//生产随机字符串，包括字母数字
func RandStr(n int) string {
	time.Sleep(time.Microsecond)
	tmp := r.New(r.NewSource(time.Now().UnixNano()))
	var letters = []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[tmp.Intn(len(letters))]
	}
	return string(b)
}

func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

func CutHtmlcode(src string) string {
	//将HTML标签全转换成小写
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)

	//去除STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")

	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")

	return strings.TrimSpace(src)
}

func IsIP(ip string) (b bool) {
	if m, _ := regexp.MatchString("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$", ip); !m {
		return false
	}
	return true
}

func IsMobile(mobile string) (b bool) {
	if m, _ := regexp.MatchString("^(13[0-9]|14[0-9]|15[0-9]|16[0-9]|17[0-9]|18[0-9]|19[0-9])[0-9]{8}$", mobile); !m {
		return false
	}
	return true
}

func IsWorldMobile(mobile string) (b bool) {
	if strings.HasPrefix(mobile, "086") {
		return IsMobile(mobile)
	}
	if m, _ := regexp.MatchString(`^([0\+]\d{2,3})?(\d{3,11})$`, mobile); !m {
		return false
	}
	return true
}

func IsEmail(email string) (b bool) {
	if m, _ := regexp.MatchString(`^([\w\.\_]{2,11})@(\w{1,}).([a-z]{2,4})$`, email); !m {
		return false
	}
	return true
}

func IsDate(date string) (b bool) {
	if m, _ := regexp.MatchString(`^2[0-9]{3}\-[0-1][0-9]\-[0-3][0-9]$`, date); !m {
		return false
	}
	return true
}

func CheckPwd(pwd string) (b bool) {
	if m, _ := regexp.MatchString("^[0-9a-zA-Z_]{6,16}$", pwd); !m {
		return false
	}
	return true
}

func CheckTradepwd(pwd string) (b bool) {
	if m, _ := regexp.MatchString("^[0-9]{6}$", pwd); !m {
		return false
	}
	return true
}

func CheckGstpwd(pwd string) (b bool) {
	if m, _ := regexp.MatchString("^[0-8]{1,24}$", pwd); !m {
		return false
	}
	return true
}

func String2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func CreateId(s string) string {
	h := md5.New()
	h.Write(String2Bytes(s + strconv.FormatInt(time.Now().UnixNano(), 10)))
	src := h.Sum(nil)
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	a := r.Intn(17)
	dst = dst[a : a+16]
	ret := make([]byte, 16)
	for m, n := range r.Perm(16) {
		ret[m] = dst[n]
	}
	return Bytes2String(ret)
}

func encrypt(data string) string {
	return Md5(Sha(data))
}

const (
	NETEASE_APPKEY    = "1b5a7b8e764452c4d100ed2f37b99f96"
	NETEASE_APPSECRET = "ee105cd14328"
)

// 加密
func encryptSMS(AppSecret, nonce, curtime string) (checksum string) {
	str := AppSecret + nonce + curtime
	return MakeSha1(str)
}

// 发送短信验证码接口
func SendSMS(phone string) (map[string]interface{}, error) {
	if !strings.HasPrefix(phone, "+") {
		phone = phone + "+"
	}
	nonce := RandStr(8)
	curtime_int64 := time.Now().Unix()
	curtime := strconv.FormatInt(curtime_int64, 10)
	checksum := encryptSMS(NETEASE_APPSECRET, nonce, curtime)
	fmt.Printf("checksum: %#v", checksum)
	v := url.Values{
		"mobile": []string{phone},
	}
	return postCall("https://api.netease.im/sms/sendcode.action", v.Encode(), NETEASE_APPKEY, nonce, curtime, checksum)
}

// 验证短信验证码接口
func CheckSMS(phone, code string) (map[string]interface{}, error) {
	if !strings.HasPrefix(phone, "+") {
		phone = phone + "+"
	}
	nonce := RandStr(8)
	curtime_int64 := time.Now().Unix()
	curtime := strconv.FormatInt(curtime_int64, 10)
	checksum := encryptSMS(NETEASE_APPSECRET, nonce, curtime)

	v := url.Values{
		"mobile": []string{phone},
		"code":   []string{code},
	}
	return postCall("https://api.netease.im/sms/verifycode.action", v.Encode(), NETEASE_APPKEY, nonce, curtime, checksum)
}

// 远程请求
func postCall(url, postdata, AppKey, Nonce, CurTime, CheckSum string) (map[string]interface{}, error) {
	resp, err := surfer.Download(&surfer.DefaultRequest{
		Url:          url,
		Method:       "POST",
		PostData:     postdata,
		EnableCookie: true,
		Header: http.Header{
			"User-Agent":   []string{"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36"},
			"Content-Type": []string{"application/x-www-form-urlencoded;charset=utf-8"},
			"AppKey":       []string{AppKey},
			"Nonce":        []string{Nonce},
			"CurTime":      []string{CurTime},
			"CheckSum":     []string{CheckSum},
		},
		DialTimeout: 10 * time.Second, //请求服务器超时
		ConnTimeout: 15 * time.Second, //下载超时
		TryTimes:    2,                //最大下载次数
		RetryPause:  0,                //重新下载前停顿时长
	})
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ret := map[string]interface{}{}
	err = json.Unmarshal(b, &ret)
	return ret, err
}
