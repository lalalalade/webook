local key = KEYS[1]

local expectedCode = ARGV[1]
local code = redis.call("get", key)
local cntKey = key..":cnt"
local cnt = tonumber(redis.call("get", cntKey))

if cnt <= 0 then
    --    说明 用户一直输错
    --    或者已经用过了
    return -1
elseif expectedCode == code then
    redis.call("set", cntKey, -1)
    return 0
else
    redis.call("decr", cntKey, -1)
    return -2
end