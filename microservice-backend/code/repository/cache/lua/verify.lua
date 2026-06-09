local key = KEYS[1]
local code = ARGV[1]

local target = redis.call("get", key)

if not target then
    -- key 不存在
    return 0
end

if target ~= code then
    -- 验证码错误
    return 1
end

-- 验证码正确，删除 key，防止重复使用
redis.call("del", key)
return 2