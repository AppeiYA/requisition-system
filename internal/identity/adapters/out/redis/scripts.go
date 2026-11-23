package identityredis

import "github.com/redis/go-redis/v9"

// The scripts below assume a single, non-clustered Redis instance: the family
// and epoch keys are derived at runtime from data stored in the session hash,
// so they cannot be declared in KEYS. The literal "identity:refresh:family:"
// and "identity:refresh:epoch:" prefixes must stay in sync with
// refreshKeyPrefix in refresh_session_repository.go.
//
// createScript snapshots the user's current epoch into the new session. Doing
// it in Lua makes the read-then-write atomic with respect to RevokeAll, so a
// session can never be created with an epoch that was already revoked.
const createScriptSource = `
local epoch = redis.call('GET', KEYS[3])
if not epoch then
	epoch = '0'
end

redis.call('HSET', KEYS[1],
	'user_id', ARGV[1],
	'family_id', ARGV[2],
	'epoch', epoch,
	'expires_at', ARGV[4])
redis.call('PEXPIRE', KEYS[1], ARGV[3])
redis.call('SET', KEYS[2], ARGV[1], 'PX', ARGV[3])

return 1
`

// rotateScript atomically consumes the session for the presented token and
// stores its replacement. Because Redis runs scripts serially, two concurrent
// rotations of the same token cannot both succeed. It also rejects a session
// whose epoch is behind the user's current epoch, which is how RevokeAll
// invalidates every existing refresh token at once.
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

local currentEpoch = redis.call('GET', 'identity:refresh:epoch:' .. data['user_id'])
if not currentEpoch then
	currentEpoch = '0'
end
local storedEpoch = data['epoch']
if not storedEpoch then
	storedEpoch = '0'
end
if storedEpoch ~= currentEpoch then
	redis.call('DEL', KEYS[1])
	return {'revoked'}
end

local familyKey = 'identity:refresh:family:' .. data['family_id']
if redis.call('EXISTS', familyKey) == 0 then
	redis.call('DEL', KEYS[1])
	return {'revoked'}
end

redis.call('DEL', KEYS[1])
redis.call('SET', KEYS[3], data['family_id'], 'PX', ARGV[1])
redis.call('HSET', KEYS[2],
	'user_id', data['user_id'],
	'family_id', data['family_id'],
	'epoch', currentEpoch,
	'expires_at', ARGV[2])
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
	createScript = redis.NewScript(createScriptSource)
	rotateScript = redis.NewScript(rotateScriptSource)
	revokeScript = redis.NewScript(revokeScriptSource)
)
