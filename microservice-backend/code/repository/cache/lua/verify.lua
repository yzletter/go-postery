local key = KEYS[1]               -- Redis 存储验证码的 key
local code = ARGV[1]              -- 要校验的验证码
local target = redis.call("get", key) -- 获取当前 key 的剩余过期时间

if target == false then
    -- key 不存在
    return false
elseif target ~= code then
    -- Code 不相等
    return false
else
    -- Code 相等
    redis.call("del", key)
    return true
end