package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"crypto/rand"
	"strconv"
	"strings"
	"net/url"
	"runtime"
	"os/exec"

	"github.com/tidwall/gjson"
)

type MediaData struct {
	is_video bool
	url string
}

func main() {
	var  (
		username string
		password string
		session string
		thread_id string
		option string
		err error
	)
	fmt.Print("[?] Would you like to use username:password or a session id? [1/2]: ")
	fmt.Scanln(&option)

	if option == "1" {
		fmt.Print("[?] Username: ")
		fmt.Scanln(&username)

		fmt.Print("[?] Password: ")
		fmt.Scanln(&password)

		session, err = login(username, password)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("[!] Logged in successfully!")
		time.Sleep(time.Second * 3)

	} else {
		fmt.Print("[?] Session: ")
		fmt.Scanln(&session)
	}
	for {
		fmt.Print("[?] Thread ID: ")
		fmt.Scanln(&thread_id)
		raven_links := get_raven_links(session, thread_id)
		if len(raven_links) != 0 {
			download_raven_links(raven_links)
			fmt.Println("[!] Finished downloading all disappearing media...")
		} else  {
			fmt.Println("[!] Couldn't find any disappearing links in this channel :(")
		}
		time.Sleep(time.Second * 5)

		fmt.Print("[?] Would you like to go again? [y/n]: ")
		fmt.Scanln(&option)
		if option != "y" || option != "Y" {
			break
		}
		clear_console()

	}
}

func exists(name string) bool {
    _, err := os.Stat(name)
    return !os.IsNotExist(err)
}

func generate_filename(id string, media_data MediaData) string {
	var filename string
	if media_data.is_video {
		filename = fmt.Sprintf("%s.mp4", id)
	} else {
		filename = fmt.Sprintf("%s.jpg", id)
	}
	return filename

}

func download_raven_links(raven_links map[string]MediaData) {
    for id, media_data := range raven_links {
	filename := generate_filename(id, media_data)
	if !exists(filename) && media_data.url != "" {
		resp, err := http.Get(media_data.url)
		if err != nil  {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		_, err = file.Write(body)
		if err != nil {
			log.Fatal(err)
		}
	}
    }
}

func get_raven_links(session string, thread_id string) map[string]MediaData {
	client := &http.Client{}
	raven_links := make(map[string]MediaData)
	request, err := http.NewRequest("GET", fmt.Sprintf("https://i.instagram.com/api/v1/direct_v2/threads/%s", thread_id), nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Add("user-agent", "Instagram 85.0.0.21.100 Android (28/9; 380dpi; 1080x2147; OnePlus; HWEVA; OnePlus6T; qcom; en_US; 146536611)")
	request.Header.Add("cookie", fmt.Sprintf("sessionid=%s", session))
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	json_buffer := string(body)
	if resp.StatusCode != 200 {
		return raven_links
	}
	for i := 0; i < 19; i++ {
		var media_data MediaData
		item_type := gjson.Get(json_buffer, fmt.Sprintf("thread.items.%d.item_type", i)).String()
		if item_type == "raven_media" {
			item_id := gjson.Get(json_buffer, fmt.Sprintf("thread.items.%d.item_id", i)).String()
			media_type := gjson.Get(json_buffer, fmt.Sprintf("thread.items.%d.visual_media.media.media_type", i)).String()

			if media_type == "1" {
				media_data.is_video = false
				temp := gjson.Get(json_buffer, fmt.Sprintf("thread.items.%d.visual_media.media.image_versions2.candidates.0.url", i))
				if temp.Exists() {
					media_data.url = temp.String()
				} else {
					media_data.url = ""
				}
			} else {
				media_data.is_video = true
				temp := gjson.Get(json_buffer, fmt.Sprintf("thread.items.%d.visual_media.media.video_versions.0.url", i))
				if temp.Exists() {
					media_data.url = temp.String()
				} else {
					media_data.url = ""
				}

			}
			raven_links[item_id] = media_data
		}
	}
	return raven_links
}

func login(username string, password string) (string, error) {
	client := &http.Client{}
	link := "https://i.instagram.com/api/v1/accounts/login/"
	data := url.Values{}
	data.Add("guid", generate_uuid())
	data.Add("enc_password", fmt.Sprintf("#PWD_INSTAGRAM:0:%s:%s", strconv.Itoa(int(time.Now().Unix())), password))
	data.Add("username", username)
	data.Add("device_id", "android-JDS095823049")
	data.Add("login_attempt_count", "0")
	req, err := http.NewRequest("POST", link, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("user-agent", "Instagram 85.0.0.21.100 Android (28/9; 380dpi; 1080x2147; OnePlus; HWEVA; OnePlus6T; qcom; en_US; 146536611)")
	req.Header.Add("host", "i.instagram.com")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("accept-encoding", "gzip, deflate")
	req.Header.Add("x-fb-http-engine", "Liger")
	req.Header.Add("connection", "close")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode == 200 {
		return resp.Header.Values("set-cookie")[4], nil
	} else {
		return "", fmt.Errorf("Could not login.")
	}

}

func generate_uuid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

func clear_console() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("cmd", "/c", "clear")
	}
        cmd.Stdout = os.Stdout
        cmd.Run()
}

