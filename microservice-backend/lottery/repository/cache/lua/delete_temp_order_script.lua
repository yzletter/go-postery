local key = KEYS[1]
local temp_order_id = ARGV[1]

local value = redis.call("GET", key)
if not value then
    return 0
end

local ok, temp_order = pcall(cjson.decode, value)
if not ok or temp_order["id"] ~= temp_order_id then
    return 0
end

return redis.call("DEL", key)
