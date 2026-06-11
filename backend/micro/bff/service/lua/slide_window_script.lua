local key = KEYS[1]
local duration = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])

local now = tonumber(ARGV[3])
local member = ARGV[4]
local startTime = now - duration

redis.call('ZREMRANGEBYSCORE', key, '-inf', startTime)
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

if cnt >= threshold then
    return "true"
else
    redis.call('ZADD', key, now, member)
    redis.call('PEXPIRE', key, duration)
    return "false"
end
