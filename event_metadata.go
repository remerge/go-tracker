package tracker

// EventMetadata represents fields in the Event that are set automatically by
// the tracker for all processed events.
type EventMetadata struct {
	Service     string
	Environment string
	Cluster     string
	Host        string
	Release     string
}

// MetadatableSimple so we can have implementors that don't depend on go-tracker
type MetadatableSimple interface {
	SetMetadata(string, string, string, string, string)
}

type Metadatable interface {
	SetMetadata(*EventMetadata)
}
