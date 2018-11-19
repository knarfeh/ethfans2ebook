// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	elastic "gopkg.in/olivere/elastic.v5"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		main()
	},
}

func init() {
	RootCmd.AddCommand(fetchCmd)
}

func getUrlSlice(doc *goquery.Document, hrefs *[]string) {
	doc.Find(".post-item .post-info").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Find("a").Attr("href")
		*hrefs = append(*hrefs, "https://ethfans.org"+href)
	})
}

func main() {
	fmt.Println("ethfans2ebook running...")
	URL := viper.GetString("URL")
	DAYTIMESTAMP := viper.GetString("DAY_TIME_STAMP")
	ESHOSTPORT := viper.GetString("ES_HOST_PORT")
	viper.SetDefault("ROUTINE_NUM", 2)

	esClient, err := elastic.NewClient(elastic.SetURL(ESHOSTPORT))
	if err != nil {
		log.Printf("Unable to connect es")
		panic(err)
	}
	defer esClient.Stop()

	bulkRequest := esClient.Bulk()
	type esDoc struct {
		Title        string `json:"title"`
		Author       string `json:"author"`
		Content      string `json:"content"`
		URL          string `json:"url"`
		DayTimeStamp string `json:"dayTimestamp"`
	}

	doc, err := goquery.NewDocument(URL)
	if err != nil {
		fmt.Println("Network issues...")
		log.Fatal(err)
	}

	items := doc.Find(".pagination .item")
	pageNum, _ := strconv.Atoi(items.Get(-2).LastChild.Data)
	fmt.Println(pageNum)

	var hrefs []string
	getUrlSlice(doc, &hrefs)
	for i := 2; i < pageNum+1; i++ {
		nowURL := fmt.Sprintf("%s?page=%d", URL, i)
		fmt.Printf("Now url: %s\n", nowURL)
		doc, _ := goquery.NewDocument(nowURL)
		getUrlSlice(doc, &hrefs)
	}
	fmt.Printf("hrefs: %v\n", hrefs)

	for _, href := range hrefs {
		document, _ := goquery.NewDocument(href)
		articleTitle := document.Find(".post-content .title").Text()
		articleContent, _ := document.Find(".post-content .content").Html()
		fmt.Printf("Got %s\n", articleTitle)
		if articleTitle == "" {
			articleTitle = "No title"
		}

		d := esDoc{
			Title:        articleTitle,
			Author:       "ethfans",
			Content:      articleContent,
			URL:          href,
			DayTimeStamp: DAYTIMESTAMP,
		}
		bulkData := elastic.NewBulkIndexRequest().Index("ethfans").Type(URL + ":content").Id(articleTitle).Doc(d)
		bulkRequest = bulkRequest.Add(bulkData)
	}

	type metaData struct {
		Type     string `json:"type"`
		Title    string `json:"title"`
		BookDesp string `json:"book_desp"`
	}
	m := metaData{
		Type:     "ethfans",
		Title:    "ethfans-org",
		BookDesp: "ethfans",
	}
	bulkMetaData := elastic.NewBulkIndexRequest().Index("eebook").Type("metadata").Id(URL).Doc(m)
	bulkFinalRequest := bulkRequest.Add(bulkMetaData)
	_, err = bulkFinalRequest.Do(context.TODO())
	if err != nil {
		fmt.Println("err: ", err)
	}
}
