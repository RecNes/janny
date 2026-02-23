package components

// statusMsg is used to update the footer status
type StatusMsg string

// dashboardMsg is used to update the dashboard's output box
type DashboardMsg string

// browseMsg tells the main model to switch to FileBrowser for path selection
type BrowseMsg struct {
	FieldIndex int
}

// selectPathMsg is returned by FileBrowser when a path is picked
type SelectPathMsg struct {
	Path       string
	FieldIndex int
}
