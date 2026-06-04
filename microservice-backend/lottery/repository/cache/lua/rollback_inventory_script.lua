local gift_key = KEYS[1]
local rollback_key = KEYS[2]
local marker_ttl = tonumber(ARGV[1])

if redis.call("EXISTS", rollback_key) == 1 then
    return 2
end

local stock = redis.call("GET", gift_key)
if not stock or tonumber(stock) == nil then
    return -1
end

redis.call("INCR", gift_key)
redis.call("SET", rollback_key, "1", "EX", marker_ttl)

return 1
