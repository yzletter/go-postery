local gift_key = KEYS[1]
local rollback_key = KEYS[2]

-- 检查锁
if redis.call("EXISTS", rollback_key) == 1 then
    return 2
end

-- 校验库存合法性
local stock = redis.call("GET", gift_key)
if not stock or tonumber(stock) == nil then
    return -1
end

-- 回补库存
redis.call("INCR", gift_key)

-- 加锁
redis.call("SET", rollback_key, "1")

return 1
