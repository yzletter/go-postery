local key = KEYS[1]               -- Redis 存储验证码的 key
local code = ARGV[1]              -- 这次要发送的验证码
local interval = tonumber(ARGV[2]) -- 发送间隔时间（限制多久内不能重复发）
local expiration = tonumber(ARGV[3]) -- 验证码本身有效期
local ttl = tonumber(redis.call("ttl", key)) -- 获取当前 key 的剩余过期时间

if ttl == -1 then
    -- key 存在，但没有过期时间
    return -1
elseif ttl == -2 or ttl < expiration - interval then
    -- 未发过验证码 或 已经发了挺久超过了 interval
    redis.call("set", key, code)
    redis.call("expire", key, expiration)
    return 1
else
    -- 已经发过验证码，且还不到 interval
   return 0
end