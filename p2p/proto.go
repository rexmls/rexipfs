package p2p

type SwapStatsStruct struct {
	ListenPort int
	Peers []string `json:",omitempty"`
	Wanted []string `json:",omitempty"`
	Haves []string `json:",omitempty"`
}

type SwapContentStruct struct {
	Send []string `json:",omitempty"`
	Content []SwapContentItemStruct `json:",omitempty"`
}

type SwapContentItemStruct struct {
	Hash string
	Content_b64 string
}