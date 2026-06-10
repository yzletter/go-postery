local key = KEYS[1]                                 -- 商品 Key
local stock = tonumber(redis.call("GET", key))      -- 获取商品 Redis 库存

-- Key 不存在，或者库存不是数字
if stock == nil then
    return -1
end

-- 库存不足
if stock <= 0 then
    return 0
end

-- 扣减库存
redis.call("DECR", key)
return 1