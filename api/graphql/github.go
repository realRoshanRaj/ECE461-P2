package graphql

import (
	dep "CLI/utils"
	"context"
	"log"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/machinebox/graphql"
)

type RespDataql1 struct { //type that storeLogs data from graphql
	Repository struct {
		Issues struct {
			TotalCount int
		}
		PullRequests struct {
			TotalCount int
		}
		Upcase struct { //README.md
			Text string
		}
		Downcase struct { //readme.md
			Text string
		}
		Capcase struct { //Readme.md
			Text string
		}
		Expcase struct { //readme.markdown
			Text string
		}
		Commits struct {
			History struct {
				TotalCount int
			}
		}
	}
}

type respDataql2 struct {
	Repository struct {
		Issues struct {
			TotalCount int `json:"totalCount"`
		} `json:"issues"`
		PullRequests struct {
			Nodes []struct {
				CreatedAt string `json:"createdAt"`
				MergedAt  string `json:"mergedAt"`
			} `json:"nodes"`
		} `json:"pullRequests"`
	} `json:"repository"`
}

func Graphql_func(repo_owner string, repo_name string, token string) []float64 { //seems to be working as long as token is storeLogd in tokens.env
	// create a new client
	client := graphql.NewClient("https://api.github.com/graphql")

	scores := [5]float64{0, 0, 0, 0, 0} //[license, RampUp, Correctness, Bus Factor(total commits before going into construct()), Responsive Maintainer]

	// make a request
	req1 := graphql.NewRequest(`
	query {
		repository(owner:"` + repo_owner + `", name:"` + repo_name + `") {
			issues(states: OPEN) {
				totalCount
			}
			pullRequests(states: MERGED){
				totalCount
			}
			upcase: object(expression: "HEAD:README.md") {
				... on Blob {
					text
				}
			}
			downcase: object(expression: "HEAD:README.md") {
				... on Blob {
					text
				}
			}
			capcase: object(expression: "HEAD:Readme.md") {
				... on Blob {
					text
				}
			}
			expcase: object(expression: "HEAD:readme.markdown") {
				... on Blob {
					text
				}
			}
			commits: object(expression: "HEAD") {
				... on Commit {
				  history {
					   totalCount
					}
				}
			}
		}
	}
	`)

	req1.Header.Add("Authorization", "Bearer "+token)
	var respData1 RespDataql1
	if err := client.Run(context.Background(), req1, &respData1); err != nil {
		log.Fatal(err)
	}

	//fmt.Println("Number of issues:", respData1.Repository.Downcase.Text)
	//40% of the last pull requests perhaps arbitrary number
	perc_PR1 := math.Min(20, float64(respData1.Repository.PullRequests.TotalCount)*float64(0.4))
	perc_PR := int(perc_PR1)
	//fmt.Println(perc_PR)

	req2 := graphql.NewRequest(`
	query {
		repository(owner:"` + repo_owner + `", name:"` + repo_name + `") {
			issues(states: CLOSED) {
				totalCount
			}
			pullRequests (last: ` + strconv.Itoa(perc_PR) + `, states: MERGED) {
				nodes{
					createdAt
					mergedAt
				}
			}
		}
	}
	`)
	req2.Header.Add("Authorization", "Bearer "+token)

	var respData2 respDataql2
	if err := client.Run(context.Background(), req2, &respData2); err != nil {
		log.Fatal(err)
	}

	difference_sum := 0.0

	for i := 0; i < perc_PR; i++ {
		date1 := respData2.Repository.PullRequests.Nodes[i].MergedAt

		y1, err := strconv.Atoi(date1[0:3])
		if err != nil {
			return scores[:]
		}
		m1, err := strconv.Atoi(date1[5:6])
		if err != nil {
			return scores[:]
		}
		d1, err := strconv.Atoi(date1[8:9])
		if err != nil {
			return scores[:]
		}
		h1, err := strconv.Atoi(date1[11:12])
		if err != nil {
			return scores[:]
		}
		date2 := respData2.Repository.PullRequests.Nodes[i].CreatedAt
		y2, err := strconv.Atoi(date2[0:3])
		if err != nil {
			return scores[:]
		}
		m2, err := strconv.Atoi(date2[5:6])
		if err != nil {
			return scores[:]
		}
		d2, err := strconv.Atoi(date2[8:9])
		if err != nil {
			return scores[:]
		}
		h2, err := strconv.Atoi(date2[11:12])
		if err != nil {
			return scores[:]
		}

		firstDate := time.Date(y1, time.Month(m1), d1, h1, 0, 0, 0, time.UTC)
		secondDate := time.Date(y2, time.Month(m2), d2, h2, 0, 0, 0, time.UTC)
		difference_sum += math.Abs(firstDate.Sub(secondDate).Hours())
	}

	difference := difference_sum / float64(perc_PR)

	//time it takes to resolve, 7 days is the max, otherwise its a zero
	if difference > float64(168) {
		scores[4] = 0
	} else {
		scores[4] = dep.RoundFloat(1-(float64(difference)/float64(168)), 3)
	}

	//closed issues / total issues score of correctness
	scores[2] = dep.RoundFloat(float64(respData2.Repository.Issues.TotalCount)/(float64(respData1.Repository.Issues.TotalCount)+float64(respData2.Repository.Issues.TotalCount)), 3)

	//rampup... has readme
	if respData1.Repository.Upcase.Text != "" {
		rm_len := float64(len(respData1.Repository.Upcase.Text))
		// fmt.Println(rm_len)
		if rm_len/float64(1000) > 5 {
			scores[1] = 1
		} else {
			scores[1] = rm_len / 5000
		}

		res1, e := regexp.MatchString(`MIT [lL]icense|[lL]icense MIT|\[MIT\]\(LICENSE\)|\[MIT\]\(\.\/LICENSE\)|lgpl-2.1|License of zlib| zlib license|Berkeley Database License|Sleepycat|Boost Software License|CeCILL version 2|Clarified Artistic License|
		Cryptix General License|EU DataGrid Software License|Eiffel Forum License, version 2|Expat License|Intel Open Source License|License of Guile|
		License of Netscape Javascript|License of Perl|Python 1.6a2|Python 2.0.1 license|Python 2.1.1 license|Python [2-9].[1-9].[1-9]|Vim version [6-9].[2-9]|
		iMatix Standard Function Library|License of the run-time units of the GNU Ada compiler|Modified BSD license|OpenLDAP License.*version 2.7|Public Domain|
		Standard ML of New Jersey Copyright License|The license of Ruby|W3C Software Notice and License|X11 License|
		Zope Public License, version 2.0|eCos license, version 2.0`, respData1.Repository.Upcase.Text)
		if res1 {
			scores[0] = 1
		} else {
			scores[0] = 0
		}
		if e != nil {
			return scores[:]
		}
	} else if respData1.Repository.Downcase.Text != "" {
		rm_len := float64(len(respData1.Repository.Downcase.Text))
		if rm_len/float64(1000) > 5 {
			scores[1] = 1
		} else {
			scores[1] = rm_len / 5000
		}

		res1, e := regexp.MatchString(`MIT [lL]icense|[lL]icense MIT|\[MIT\]\(LICENSE\)|\[MIT\]\(\.\/LICENSE\)|lgpl-2.1|License of zlib| zlib license|Berkeley Database License|Sleepycat|Boost Software License|CeCILL version 2|Clarified Artistic License|
		Cryptix General License|EU DataGrid Software License|Eiffel Forum License, version 2|Expat License|Intel Open Source License|License of Guile|
		License of Netscape Javascript|License of Perl|Python 1.6a2|Python 2.0.1 license|Python 2.1.1 license|Python [2-9].[1-9].[1-9]|Vim version [6-9].[2-9]|
		iMatix Standard Function Library|License of the run-time units of the GNU Ada compiler|Modified BSD license|OpenLDAP License.*version 2.7|Public Domain|
		Standard ML of New Jersey Copyright License|The license of Ruby|W3C Software Notice and License|X11 License|
		Zope Public License, version 2.0|eCos license, version 2.0`, respData1.Repository.Downcase.Text)
		if res1 {
			scores[0] = 1
		} else {
			scores[0] = 0
		}
		if e != nil {
			return scores[:]
		}
	} else if respData1.Repository.Capcase.Text != "" {
		rm_len := float64(len(respData1.Repository.Capcase.Text))
		// fmt.Println(rm_len)
		if rm_len/float64(1000) > 5 {
			scores[1] = 1
		} else {
			scores[1] = rm_len / 5000
		}

		res1, e := regexp.MatchString(`MIT [lL]icense|[lL]icense MIT|\[MIT\]\(LICENSE\)|\[MIT\]\(\.\/LICENSE\)|lgpl-2.1|License of zlib| zlib license|Berkeley Database License|Sleepycat|Boost Software License|CeCILL version 2|Clarified Artistic License|
		Cryptix General License|EU DataGrid Software License|Eiffel Forum License, version 2|Expat License|Intel Open Source License|License of Guile|
		License of Netscape Javascript|License of Perl|Python 1.6a2|Python 2.0.1 license|Python 2.1.1 license|Python [2-9].[1-9].[1-9]|Vim version [6-9].[2-9]|
		iMatix Standard Function Library|License of the run-time units of the GNU Ada compiler|Modified BSD license|OpenLDAP License.*version 2.7|Public Domain|
		Standard ML of New Jersey Copyright License|The license of Ruby|W3C Software Notice and License|X11 License|
		Zope Public License, version 2.0|eCos license, version 2.0`, respData1.Repository.Capcase.Text)
		if res1 {
			scores[0] = 1
		} else {
			scores[0] = 0
		}
		if e != nil {
			return scores[:]
		}
	} else if respData1.Repository.Expcase.Text != "" {
		rm_len := float64(len(respData1.Repository.Expcase.Text))
		if rm_len/float64(1000) > 5 {
			scores[1] = 1
		} else {
			scores[1] = rm_len / 5000
		}

		res1, e := regexp.MatchString(`MIT [lL]icense|[lL]icense MIT|\[MIT\]\(LICENSE\)|\[MIT\]\(\.\/LICENSE\)|lgpl-2.1|License of zlib| zlib license|Berkeley Database License|Sleepycat|Boost Software License|CeCILL version 2|Clarified Artistic License|
Cryptix General License|EU DataGrid Software License|Eiffel Forum License, version 2|Expat License|Intel Open Source License|License of Guile|
License of Netscape Javascript|License of Perl|Python 1.6a2|Python 2.0.1 license|Python 2.1.1 license|Python [2-9].[1-9].[1-9]|Vim version [6-9].[2-9]|
iMatix Standard Function Library|License of the run-time units of the GNU Ada compiler|Modified BSD license|OpenLDAP License.*version 2.7|Public Domain|
Standard ML of New Jersey Copyright License|The license of Ruby|W3C Software Notice and License|X11 License|
Zope Public License, version 2.0|eCos license, version 2.0`, respData1.Repository.Expcase.Text)
		if res1 {
			scores[0] = 1
		} else {
			scores[0] = 0
		}
		if e != nil {
			return scores[:]
		}
	} else {
		scores[1] = 0
	}

	//will serve as denominator *NOT FINAL SCORE*
	scores[3] = float64(respData1.Repository.Commits.History.TotalCount)

	return scores[:]
}
