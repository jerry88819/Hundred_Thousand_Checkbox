package redis

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
    ctx      = context.Background()
    rdb      *redis.Client
    stateKey = "checkbox_state"
    ROWS     = 200
    COLS     = 500
)

func Init() *redis.Client {
    rdb = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379", // Redis 
        Password: "",               
        DB:       0,                
    })

    // 測試連線是否成功
    _, err := rdb.Ping(ctx).Result()
    if err != nil {
        panic(err)
    }

    return rdb
} // Init()

func SaveStateToRedis(index int, value bool) error {
    var bitValue int
    if value {
        bitValue = 1
    } else {
        bitValue = 0
    }
    return rdb.SetBit(ctx, stateKey, int64(index), bitValue).Err()
} // SaveStateToRedis()

func GetStateFromRedis() ([]bool, error) {
    ch := make(chan struct{}, 100 )
    var wg sync.WaitGroup
    state := make([]bool, ROWS*COLS)

    for i := 0; i < ROWS*COLS; i++ {
        wg.Add(1)
        ch <- struct{}{}
        go func(i int) {
            defer wg.Done()
            defer func() { <-ch }()
            bit, err := rdb.GetBit(ctx, stateKey, int64(i)).Result()
            if err != nil {
                return
            }
            state[i] = bit == 1
        }(i)
    } // for()

    wg.Wait()

    return state, nil
} // GetStateFromRedis()
