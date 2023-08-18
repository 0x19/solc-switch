package solc

// VersionInfo represents a simplified structure containing only the version tag name and an indication if it's the latest/prerelease version.
type VersionInfo struct {
	TagName      string `json:"tag_name"`
	IsLatest     bool   `json:"is_latest"`
	IsPrerelease bool   `json:"is_prerelease"`
}

// Version represents the structure of a Solidity version.
// It contains details about a specific release version of Solidity.
type Version struct {
	// URL is the API URL for this release.
	URL string `json:"url"`
	// AssetsURL is the URL to fetch the assets for this release.
	AssetsURL string `json:"assets_url"`
	// UploadURL is the URL to upload assets for this release.
	UploadURL string `json:"upload_url"`
	// HTMLURL is the web URL for this release.
	HTMLURL string `json:"html_url"`
	// ID is the unique identifier for this release.
	ID int `json:"id"`
	// NodeID is the unique node identifier for this release.
	NodeID string `json:"node_id"`
	// TagName is the git tag associated with this release.
	TagName string `json:"tag_name"`
	// TargetCommitish is the commit this release is associated with.
	TargetCommitish string `json:"target_commitish"`
	// Name is the name of the release.
	Name string `json:"name"`
	// Draft indicates if this release is a draft.
	Draft bool `json:"draft"`
	// Prerelease indicates if this release is a pre-release.
	Prerelease bool `json:"prerelease"`
	// CreatedAt is the timestamp when this release was created.
	CreatedAt string `json:"created_at"`
	// PublishedAt is the timestamp when this release was published.
	PublishedAt string `json:"published_at"`
	// Assets is a list of assets associated with this release.
	Assets []Asset `json:"assets"`
	// TarballURL is the URL to download the tarball archive of this release.
	TarballURL string `json:"tarball_url"`
	// ZipballURL is the URL to download the zip archive of this release.
	ZipballURL string `json:"zipball_url"`
	// Body is the release notes for this release.
	Body string `json:"body"`
	// Reactions contains the reactions for this release.
	Reactions Reactions `json:"reactions"`
	// Author is the user who published this release.
	Author Author `json:"author"`
}

// GetVersionInfo returns a VersionInfo struct containing the version's tag name and an indication if it's the latest version.
func (v *Version) GetVersionInfo(latestVersionTag string) VersionInfo {
	return VersionInfo{
		TagName:      v.TagName,
		IsLatest:     v.TagName == latestVersionTag,
		IsPrerelease: v.Prerelease,
	}
}

// Asset represents a downloadable asset associated with a release.
type Asset struct {
	// URL is the API URL for this asset.
	URL string `json:"url"`
	// ID is the unique identifier for this asset.
	ID int `json:"id"`
	// NodeID is the unique node identifier for this asset.
	NodeID string `json:"node_id"`
	// Name is the name of the asset.
	Name string `json:"name"`
	// Label is an optional label for the asset.
	Label string `json:"label"`
	// Uploader is the user who uploaded this asset.
	Uploader Author `json:"uploader"`
	// ContentType is the MIME type of the asset.
	ContentType string `json:"content_type"`
	// State is the state of the asset (e.g., "uploaded").
	State string `json:"state"`
	// Size is the size of the asset in bytes.
	Size int `json:"size"`
	// DownloadCount is the number of times this asset has been downloaded.
	DownloadCount int `json:"download_count"`
	// CreatedAt is the timestamp when this asset was created.
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the timestamp when this asset was last updated.
	UpdatedAt string `json:"updated_at"`
	// BrowserDownloadURL is the URL to download the asset.
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Author represents the user who published a release or uploaded an asset.
type Author struct {
	// Login is the username of the author.
	Login string `json:"login"`
	// ID is the unique identifier for the author.
	ID int `json:"id"`
	// NodeID is the unique node identifier for the author.
	NodeID string `json:"node_id"`
	// AvatarURL is the URL to the author's avatar.
	AvatarURL string `json:"avatar_url"`
	// URL is the API URL for the author.
	URL string `json:"url"`
	// HTMLURL is the web URL for the author's profile.
	HTMLURL string `json:"html_url"`
	// FollowersURL is the URL to fetch the author's followers.
	FollowersURL string `json:"followers_url"`
	// FollowingURL is the URL to see who the author is following.
	FollowingURL string `json:"following_url"`
	// GistsURL is the URL to see the author's gists.
	GistsURL string `json:"gists_url"`
	// StarredURL is the URL to see what repositories the author has starred.
	StarredURL string `json:"starred_url"`
	// SubscriptionsURL is the URL to see the author's subscriptions.
	SubscriptionsURL string `json:"subscriptions_url"`
	// OrganizationsURL is the URL to see the organizations the author belongs to.
	OrganizationsURL string `json:"organizations_url"`
	// ReposURL is the URL to see the author's repositories.
	ReposURL string `json:"repos_url"`
	// EventsURL is the URL to see the author's events.
	EventsURL string `json:"events_url"`
	// ReceivedEventsURL is the URL to see events received by the author.
	ReceivedEventsURL string `json:"received_events_url"`
	// Type indicates the type of the user (e.g., "User" or "Organization").
	Type string `json:"type"`
	// SiteAdmin indicates if the author is a site administrator.
	SiteAdmin bool `json:"site_admin"`
}

// Reactions represents the reactions to a release.
type Reactions struct {
	// URL is the API URL for these reactions.
	URL string `json:"url"`
	// TotalCount is the total number of reactions.
	TotalCount int `json:"total_count"`
	// PlusOne is the number of "+1" reactions.
	PlusOne int `json:"+1"`
	// MinusOne is the number of "-1" reactions.
	MinusOne int `json:"-1"`
	// Laugh is the number of "laugh" reactions.
	Laugh int `json:"laugh"`
	// Hooray is the number of "hooray" reactions.
	Hooray int `json:"hooray"`
	// Confused is the number of "confused" reactions.
	Confused int `json:"confused"`
	// Heart is the number of "heart" reactions.
	Heart int `json:"heart"`
	// Rocket is the number of "rocket" reactions.
	Rocket int `json:"rocket"`
	// Eyes is the number of "eyes" reactions.
	Eyes int `json:"eyes"`
}
