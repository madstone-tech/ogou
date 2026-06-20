package engine

// Reporter receives events during scenario execution.
type Reporter interface {
	OnScenarioStart(name string, vus int)
	OnPhaseStart(name string, vus int)
	OnVUStart(vuID int)
	OnStepStart(vuID int, stepName string)
	OnStepResult(vuID int, result *Result)
	OnVUEnd(vuID int)
	OnPhaseEnd(name string, results []*Result)
	OnScenarioEnd(results []*Result)
}
