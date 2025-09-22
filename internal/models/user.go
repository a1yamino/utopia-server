package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Policies 定义了角色的权限集合。
type Policies map[string]interface{}

// Value - 实现 driver.Valuer 接口，以便 GORM 可以将 Policies 写入数据库。
func (p Policies) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan - 实现 sql.Scanner 接口，以便 GORM 可以从数据库读取 Policies。
func (p *Policies) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &p)
}

// Role 定义了用户角色及其权限。
type Role struct {
	ID       int      `json:"id" gorm:"primaryKey"`
	Name     string   `json:"name" gorm:"unique"`
	Policies Policies `json:"policies" gorm:"type:json"`
}

// User 代表一个系统用户。
type User struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"unique"`
	PasswordHash string    `json:"-"` // 密码哈希不应被序列化到 JSON 中
	RoleID       int       `json:"roleId"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}
