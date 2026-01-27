local key = KEYS[1] -- key
local field = ARGV[1]   -- 字段
local delta = tonumber(ARGV[2]) -- 修改值
local exits = redis.call("EXISTS", key) -- 当前 key 是否存在

if exits == 1 then
    redis.call("HINCRBY", key, field, delta)
    return 1    -- 改变成功
else
    return 0    -- 改变失败
end

