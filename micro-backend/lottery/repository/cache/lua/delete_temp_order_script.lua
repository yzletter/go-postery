local key = KEYS[1]
local temp_order_id = ARGV[1]

-- 获取 JSON 序列化后的结果
local body = redis.call("GET", key)
if not body then
    return 0
end

-- 获取 Redis 中临时订单 ID 进行比较
local ok, temp_order = pcall(cjson.decode, body)
if not ok or temp_order["id"] ~= temp_order_id then
    return 0
end

return redis.call("DEL", key)
