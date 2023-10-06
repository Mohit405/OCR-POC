package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/julienschmidt/httprouter"
)

var textractSession *textract.Textract

func init() {
	textractSession = textract.New(session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-2"),
		Credentials: credentials.NewStaticCredentials("AKIAVYU3F7BN2HAJU5EF", "L/ClvF9e0ZgeXrNCploN+e65Be21efvW7pgV8ZMx", ""),
	})))
}

func (app *application) TextExtractor(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var maxRows int
	OcrData := []map[string]string{}
	colheaderId := []string{}
	colheaderValue := []string{}

	colvalueID := make(map[int][]string)
	coldata := make(map[int][]string)
	extractedData := make(map[string][]string)

	file, err := ioutil.ReadFile("invoice1.png")
	if err != nil {
		panic(err)
	}

	resp, err := textractSession.AnalyzeDocument(&textract.AnalyzeDocumentInput{
		Document: &textract.Document{
			Bytes: file,
		},
		FeatureTypes: aws.StringSlice([]string{"TABLES"}),
	})

	if err != nil {
		panic(err)
	}

	fmt.Fprint(w, resp.String())

	tablesids := []string{}
	for i := 0; i < len(resp.Blocks); i++ {
		if *resp.Blocks[i].BlockType == "TABLE" && *resp.Blocks[i].EntityTypes[0] == "STRUCTURED_TABLE" {
			for _, val := range resp.Blocks[i].Relationships[0].Ids {
				tablesids = append(tablesids, *val)
			}
			break
		}
	}

	for i := 0; i < len(tablesids); i++ {
		for j := 0; j < len(resp.Blocks); j++ {
			if *resp.Blocks[j].BlockType == "CELL" && *resp.Blocks[j].Id == tablesids[i] && *resp.Blocks[j].RowIndex == 1 {
				colheaderId = append(colheaderId, *resp.Blocks[j].Relationships[0].Ids[0])
				break
			}
		}
	}

	for i := 0; i < len(colheaderId); i++ {
		for j := 0; j < len(resp.Blocks); j++ {
			if *resp.Blocks[j].BlockType == "LINE" && *resp.Blocks[j].Relationships[0].Ids[0] == colheaderId[i] {
				colheaderValue = append(colheaderValue, *resp.Blocks[j].Text)
				break
			}
		}
	}

	for i := len(colheaderId); i < len(tablesids); i++ {
		for j := 0; j < len(resp.Blocks); j++ {
			if *resp.Blocks[j].BlockType == "CELL" && *resp.Blocks[j].Id == tablesids[i] && (*resp.Blocks[j].RowIndex == 2 || *resp.Blocks[j].RowIndex == 3) {
				if len(resp.Blocks[j].Relationships) > 0 {
					index := resp.Blocks[j].ColumnIndex
					colvalueID[int(*index)] = append(colvalueID[int(*index)], *resp.Blocks[j].Relationships[0].Ids[0])
				}

				break
			}
		}
	}

	for key, val := range colvalueID {
		for _, id := range val {
			for j := 0; j < len(resp.Blocks); j++ {
				if *resp.Blocks[j].BlockType == "LINE" && *resp.Blocks[j].Relationships[0].Ids[0] == id {
					coldata[key] = append(coldata[key], *resp.Blocks[j].Text)
					size := len(coldata[key])
					if maxRows < size {
						maxRows = size
					}
					break
				}
			}
		}

	}

	for i := 1; i <= len(colheaderValue); i++ {
		if len(coldata[i]) == 0 {
			for j := 0; j < maxRows; j++ {
				extractedData[colheaderValue[i-1]] = append(extractedData[colheaderValue[i-1]], "")
			}

		} else {
			extractedData[colheaderValue[i-1]] = append(extractedData[colheaderValue[i-1]], coldata[i]...)
			maxRows = len(coldata[i])
		}
	}

	log.Println(extractedData)
	for i := 0; i < maxRows; i++ {
		temp := map[string]string{}
		for j := 0; j < len(colheaderValue); j++ {
			temp[colheaderValue[j]] = extractedData[colheaderValue[j]][i]
		}

		OcrData = append(OcrData, temp)
	}
	data, err := json.MarshalIndent(OcrData, "", "\t")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(data))
}
