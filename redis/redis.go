package redis

import (
    "github.com/go-redis/redis/v8"
    "context"
)

var (
    ctx      = context.Background()
    rdb      *redis.Client
    stateKey = "stateKey"
    ROWS     = 100
    COLS     = 100
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
    state := make([]bool, ROWS*COLS)
    for i := 0; i < ROWS*COLS; i++ {
        bit, err := rdb.GetBit(ctx, stateKey, int64(i)).Result()
        if err != nil {
            return nil, err
        }
        state[i] = bit == 1
    }
    return state, nil
} // GetStateFromRedis()
