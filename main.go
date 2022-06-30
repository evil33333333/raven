package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tidwall/gjson"
)

type MediaData struct {
	is_video bool
	url string
}

func main() {
	var  (
		session string
		thread_id string
	)
	fmt.Print("[+] Session: ")
	fmt.Scanln(&session)

	fmt.Print("[+] Thread ID: ")
	fmt.Scanln(&thread_id)
	raven_links := get_raven_links(session, thread_id)
	download_raven_links(raven_links)

	fmt.Println("[!] Finished downloading all disappearing media...")
	time.Sleep(time.Second * 5)

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


