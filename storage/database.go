package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

type Page struct {
	ID        string
	Markdown  string
	HTML      string
	IsBurn    bool
	ExpiresAt *time.Time
	Password  string
	CreatedAt time.Time
}

var DB *sql.DB

// InitDB 对接挂载卷数据库缓存服务，建立并升级对应的核心高阶隐私数据列方案以支持动态表查询。
func InitDB(driver, dataSourceName string) {
	var err error
	DB, err = sql.Open(driver, dataSourceName)
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}

	if driver == "libsql" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := DB.PingContext(ctx); err != nil {
			log.Fatalf("数据库连通性校验失败: %v", err)
		}
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS pages (
		id TEXT PRIMARY KEY,
		markdown TEXT NOT NULL,
		html TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = DB.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("创建数据主干总表崩溃失败: %v", err)
	}

	// 利用 ALTER TABLE 执行在后续高阶方案发布系统更新上的数据横向拓展兼容写入，遇上冲突自抑制
	_, _ = DB.Exec("ALTER TABLE pages ADD COLUMN is_burn BOOLEAN DEFAULT FALSE;")
	_, _ = DB.Exec("ALTER TABLE pages ADD COLUMN expires_at DATETIME;")
	_, _ = DB.Exec("ALTER TABLE pages ADD COLUMN password TEXT;")

	// 触发定时清理回收线程
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			DB.Exec("DELETE FROM pages WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP")
		}
	}()
}

// SavePage 对参数进行兼容接收保护落成存储数据库并回吐双密匙组合。
func SavePage(markdown, html string, isBurn bool, expireDuration time.Duration, withPassword bool) (string, string, error) {
	var id string

	// 设置自动避免碰撞的最多次尝试安全冗余设定
	for i := 0; i < 5; i++ {
		newID, err := GenerateID(8)
		if err != nil {
			return "", "", err
		}
		var exists bool
		err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM pages WHERE id = ?)", newID).Scan(&exists)
		if err != nil {
			return "", "", err
		}
		if !exists {
			id = newID
			break
		}
	}

	if id == "" {
		return "", "", errors.New("无法成功分配空闲短链接访问 ID 极低可能性的报错暴露")
	}

	var nullPwd sql.NullString
	if withPassword {
		nullPwd.Valid = true
		nullPwd.String, _ = GeneratePassword(4)
	}

	var nullExpires sql.NullTime
	if expireDuration > 0 {
		nullExpires.Valid = true
		nullExpires.Time = time.Now().UTC().Add(expireDuration)
	}

	_, err := DB.Exec(`INSERT INTO pages (id, markdown, html, is_burn, expires_at, password) VALUES (?, ?, ?, ?, ?, ?)`,
		id, markdown, html, isBurn, nullExpires, nullPwd)

	return id, nullPwd.String, err
}

// GetPage 对查询单条记录提供并实施拦截、自动时间验证校验以维护保护措施的安全。
func GetPage(id string) (*Page, error) {
	var p Page
	var rawExpires sql.NullString
	var rawPassword sql.NullString
	var rawCreatedAt sql.NullString

	// 核武级数据抗性封装：利用原生 CAST(.. AS TEXT) 抵御一切因为 SQLite 无类型机制与 PureGo 驱动反序列化匹配失败引发的崩盘！
	err := DB.QueryRow("SELECT id, markdown, html, is_burn, CAST(expires_at AS TEXT), password, CAST(created_at AS TEXT) FROM pages WHERE id = ?", id).
		Scan(&p.ID, &p.Markdown, &p.HTML, &p.IsBurn, &rawExpires, &rawPassword, &rawCreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("寻找失效，系统拒绝呈现未声明查回数据包")
		}
		return nil, err
	}

	if rawPassword.Valid {
		p.Password = rawPassword.String
	}

	// 柔性转换提取出的纯字符串回真实过期刻度
	if rawExpires.Valid && rawExpires.String != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", rawExpires.String[:19]); err == nil {
			p.ExpiresAt = &t
		} else if t2, err2 := time.Parse(time.RFC3339, rawExpires.String); err2 == nil {
			p.ExpiresAt = &t2
		}
	}

	// 回收创建时间的固定坐标系
	if rawCreatedAt.Valid && rawCreatedAt.String != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", rawCreatedAt.String[:19]); err == nil {
			p.CreatedAt = t
		} else if t2, err2 := time.Parse(time.RFC3339, rawCreatedAt.String); err2 == nil {
			p.CreatedAt = t2
		} else {
			p.CreatedAt = time.Now()
		}
	} else {
		p.CreatedAt = time.Now()
	}

	// 核实动态到期的防守条件
	if p.ExpiresAt != nil && time.Now().UTC().After(*p.ExpiresAt) {
		DeletePage(id)
		return nil, errors.New("拦截触发：您访问的正文已经正式被执行超期摧毁拦截被抹杀！")
	}

	return &p, nil
}

// DeletePage 在针对诸如立即销毁状态或者定时巡检发生时暴露出独立的拦截执行操作
func DeletePage(id string) {
	DB.Exec("DELETE FROM pages WHERE id = ?", id)
}
