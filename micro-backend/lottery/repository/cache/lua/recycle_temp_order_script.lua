local temp_key = KEYS[1]
local gift_key = KEYS[2]
local temp_order_id = ARGV[1]
local gift_id = ARGV[2]

-- 读取当前用户的临时订单。
-- 如果 key 不存在，说明订单已经被支付、放弃、过期删除，或被其他消费者处理过。
-- 这种情况不需要恢复库存，返回 0 让调用方 Ack 消息即可。
local body = redis.call("GET", temp_key)
if not body then
    return 0
end

-- 解析临时订单 JSON。
-- TempOrder 在 Go 中的 JSON tag 使用 string，因此 id/gift_id 可能以字符串形式存储。
-- 这里统一用 tostring 比较，避免 Lua 数字和字符串类型差异导致误判。
local ok, temp_order = pcall(cjson.decode, body)

-- 必须同时校验订单 ID 和礼物 ID：
-- 1. 订单 ID 不一致：说明这是旧延迟消息，不能删除当前用户的新临时订单。
-- 2. 礼物 ID 不一致：说明 Go 层读取到的 gift_id 与 Redis 当前值不一致，不能回补错误库存。
-- 这两种情况都返回 0，调用方可以 Ack 当前消息。
if not ok or tostring(temp_order["id"]) ~= temp_order_id or tostring(temp_order["gift_id"]) ~= gift_id then
    return 0
end

-- 检查库存 key。
-- 如果库存 key 丢失或值不是数字，继续执行 INCR 可能会创建错误库存或污染数据。
-- 返回 -1 交给调用方重试/告警，不删除临时订单。
local stock = redis.call("GET", gift_key)
if not stock or tonumber(stock) == nil then
    return -1
end

-- 原子回收：
-- 先恢复库存，再删除临时订单。
-- 因为整个脚本原子执行，其他客户端不会看到只执行了一半的状态。
redis.call("INCR", gift_key)
redis.call("DEL", temp_key)

-- 回收完成，调用方可以 Ack 延迟消息。
return 1
