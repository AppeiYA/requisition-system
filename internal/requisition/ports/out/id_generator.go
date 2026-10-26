package requisitionout

type IDGenerator interface {
	GenerateID(prefix string) (string, error)
}
