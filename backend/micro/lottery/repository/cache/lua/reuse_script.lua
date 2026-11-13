local record_key = KEYS[1]

local request_id = ARGV[1]
local user_id = ARGV[2]

-- 参数校验
if not request_id or request_id == "" or not user_id or user_id == "" then
    return {-1, request_id or "", ""}
end

-- 返回值
-- 0 没有记录
-- 1 复用记录
-- -1 非法数据

-- 复用旧预占记录
if redis.call("EXISTS", record_key) == 1 then
    local record = redis.call("HMGET", record_key, "request_id", "user_id", "gift_id")
    -- 记录的数据合法性检验
    if not record[1] or record[1] == "" or record[2] ~= user_id or not record[3] or record[3] == "" then
        return {-1, record[1], record[3]}
    end
    return {1, record[1], record[3]}
end

-- 没有记录复用
return {0, "", ""}
