package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jhoonb/archivex"
)

const (
	domain = "https://manazip.com"
)

func DownloadImage(strURL string, fileName string) error {
	res, err := http.Get(strURL)
	if err != nil {
		log.Fatalf("http.Get -> %v", err)
		return err
	}

	data, err := ioutil.ReadAll(res.Body)

	res.Body.Close()
	if err != nil {
		log.Fatalf("ioutil.ReadAll -> %v", err)
		return err
	}

	res.Body.Close()

	ioutil.WriteFile(fileName, data, 0666)

	return err
}

func BookScrape(eid string, title string) error {
	doc, err := goquery.NewDocument(domain + "/comic_view.php?pid=&cid=1&eid=" + eid)
	if err != nil {
		log.Fatal(err)
		return err
	}

	size := doc.Find(".document_img").Size()
	fmt.Print(size)
	doc.Find(".document_img").Each(func(i int, s *goquery.Selection) {
		src := s.AttrOr("src", "")
		filename := fmt.Sprintf("./%s/%04d.jpg", title, i)

		if strings.Index(src, "http") == -1 {
			src = domain + src
		}
		//fmt.Println("\tDOWNLOAD : "+src)
		fmt.Printf("\r%s - [%d/%d]", title, i+1, size)
		DownloadImage(src, filename)
	})
	fmt.Println("")
	return err
}

func BookArchive(filename string) {
	os.Chdir("./" + filename)

	zip := new(archivex.ZipFile)
	zip.Create("../" + filename + ".zip")
	zip.AddAll("./", true)
	zip.Close()
	os.Chdir("..")
}

func RemoveFolers() {
	files, _ := ioutil.ReadDir("./")
	i := 0
	for _, f := range files {
		if f.IsDir() {
			i = i + 1
			defer os.RemoveAll(f.Name())
		}
	}
	// if i > 0 {
	// 	RemoveFolers()
	// }
}

func BookEpisode(cid string, skip int) error {
	doc, err := goquery.NewDocument(domain + "/comics.php?pid=1&cat=1&cid=" + cid)
	if err != nil {
		log.Fatal(err)
		return err
	}

	bookTitle := doc.Find(".title").First().Text()
	os.Mkdir(bookTitle, 07000)
	os.Chdir("./" + bookTitle)

	fmt.Printf("TITLE : %s [SKIP : %d]\n\r", bookTitle, skip)
	// 분류 별로 찾기
	doc.Find(".episode_tr").Each(func(i int, s *goquery.Selection) {
		if i < skip {
			return
		}
		epdID := s.AttrOr("data-episode-id", "")
		titleSelection := s.Find(".toon_title").First()
		bookTitle := titleSelection.Find("span").First().Text()
		chapter := titleSelection.Text()
		chapter = strings.Replace(chapter, bookTitle, "", -1)
		chapter = strings.Trim(chapter, "\n\r ")
		chapter = strings.Replace(chapter, "!", "", -1)
		chapter = strings.Replace(chapter, "?", "", -1)
		chapter = strings.Replace(chapter, "\r", "", -1)
		chapter = strings.Replace(chapter, "\n", "-", -1)

		os.Mkdir("./"+chapter, 0777)
		// 책 스크랩
		BookScrape(epdID, chapter)
		// 책 압축
		BookArchive(chapter)
	})

	fmt.Println("\n\r\n\rDelete tmp folders")
	RemoveFolers()
	os.Chdir("..")
	return err
}

func FileRead(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sp := scanner.Text()
		sps := strings.Split(sp, "\t")
		skip := "0"
		if len(sps) > 1 {
			skip = sps[1]
		}
		bid := strings.Trim(sps[0], " ")
		skipCount, _ := strconv.Atoi(skip)
		BookEpisode(bid, skipCount)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("manzip [Cartoon ID] [SKIP COUNT]")
		os.Exit(2)
	}
	skip := 0
	if len(os.Args) == 3 {
		skip, _ = strconv.Atoi(os.Args[2])
	}
	file := os.Args[1]
	if _, err := os.Stat(file); err != nil {
		BookEpisode(os.Args[1], skip)
		os.Exit(0)
	}
	FileRead(file)
}
