package conf

var (
	UserConf map[string]*User
)

type User struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Group    string `json:"group" binding:"required"`
}

func initUser(filename string) {
	userConf := make([]*User, 0)
	if err := readConfFile(filename, &userConf); err != nil {
		panic(err)
	}
	UserConf = make(map[string]*User, 0)
	for _, user := range userConf {
		UserConf[user.Username] = user
	}
}

func UserInGroup(user *User, group string) bool {
	switch group {
	case "user":
		return user.Group == "user" || user.Group == "admin"
	case "admin":
		return user.Group == "admin"
	default:
		return false
	}
}
