package v1alpha1

func (a *APIKeyAuthResponse) CanAccessModel(inModel string) bool {
	if a == nil {
		return false
	}
	for _, m := range a.AllowModels {
		if m == "*" {
			return true
		}
		if inModel == m {
			return true
		}
	}
	return false
}
