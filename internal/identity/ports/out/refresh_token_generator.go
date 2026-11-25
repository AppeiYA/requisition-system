package identityout

// RefreshTokenGenerator is the historical name of the opaque token contract.
// It is kept as an alias so existing wiring and mocks keep compiling while new
// use cases depend on OpaqueTokenGenerator.
type RefreshTokenGenerator = OpaqueTokenGenerator
