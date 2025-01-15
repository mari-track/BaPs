package code

import (
	"errors"
	"sync"
	"time"
)

var MaxFialNum = 5

type Code struct {
	codeMap  map[string]*CodeInfo // 验证码字典
	codeSync sync.RWMutex
}

type CodeInfo struct {
	Code    int32     // 验证码
	endTime time.Time // 到期时间
	FialNum int
}

func NewCode() *Code {
	return &Code{
		codeMap:  make(map[string]*CodeInfo),
		codeSync: sync.RWMutex{},
	}
}

// CheckCodeTime 定时检查一次是否有验证码过期
func (x *Code) CheckCodeTime() {
	ticker := time.NewTicker(time.Second * 300) // 五分钟验证一次
	for {
		<-ticker.C
		for account, codeInfo := range x.GetAllCode() {
			if time.Now().After(codeInfo.endTime) {
				x.DelCode(account)
			}
		}
	}
}

// GetCodeInfo 通过邮箱获取缓存的验证码
func (x *Code) GetCodeInfo(account string) *CodeInfo {
	if x == nil {
		return nil
	}
	x.codeSync.RLock()
	defer x.codeSync.RUnlock()
	code, ok := x.codeMap[account]
	if !ok {
		return nil
	}
	if !code.endTime.After(time.Now()) ||
		code.FialNum >= MaxFialNum {
		delete(x.codeMap, account)
		return nil
	}
	return code
}

// GetAllCode 获取全部已缓存的验证码
func (x *Code) GetAllCode() map[string]*CodeInfo {
	if x == nil {
		return nil
	}
	list := make(map[string]*CodeInfo)
	x.codeSync.RLock()
	defer x.codeSync.RUnlock()
	for k, v := range x.codeMap {
		list[k] = v
	}
	return list
}

// SetCode 设置邮箱的验证码 直接刷新
func (x *Code) SetCode(account string, code int32) error {
	if x == nil {
		return errors.New("Code is nil")
	}
	x.codeSync.Lock()
	defer x.codeSync.Unlock()
	x.codeMap[account] = &CodeInfo{
		Code:    code,
		endTime: time.Now().Add(30 * time.Minute), // 30分钟有效期
	}
	return nil
}

// DelCode 删除指定邮箱缓存的验证码
func (x *Code) DelCode(account string) {
	if x == nil {
		return
	}
	x.codeSync.Lock()
	defer x.codeSync.Unlock()
	if _, ok := x.codeMap[account]; ok {
		delete(x.codeMap, account)
	}
}
