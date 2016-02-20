package main

import (
	"fmt"
	"log"
  "os"
  "io/ioutil"
  "strings"
  "net/http"
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

func BookScrape(eid string, title string) error{
	doc, err := goquery.NewDocument(domain+"/comic_view.php?pid=&cid=1&eid="+eid)
	if err != nil {
		log.Fatal(err)
    return err
	}

	// title := (strings.Replace(doc.Find("title").Text(), " - 만화ZIP(MANAZIP)", "", 1))
	// fmt.Println(title)

  os.Mkdir(title, 0777)
  doc.Find(".document_img").Each(func(i int, s *goquery.Selection) {
    src := s.AttrOr("src", "")
    filename := fmt.Sprintf("./%s/%04d.jpg",  title, i)

    if(strings.Index(src, "http") == -1) {
      src = domain+src
    }
    fmt.Println("\tDOWNLOAD : "+src)
    DownloadImage(src, filename)
    // .Printf("%s\n", s.Attr("src"))
  })

  return err
}

func BookArchive(folder string) {
  os.Chdir("./"+folder)
  zip := new(archivex.ZipFile)
  zip.Create("../"+folder+".zip")
  zip.AddAll("./", true)
  zip.Close()
  os.Chdir("..")
}

func BookEpisode(cid string) error {
  doc, err := goquery.NewDocument(domain+"/comics.php?pid=1&cat=1&cid="+cid)
	if err != nil {
		log.Fatal(err)
    return err
	}

  bookTitle := doc.Find(".title").First().Text()
  os.Mkdir(bookTitle, 07000)
  os.Chdir("./"+bookTitle)

  doc.Find(".episode_tr").Each(func(i int, s *goquery.Selection) {
    epdID := s.AttrOr("data-episode-id", "")
    titleSelection := s.Find(".toon_title").First()
    bookTitle := titleSelection.Find("span").First().Text()
    chapter := titleSelection.Text()
    chapter = strings.Replace(chapter, bookTitle, "", -1)
    chapter = strings.Trim(chapter, "\r\n ")
    chapter = strings.Replace(chapter, "\r", "", -1)
    chapter = strings.Replace(chapter, "\n", "-", -1)

    fmt.Println(chapter)
    BookScrape(epdID, chapter)
    BookArchive(chapter)
  })
  os.Chdir("..")
  return err
}

func main() {
  if( len(os.Args) < 2) {
    fmt.Println("manzip [Cartoon ID]")
    os.Exit(2)
  }
  BookEpisode(os.Args[1])
}
