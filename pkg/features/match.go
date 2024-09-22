package features

type SNI interface {
}

// DomainMatch 域名匹配
func DomainMatch(h string) string {
	hits := DomainAc.MatchThreadSafe([]byte(h))
	if hits == nil {
		return ""
	}
	if name, ok := DomainMap[hits[0]]; ok {
		//zap.L().Info("匹配到域名信息", zap.String("hostname", h), zap.String("name", name), zap.String("domain", DomainMap[hits[0]]))
		return name
	}
	return ""
}
