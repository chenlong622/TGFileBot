package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	handleUrl "net/url"
	"strconv"
)

// buildIDs 根据配置重建 IDs 以支持 O(1) 权限查询
func (infos *Infos) buildIDs() {
	conf := infos.Conf.Load()

	infos.Mutex.Lock()
	defer infos.Mutex.Unlock()
	// 检查UserID是否在IDs中
	if value, ok := infos.IDs[conf.UserID]; !ok {
		value.IsAdmin = true
		value.IsWhite = true
		infos.IDs[conf.UserID] = value
	}

	// 检查AdminIDs是否在IDs中
	for _, id := range conf.AdminIDs {
		if value, ok := infos.IDs[id]; !ok {
			value.IsAdmin = true
			value.IsWhite = true
			infos.IDs[id] = value
		}
	}

	// 检查WhiteIDs是否在IDs中
	for _, id := range conf.WhiteIDs {
		if value, ok := infos.IDs[id]; !ok {
			value.IsWhite = true
			infos.IDs[id] = value
		}
	}

	infos.rebuildHashIndex()
}

// rebuildHashIndex 根据当前 IDs 重建 hash -> uid 反查表, 调用方必须已持有 infos.Mutex 写锁
func (infos *Infos) rebuildHashIndex() {
	index := make(map[string]int64, len(infos.IDs))
	for uid := range infos.IDs {
		if hash := infos.calculateHash(uid); hash != "" {
			index[hash] = uid
		}
	}
	infos.HashIndex = index
}

func (infos *Infos) isAdmin(id int64) bool {
	infos.Mutex.RLock()
	defer infos.Mutex.RUnlock()
	if value, ok := infos.IDs[id]; ok {
		return value.IsAdmin
	}
	return false
}

func (infos *Infos) isWhite(id int64) bool {
	infos.Mutex.RLock()
	defer infos.Mutex.RUnlock()
	if value, ok := infos.IDs[id]; ok {
		return value.IsWhite
	}
	return false
}

// calculateHash 为指定用户 ID 生成 6 位 MD5 哈希, 用于鉴权
func (infos *Infos) calculateHash(userID int64) string {
	password := infos.Conf.Load().Password
	if password == "" {
		return ""
	}
	res := fmt.Sprintf("%d%s", userID, password)
	src := md5.Sum([]byte(res))
	return hex.EncodeToString(src[:])[:6]
}

// checkHash 根据哈希值查找对应的用户 ID, 返回 0 表示未找到
// 通过 HashIndex 反查表 O(1) 查找, 而非每次都对 IDs 做线性扫描
func (infos *Infos) checkHash(hash string) int64 {
	if hash == "" {
		return 0
	}
	infos.Mutex.RLock()
	defer infos.Mutex.RUnlock()
	return infos.HashIndex[hash]
}

// checkPass 验证 HTTP 请求中的访问密码或哈希
func checkPass(params handleUrl.Values) error {
	confPassword := infos.Conf.Load().Password
	if confPassword != "" {
		hash := params.Get("hash") // 基于用户 ID 的哈希校验
		password := params.Get("key")
		switch {
		case password != "":
			if password != confPassword {
				return errors.New("无效的密码")
			}
		case hash != "":
			value := params.Get("uid")
			uid, err := strconv.ParseInt(value, 10, 64)
			if err == nil && uid != 0 {
				if hash != infos.calculateHash(uid) {
					return errors.New("无效的哈希密码")
				}
			} else {
				log.Printf("UID无效: %s", value)
				return errors.New("无效的UID")
			}
		default:
			return errors.New("您没有权限访问此链接")
		}
	}
	return nil
}
