package browser

type Browser struct {
	CurrentWidth  int
	CurrentHeight int
}

func NewBrowser(Width int, Height int) *Browser {
	return &Browser{
		CurrentWidth:  Width,
		CurrentHeight: Height,
	}
}
