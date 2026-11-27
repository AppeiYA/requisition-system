package attachmentout

type IDGenerator interface {
	GenerateID(prefix string) (string, error)
}
