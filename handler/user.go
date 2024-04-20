package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	dblayer "github.com/bladewaltz9/FileStoreServer/db"
	"github.com/bladewaltz9/FileStoreServer/util"
)

const (
	PWD_SALT               = "!@#432"
	TOKEN_SALT             = "!@#%kdskla"
	TOKEN_EXPIRATION_HOURS = 24
)

// SignupHandler: handle user sign up request
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// return the signup html page
		data, err := os.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		username := r.Form.Get("username")
		passwd := r.Form.Get("password")

		if len(username) < 3 || len(passwd) < 5 {
			w.Write([]byte("Invalid parameter"))
			return
		}
		// encrypt the  password by Sha1
		encPasswd := util.Sha1([]byte(passwd + PWD_SALT))

		suc := dblayer.UserSignup(username, encPasswd)
		if suc {
			w.Write([]byte("SUCCESS"))
		} else {
			w.Write([]byte("FAILED"))
		}
	}
}

// SignHandler: handle user sign in request
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// return the signin html page
		data, err := os.ReadFile("./static/view/signin.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	encPasswd := util.Sha1([]byte(passwd + PWD_SALT))
	// check username and password
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	// generate token
	token := GenToken(username)
	suc := dblayer.UpdateToken(username, token)
	if !suc {
		w.Write([]byte("FAILED"))
		return
	}

	// redirect to home page after successfuy login
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// parse request parameters
	r.ParseForm()
	username := r.Form.Get("username")
	// token := r.Form.Get("token")

	// // check if token is valid
	// isTokenValid := IsTokenValid(username, token)
	// if !isTokenValid {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	return
	// }

	// query user info
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// assemble data and respond to user
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

// GenToken: generate token by username, timestamp and TOKEN_SALT
func GenToken(username string) string {
	// 40 char: md5(username + timestamp + TOKEN_SALT) + timestamp[:8]
	timestamp := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + timestamp + TOKEN_SALT))
	return tokenPrefix + timestamp[:8]
}

// IsTokenValid: check if token is valid
func IsTokenValid(username string, token string) bool {
	if len(token) != 40 {
		return false
	}
	// check if token has expired
	createTimestamp, _ := strconv.ParseInt(token[len(token)-8:], 16, 64)
	createTime := time.Unix(createTimestamp, 0)
	duration := time.Since(createTime)
	durationHours := int(duration.Hours())

	if durationHours >= TOKEN_EXPIRATION_HOURS {
		return false
	}

	// check if the token is same as the token in the database
	tokenDB, err := dblayer.GetToken(username)
	if err != nil || token != tokenDB {
		return false
	}

	return true
}
