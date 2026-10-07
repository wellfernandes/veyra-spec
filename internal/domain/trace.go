package domain

// TraceCompleteness indicates whether a Trace contains a partial or complete execution.
type TraceCompleteness uint8

const (
	// TraceCompletenessUnknown means the source cannot determine whether observation is complete.
	TraceCompletenessUnknown TraceCompleteness = iota
	// TraceCompletenessPartial means the Trace is a known prefix of an execution.
	TraceCompletenessPartial
	// TraceCompletenessComplete means the Trace contains the complete execution being validated.
	TraceCompletenessComplete
)

type Trace struct {
	ObservedEvents []ObservedEvent
	Completeness   TraceCompleteness
}
