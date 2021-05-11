package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	prefix = "http://192.168.2.1"
	user   = "root"
	pass   = ""
)

func main() {
	log.Println("logging in")
	c := login()
	log.Println("checking enabled")
	enabled := getEnabled(c)
	for i := 3; i > -1; i-- {
		if enabled {
			log.Println("already enabled, disable in", i, "seconds")
		} else {
			log.Println("already disabled, enable in", i, "seconds")
		}
		time.Sleep(1 * time.Second)
	}
	if enabled {
		log.Println("disabling")
		switchV2ray(c, false)
	} else {
		log.Println("enabling")
		switchV2ray(c, true)
	}
	enabled = getEnabled(c)
	if enabled {
		log.Println("current status is enabled")
	} else {
		log.Println("current status is disabled")
	}
}

func login() *http.Cookie {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	form := url.Values{}
	form.Add("luci_username", user)
	form.Add("luci_password", pass)
	req, err := http.NewRequest("POST", prefix+"/cgi-bin/luci/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	for _, c := range resp.Cookies() {
		if c.Name == "sysauth" {
			return c
		}
	}
	log.Fatal("login failed")
	return nil
}

func getEnabled(cookie *http.Cookie) bool {
	req, err := http.NewRequest("GET", prefix+"/_api/v2ray", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	var r struct {
		Result []map[string]string `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	resp.Body.Close()
	return r.Result[0]["v2ray_basic_enable"] == "1"
}

func switchV2ray(cookie *http.Cookie, on bool) {
	enable := "0"
	if on {
		enable = "1"
	}
	data := `{"method":"v2ray_config.sh","params":[1],"fields":{"v2ray_basic_enable":"` + enable + `"}}`
	req, err := http.NewRequest("POST", prefix+"/_api/", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
}
