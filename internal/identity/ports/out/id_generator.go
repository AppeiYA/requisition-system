package identityout

type IDGenerator interface {
	GenerateID(prefix string) (string, error)
}
