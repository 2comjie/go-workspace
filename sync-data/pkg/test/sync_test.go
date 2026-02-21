package test

import (
	"context"
	"fmt"
	"math/rand"
	"sync-data/pkg/db_impl"
	"sync-data/pkg/mark"
	"sync-data/pkg/redis_impl"
	"sync-data/pkg/sync"
	"sync-data/pkg/sync_def"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type User struct {
	ID       int64     `sync:"primary=1,cache=1" db:"id"`
	Name     string    `sync:"primary=2,cache=2" db:"name"`
	Age      int32     `db:"age"`
	Sex      string    `db:"sex"`
	Score    float64   `sync:"isScore=true" db:"score"`
	CreateAt time.Time `db:"created_at"`
	UpdateAt time.Time `db:"updated_at"`
}

func TestSync(t *testing.T) {
	// 连接 Redis
	rc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{"127.0.0.1:6379"},
	})
	defer rc.Close()

	// 连接 MySQL 数据库
	dsn := "root:123@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(20)

	// 清空测试数据
	ctx := context.Background()
	_, err = db.ExecContext(ctx, "TRUNCATE TABLE user")
	if err != nil {
		t.Fatalf("failed to truncate table: %v", err)
	}

	// 清空 Redis 测试数据
	err = rc.FlushDB(ctx).Err()
	if err != nil {
		t.Fatalf("failed to flush redis: %v", err)
	}

	redisHandler := redis_impl.NewBaseRedisSyncHandler[User](rc, sync_def.RedisOption[User]{
		DataRedisPrefix: "cache:",
		ExpireDuration:  -1,
	})

	dbHandler := db_impl.NewMysqlHandler[User](db, sync_def.DbOption[User]{
		TableName: "user",
	})

	marker := mark.NewSimpleMarker[User](rc, "dirty", sync_def.BaseOption[User]{})
	synchronizer := sync.NewSynchronizer[User](rc, redisHandler, dbHandler, marker, func() sync_def.FlushConfig {
		return sync_def.FlushConfig{
			FlushInterval: 5 * time.Second, // 5秒同步一次
			Batch:         100,
			ExpireTime:    -1,
		}
	}, sync_def.SyncOption[User]{
		BaseOption:         sync_def.BaseOption[User]{},
		RedisLockKeyPrefix: "lock:",
	})

	// 测试配置
	const (
		numUsers      = 1000 // 生成1000个用户
		numGoroutines = 50   // 50个并发协程
		numOperations = 2000 // 总共执行2000次操作
	)

	t.Logf("开始测试：生成 %d 个用户，使用 %d 个并发协程，执行 %d 次操作", numUsers, numGoroutines, numOperations)

	// 随机生成用户数据
	users := generateRandomUsers(numUsers)

	// 统计变量
	var saveCount atomic.Int64
	var loadCount atomic.Int64
	var deleteCount atomic.Int64
	var errorCount atomic.Int64

	startTime := time.Now()

	// 并发测试：写入数据
	t.Run("ConcurrentSave", func(t *testing.T) {
		doneChan := make(chan struct{})

		for i := 0; i < numGoroutines; i++ {
			go func(workerID int) {
				defer func() { doneChan <- struct{}{} }()

				for j := 0; j < numOperations/numGoroutines; j++ {
					// 随机选择一个用户进行保存
					userIdx := rand.Intn(numUsers)
					user := users[userIdx]
					user.UpdateAt = time.Now()

					err := synchronizer.SaveOne(&user, true)
					if err != nil {
						errorCount.Add(1)
						t.Logf("Worker %d: SaveOne error: %v", workerID, err)
					} else {
						saveCount.Add(1)
					}

					// 随机延迟
					time.Sleep(time.Millisecond * time.Duration(rand.Intn(5)))
				}
			}(i)
		}

		// 等待所有协程完成
		for i := 0; i < numGoroutines; i++ {
			<-doneChan
		}
	})

	saveElapsed := time.Since(startTime)
	t.Logf("保存操作完成：成功 %d 次，耗时 %v", saveCount.Load(), saveElapsed)

	// 并发测试：读取数据
	t.Run("ConcurrentLoad", func(t *testing.T) {
		doneChan := make(chan struct{})

		for i := 0; i < numGoroutines; i++ {
			go func(workerID int) {
				defer func() { doneChan <- struct{}{} }()

				for j := 0; j < 200; j++ {
					userIdx := rand.Intn(numUsers)
					user := users[userIdx]

					loadedUser, err := synchronizer.LoadOne(&User{ID: user.ID, Name: user.Name}, true)
					if err != nil {
						errorCount.Add(1)
						t.Logf("Worker %d: LoadOne error: %v", workerID, err)
					} else {
						loadCount.Add(1)
						if loadedUser != nil {
							// 验证数据正确性
							if loadedUser.ID != user.ID || loadedUser.Name != user.Name {
								t.Errorf("数据不一致: expected ID=%d Name=%s, got ID=%d Name=%s",
									user.ID, user.Name, loadedUser.ID, loadedUser.Name)
							}
						}
					}

					time.Sleep(time.Millisecond * time.Duration(rand.Intn(3)))
				}
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			<-doneChan
		}
	})

	loadElapsed := time.Since(startTime) - saveElapsed
	t.Logf("读取操作完成：成功 %d 次，耗时 %v", loadCount.Load(), loadElapsed)

	// 等待数据同步到数据库
	t.Log("等待数据从 Redis 同步到 MySQL...")
	time.Sleep(10 * time.Second)

	// 验证数据持久化
	t.Run("VerifyPersistence", func(t *testing.T) {
		var count int
		err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM user")
		if err != nil {
			t.Fatalf("查询数据库失败: %v", err)
		}

		t.Logf("数据库中共有 %d 条记录", count)

		if count == 0 {
			t.Error("数据库中没有数据，持久化失败")
		}

		// 随机抽查几条数据
		for i := 0; i < 10; i++ {
			userIdx := rand.Intn(numUsers)
			expectedUser := users[userIdx]

			var dbUser User
			err := db.GetContext(ctx, &dbUser, "SELECT * FROM user WHERE id = ? AND name = ?",
				expectedUser.ID, expectedUser.Name)
			if err != nil {
				t.Logf("警告：用户 ID=%d Name=%s 未找到: %v", expectedUser.ID, expectedUser.Name, err)
			} else {
				t.Logf("验证成功：ID=%d Name=%s Age=%d Sex=%s Score=%.2f",
					dbUser.ID, dbUser.Name, dbUser.Age, dbUser.Sex, dbUser.Score)
			}
		}
	})

	t.Run("ConcurrentDelete", func(t *testing.T) {
		doneChan := make(chan struct{})
		deleteUsers := users[:100] // 删除前100个用户

		for i := 0; i < 10; i++ {
			go func(workerID int) {
				defer func() { doneChan <- struct{}{} }()

				for j := 0; j < 10; j++ {
					user := deleteUsers[workerID*10+j]
					err := synchronizer.DelOne(&User{ID: user.ID, Name: user.Name}, true)
					if err != nil {
						errorCount.Add(1)
						t.Logf("Worker %d: DelOne error: %v", workerID, err)
					} else {
						deleteCount.Add(1)
					}
				}
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-doneChan
		}
	})

	t.Logf("删除操作完成：成功 %d 次", deleteCount.Load())

	// 等待删除同步
	time.Sleep(10 * time.Second)

	// 验证删除
	var finalCount int
	err = db.GetContext(ctx, &finalCount, "SELECT COUNT(*) FROM user")
	if err != nil {
		t.Fatalf("查询最终数据失败: %v", err)
	}

	totalElapsed := time.Since(startTime)
	t.Logf("\n========== 测试总结 ==========")
	t.Logf("总耗时: %v", totalElapsed)
	t.Logf("保存操作: %d 次", saveCount.Load())
	t.Logf("读取操作: %d 次", loadCount.Load())
	t.Logf("删除操作: %d 次", deleteCount.Load())
	t.Logf("错误次数: %d 次", errorCount.Load())
	//t.Logf("最终数据库记录数: %d 条", finalCount)
	t.Logf("==============================")
}

// generateRandomUsers 生成随机用户数据
func generateRandomUsers(count int) []User {
	rand.Seed(time.Now().UnixNano())

	firstNames := []string{"张", "李", "王", "刘", "陈", "杨", "赵", "黄", "周", "吴", "徐", "孙", "胡", "朱", "高", "林", "何", "郭", "马", "罗"}
	lastNames := []string{"伟", "芳", "娜", "秀英", "敏", "静", "丽", "强", "磊", "军", "洋", "勇", "艳", "杰", "涛", "明", "超", "娟", "波", "辉"}
	sexes := []string{"男", "女"}

	users := make([]User, count)
	for i := 0; i < count; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]

		users[i] = User{
			ID:       int64(i + 1),
			Name:     fmt.Sprintf("%s%s%d", firstName, lastName, i+1),
			Age:      int32(18 + rand.Intn(50)), // 18-67岁
			Sex:      sexes[rand.Intn(len(sexes))],
			Score:    float64(60+rand.Intn(41)) + rand.Float64(), // 60-100分
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
		}
	}

	return users
}
