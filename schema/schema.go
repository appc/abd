package schema

type ABDMetadataFormat struct {
	Name    string            `json:"string"`
	Labels  map[string]string `json:"labels"`
	Mirrors []string          `json:"mirrors"`
}

type ABDMetadataFetchStrategyConfiguration struct {
	Prefix   string `json:"prefix"`
	Strategy string `json:"strategy"`
}
