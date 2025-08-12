package browser

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
}

func NewBrowser(Width int, Height int, WelcomeLink string, SUa string, IBoxHeight float32) *Browser {
	return &Browser{
		CurrentWidth:   Width,
		CurrentHeight:  Height,
		Link:           WelcomeLink,
		Ua:             SUa,
		InputText:      WelcomeLink,
		CursorPosition: len(WelcomeLink),

		InputBoxHeight: IBoxHeight,

		// always false!!!
		InputBoxFocused: false,
		BlinkTimer:      0.0,
	}
}
