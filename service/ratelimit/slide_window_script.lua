local key = KEYS[1]                        -- 限流对象
local duration = tonumber(ARGV[1])         -- 窗口大小
local threshold = tonumber(ARGV[2])        -- 阈值

local now = tonumber(ARGV[3])              -- 当前时间
local startTime = now - duration           -- 起始时间

redis.call('ZREMRANGEBYSCORE', key, '-inf', startTime)      -- 移除 startTime 之前所有值
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')       -- 统计当前 IP 请求了多少次

if cnt >= threshold then                   -- 当前 IP 有效请求次数 >= 阈值, 执行限流
    return "true"
else                                       -- 当前 IP 有效请求次数 < 阈值, 不执行限流
    redis.call('ZADD', key, now, now)      -- 把 score 和 value 都设置成 now 添加进 set 中
    redis.call('PEXPIRE', key, duration)   -- 设置过期时间
    return "false"
end
