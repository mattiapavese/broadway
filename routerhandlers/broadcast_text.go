package routerhandlers

type BroadcastOnChannelTextBody struct {
	Channkey     string `json:"channkey"`
	HTTPCallback string `json:"http-callback"`
	Message      string `json:"message"`
}
