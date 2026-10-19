package invitationout

type IDGenerator interface {
	GenerateID(prefix string) (string, error)
}
