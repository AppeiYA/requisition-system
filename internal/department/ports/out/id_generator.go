package departmentout

type IDGenerator interface {
	GenerateID(prefix string) (string, error)
}
