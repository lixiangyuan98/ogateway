package conf

var (
	VideoConf []*VideoServer
)

type VideoServer struct {
	Host  string `json:"host" binding:"required"`
	Group string `json:"group" binding:"required"`
}

func initVideo(filename string) {
	VideoConf = make([]*VideoServer, 0)
	if err := readConfFile(filename, &VideoConf); err != nil {
		panic(err)
	}
}
