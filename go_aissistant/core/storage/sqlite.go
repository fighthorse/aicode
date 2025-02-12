package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db        *sql.DB
	batchSize int
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Printf("打开数据库失败: %v", err)
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Printf("数据库连接失败: %v", err)
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 启用WAL模式提升性能
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Printf("设置WAL模式失败: %v", err)
		return nil, fmt.Errorf("设置WAL模式失败: %v", err)
	}

	// 创建表结构
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS chat_history (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        query TEXT NOT NULL,
        response TEXT NOT NULL,
        model TEXT NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_timestamp ON chat_history(timestamp);
    `

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Printf("创建表失败: %v", err)
		return nil, fmt.Errorf("创建表失败: %v", err)
	}

	return &SQLiteStorage{
		db:        db,
		batchSize: 50,
	}, nil
}

func (s *SQLiteStorage) SaveChatRecord(query, response, model string) error {
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("开始事务失败: %v", err)
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
        INSERT INTO chat_history(query, response, model) 
        VALUES(?, ?, ?)
    `)
	if err != nil {
		log.Printf("准备语句失败: %v", err)
		return fmt.Errorf("准备语句失败: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(query, response, model); err != nil {
		log.Printf("执行插入语句失败: %v", err)
		return fmt.Errorf("执行插入语句失败: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("提交事务失败: %v", err)
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

func (s *SQLiteStorage) GetRecentHistory(num int64) ([]map[string]string, error) {
	rows, err := s.db.Query(`
        SELECT id, query, response, model, timestamp
        FROM chat_history
        ORDER BY timestamp DESC
        LIMIT ?
    `, num)

	if err != nil {
		log.Printf("查询最近历史记录失败: %v", err)
		return nil, fmt.Errorf("查询最近历史记录失败: %v", err)
	}
	defer rows.Close()

	var history []map[string]string
	for rows.Next() {
		var id int
		var query, response, model, timestamp string
		if err := rows.Scan(&id, &query, &response, &model, &timestamp); err != nil {
			log.Printf("扫描行失败: %v", err)
			return nil, fmt.Errorf("扫描行失败: %v", err)
		}
		history = append(history, map[string]string{
			"id":        fmt.Sprintf("%d", id),
			"query":     query,
			"response":  response,
			"model":     model,
			"timestamp": timestamp,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("遍历行时出错: %v", err)
		return nil, fmt.Errorf("遍历行时出错: %v", err)
	}

	return history, nil
}

func (s *SQLiteStorage) GetHistoryByTimeRange(start, end time.Time) ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`
        SELECT id, query, response, model, timestamp 
        FROM chat_history 
        WHERE timestamp BETWEEN ? AND ?
        ORDER BY timestamp DESC
    `, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		log.Printf("查询时间范围内的历史记录失败: %v", err)
		return nil, fmt.Errorf("查询时间范围内的历史记录失败: %v", err)
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var id int
		var query, response, model, timestamp string
		if err := rows.Scan(&id, &query, &response, &model, &timestamp); err != nil {
			log.Printf("扫描行失败: %v", err)
			return nil, fmt.Errorf("扫描行失败: %v", err)
		}

		history = append(history, map[string]interface{}{
			"id":        id,
			"query":     query,
			"response":  response,
			"model":     model,
			"timestamp": timestamp,
		})
	}

	if err := rows.Err(); err != nil {
		log.Printf("遍历行时出错: %v", err)
		return nil, fmt.Errorf("遍历行时出错: %v", err)
	}

	return history, nil
}

func (s *SQLiteStorage) CleanOldRecords(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	_, err := s.db.Exec(`
        DELETE FROM chat_history 
        WHERE timestamp < ?
    `, cutoff.Format(time.RFC3339))
	if err != nil {
		log.Printf("清理旧记录失败: %v", err)
		return fmt.Errorf("清理旧记录失败: %v", err)
	}
	return nil
}

func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			log.Printf("关闭数据库失败: %v", err)
			return fmt.Errorf("关闭数据库失败: %v", err)
		}
	}
	return nil
}

func (s *SQLiteStorage) ClearHistory() error {
	_, err := s.db.Exec(`DELETE FROM chat_history`)
	if err != nil {
		log.Printf("清除历史记录失败: %v", err)
		return fmt.Errorf("清除历史记录失败: %v", err)
	}
	return nil
}

func (s *SQLiteStorage) DeleteEntry(id int) error {
	_, err := s.db.Exec(`DELETE FROM chat_history WHERE id = ?`, id)
	if err != nil {
		log.Printf("删除条目失败: %v", err)
		return fmt.Errorf("删除条目失败: %v", err)
	}
	return nil
}
