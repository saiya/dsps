package unix

// UlimitRequirement describes required resource limit to run this server.
type UlimitRequirement struct {
	NoFiles int
}
