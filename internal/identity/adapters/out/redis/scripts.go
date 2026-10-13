package identityredis

import "github.com/redis/go-redis/v9"

// The scripts below assume a single, non-clustered Redis instance: the family
// key is derived at runtime from the family id stored in the session hash, so
// it cannot be declared in KEYS. The literal "identity:refresh:family:" prefix
// must stay in sync with refreshKeyPrefix in refresh_session_repository.go.
//
// rotateScript atomically consumes the session for the presented token and
// stores its replacement. Because Redis runs scripts serially, two concurrent
// rotations of the same token cannot both succeed.
const rotateScriptSource = `
local session = redis.call('HGETALL', KEYS[1])
if #session == 0 then
	if redis.call('EXISTS', KEYS[3]) == 1 then
		local family = redis.call('GET', KEYS[3])
		if family then
			redis.call('DEL', 'identity:refresh:family:' .. family)
		end
		return {'reused'}
	end
	return {'notfound'}
end

local data = {}
for i = 1, #session, 2 do
	data[session[i]] = session[i + 1]
end

local familyKey = 'identity:refresh:family:' .. data['family_id']
if redis.call('EXISTS', familyKey) == 0 then
	redis.call('DEL', KEYS[1])
	return {'revoked'}
end

redis.call('DEL', KEYS[1])
redis.call('SET', KEYS[3], data['family_id'], 'PX', ARGV[1])
redis.call('HSET', KEYS[2], 'user_id', data['user_id'], 'family_id', data['family_id'], 'expires_at', ARGV[2])
redis.call('PEXPIRE', KEYS[2], ARGV[1])
redis.call('PEXPIRE', familyKey, ARGV[1])

return {'ok', data['user_id'], data['family_id']}
`

// revokeScript deletes a session and its family, making logout idempotent.
const revokeScriptSource = `
local family = redis.call('HGET', KEYS[1], 'family_id')
redis.call('DEL', KEYS[1])
if family then
	redis.call('DEL', 'identity:refresh:family:' .. family)
end
return 1
`

var (
	rotateScript = redis.NewScript(rotateScriptSource)
	revokeScript = redis.NewScript(revokeScriptSource)
)
