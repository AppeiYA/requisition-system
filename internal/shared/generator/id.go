package generator

type IDGenerator struct{}

func (g *IDGenerator) GenerateID(prefix string) (string, error) {
	gen, err := generateRandomString(8)
	if err != nil {
		return "", err
	}
	return prefix + "-" + gen, nil
}
