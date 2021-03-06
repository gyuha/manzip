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

func downloadImage(strURL string, fileName string) error {
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

func bookScrape(eid string, title string) error {
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
		downloadImage(src, filename)
	})
	fmt.Println("")
	return err
}

func bookArchive(filename string) {
	os.Chdir("./" + filename)

	zip := new(archivex.ZipFile)
	zip.Create("../" + filename + ".zip")
	zip.AddAll("./", true)
	zip.Close()
	os.Chdir("..")
}

func removeFolers() {
	files, _ := ioutil.ReadDir("./")
	i := 0
	for _, f := range files {
		if f.IsDir() {
			i = i + 1
			// defer os.RemoveAll(f.Name())
			os.RemoveAll(f.Name())
		}
	}
	if i > 0 {
		// 지워지지 않는 경우가 발생해서 재도전
		removeFolers()
	}
}

func titleFilter(title string) string {
	rep := []string{"!", "?", "<", ">", "[", "]", "\n", "\t", "\r"}
	for _, each := range rep {
		title = strings.Replace(title, each, "", -1)
	}
	title = strings.Trim(title, "\n\r ")
	return title
}

func bookEpisode(cid string, skip int) error {
	doc, err := goquery.NewDocument(domain + "/comics.php?pid=1&cat=1&cid=" + cid)
	if err != nil {
		log.Fatal(err)
		return err
	}

	bookTitle := doc.Find(".title").First().Text()
	bookTitle = titleFilter(bookTitle)

	if len(bookTitle) == 0 {
		return err
	}
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
		chapter = strings.Trim(chapter, "\n\r\t ")
		chapter = titleFilter(chapter)

		if len(chapter) == 0 {
			return
		}
		os.Mkdir("./"+chapter, 0777)
		// 책 스크랩
		bookScrape(epdID, chapter)
		// 책 압축
		bookArchive(chapter)
	})

	fmt.Println("\n\rDelete tmp folders")
	removeFolers()
	os.Chdir("..")
	return err
}

func fileRead(filename string) error {
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
		bookEpisode(bid, skipCount)
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
		bookEpisode(os.Args[1], skip)
		os.Exit(0)
	}
	fileRead(file)
}
