package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"os"
	"regexp"
	"strings"
	"time"
)

type Article struct {
	URI     string `json:"uri"`
	Title   string `json:"title"`
	Preview string `json:"preview"`
	Date    string `json:"date"`
	Source  string `json:"source"`
	Body    string `json:"body"`
}

var (
	articleListFilename = "./raw/article-list.json"
	fullarticlesFilename = "./prepared/full-articles.json"
)

func main() {
	var articles []Article
		if _, err := os.Stat(articleListFilename); os.IsNotExist(err) {
		fmt.Println("The article list file does not exist, collecting articles from google")
		articles = collectArticles()
	} else {
		fmt.Println("The article list file exists, reading articles from it")
		al, err := os.ReadFile(articleListFilename)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(al, &articles)
		if err != nil {
			panic(err)
		}
	}
	for i := range articles {
		processArticle(&articles[i])
	}
	writeJSON(articles, fullarticlesFilename)
	fmt.Printf("Total number of articles extracted = %d\n", len(articles))
}

func processArticle(a *Article) {
	collector := colly.NewCollector()
	texts := []string{}
	collector.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("p", func(_ int, el *colly.HTMLElement) {
			texts = append(texts, el.Text)
		})
		a.Body = strings.Join(texts, "\n")
	})
	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting", request.URL.String())
		time.Sleep(1 * time.Second)
	})
	collector.Visit(a.URI)
}

func collectArticles() []Article {
	articles := make([]Article, 0)
	uris := make(map[string]int)
	collector := colly.NewCollector(colly.AllowedDomains("google.com", "www.google.com"))
	pattern := `\/url\?q=(.*?)&sa=U`
	re := regexp.MustCompile(pattern)
	collector.OnHTML(".Gx5Zad", func(element *colly.HTMLElement) {
		uri := element.ChildAttr("a", "href")
		match := re.FindStringSubmatch(uri)
		if match == nil || len(match) < 2 {
			return
		}
		uri = match[1]
		_, found := uris[uri]
		if found {
			fmt.Println("Found a duplicate")
			return
		}
		title := element.ChildText("div.vvjwJb")
		source := element.ChildText("div.UPmit")
		preview := element.ChildText("div.s3v9rd")
		date := element.ChildText("span.r0bn4c")
		article := Article{
			URI:     uri,
			Title:   title,
			Source:  source,
			Preview: preview,
			Date:    date,
		}
		articles = append(articles, article)
		uris[uri] = 1
	})

	collector.OnRequest(func(request *colly.Request) {
		fmt.Println("Visiting", request.URL.String())
		time.Sleep(1 * time.Second)
	})

	for i := 0; i < 47; i++ {
		site := "https://www.google.com/search?q=%22major+league+cricket%22&source=lmns&tbm=nws&bih=938&biw=1920&hl=en&start=" + fmt.Sprintf("%d", i*10)
		collector.Visit(site)
	}

	fmt.Printf("Total number = %d\n", len(articles))
	writeJSON(articles, articleListFilename)
	return articles
}

func writeJSON(data []Article, fname string) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		fmt.Println("Unable to create json file")
		return
	}

	_ = os.WriteFile(fname, file, 0644)
}
