package strategy

func GetStrategy(strategy string) Strategy {
	switch strategy {
	case "default":
		return Variation4
	default:
		return nil
	}
}
