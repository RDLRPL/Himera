package browser

type RenderState struct {
	NeedsRedraw   bool
	LastWidth     int
	LastHeight    int
	LastZoom      float32
	LastScroll    float32
	LastInputText string
	LastFocused   bool
	LastCursorPos int
}

type Browser struct {
	CurrentWidth    int
	CurrentHeight   int
	Link            string
	Ua              string
	InputText       string
	CursorPosition  int
	InputBoxHeight  float32
	InputBoxFocused bool
	BlinkTimer      float32
	RState          *RenderState
}

func NewBrowser(Width int, Height int, WelcomeLink string, SUa string, IBoxHeight float32) *Browser {
	return &Browser{
		CurrentWidth:  Width,
		CurrentHeight: Height,
		Link:          WelcomeLink,
		Ua:            SUa,

		// Initially const, then mut!!!
		InputBoxFocused: false,
		BlinkTimer:      0.0,
		RState:          &RenderState{NeedsRedraw: true},
		InputText:       WelcomeLink,
		CursorPosition:  len(WelcomeLink),
		InputBoxHeight:  IBoxHeight,
	}
}
