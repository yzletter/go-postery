local gift_key = KEYS[1]
local record_key = KEYS[2]

local request_id = ARGV[1]
local user_id = ARGV[2]
local gift_id = ARGV[3]

-- 参数校验
if not request_id or request_id == "" or not user_id or user_id == "" or not gift_id or gift_id == "" then
    return {-1, request_id or "", gift_id or ""}
end

-- 返回值
-- 1 扣减成功
-- 2 复用记录
-- 0 库存不足
-- -1 非法数据

-- 复用旧预占记录
if redis.call("EXISTS", record_key) == 1 then
    local record = redis.call("HMGET", record_key, "request_id", "user_id", "gift_id")
    -- 记录的数据合法性检验
    if not record[1] or record[1] == "" or record[2] ~= user_id or not record[3] or record[3] == "" then
        return {-1, request_id, gift_id}
    end
    return {2, record[1], record[3]}
end

-- 尝试扣减
local stock = tonumber(redis.call("GET", gift_key))         -- 获取商品 Redis 库存

if stock == nil then                                        -- Key 不存在，或者库存不是数字
    return {-1, request_id, gift_id}
end

if stock <= 0 then                                          -- 库存不足
    return {0, request_id, gift_id}
end

redis.call("DECR", gift_key)                                -- 扣减库存

-- 写入新预占记录
redis.call("HMSET", record_key, "request_id", request_id, "user_id", user_id, "gift_id", gift_id)

return {1, request_id, gift_id}