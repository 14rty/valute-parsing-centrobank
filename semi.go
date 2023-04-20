package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//type ValCurs struct {
//	XMLName xml.Name `xml:"ValCurs"`
//	Text    string   `xml:"chardata"`
//	Date    string   `xml:"Date,attr"`
//	Name    string   `xml:"name,attr"`
//	Valute  []Val    `xml:"Valute"`
//}
//
//type Val struct {
//	Text     string `xml:"chardata"`
//	ID       string `xml:"ID,attr"`
//	NumCode  string `xml:"NumCode"`
//	CharCode string `xml:"CharCode"`
//	Nominal  string `xml:"Nominal"`
//	Name     string `xml:"Name"`
//	Value    string `xml:"Value"`
//}

type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Text    string   `xml:",chardata"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valute  []struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"ID,attr"`
		NumCode  string `xml:"NumCode"`
		CharCode string `xml:"CharCode"`
		Nominal  string `xml:"Nominal"`
		Name     string `xml:"Name"`
		Value    string `xml:"Value"`
	} `xml:"Valute"`
}

type DataValue struct {
	Date  string
	Name  string
	Value float64
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func Use(n int64) {

}

func TimeManager(count int) string { //определяем настоящее время и дату 90 дней назад, после чего по форме подставляем в url
	var currentTime time.Time
	var flagTime time.Time
	var rawURL = "http://www.cbr.ru/scripts/XML_daily_eng.asp?date_req="

	now := time.Now()
	currentTime = now.Add(-1 * time.Hour * 24 * time.Duration(count))
	switch currentTime {
	case flagTime:
		currentURL := rawURL + now.Format("02/01/2006")
		return currentURL
	default:
		currentURL := rawURL + currentTime.Format("02/01/2006")
		return currentURL
	}

}

func DeleteAll() {
	for i := 0; i < 90; i++ { //создаем ссылки для загрузки данных за последние 90 дней

		e := os.Remove((strconv.Itoa(i) + ".xml"))
		if e != nil {
			log.Fatal(e)
		}

		g := os.Remove(("r" + strconv.Itoa(i) + ".xml"))
		if g != nil {
			log.Fatal(g)
		}
	}
}

func ModiFile(source string) {
	fin, err := os.Open(source)
	if err != nil {
		panic(err)
	}
	defer fin.Close()

	fout, err := os.Create(("r" + source))
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	// удаляем первые 45 байт ответа
	_, err = fin.Seek(45, io.SeekStart)
	if err != nil {
		panic(err)
	}

	n, err := io.Copy(fout, fin)

	Use(n)

}

func main() {
	var currURL string
	//var files []string
	var count = 0
	for i := 0; i < 90; i++ { //создаем ссылки для загрузки данных за последние 90 дней
		currURL = TimeManager(count)
		count++
		//fmt.Println(currURL)

		DownloadFile((strconv.Itoa(i) + ".xml"), currURL)
		ModiFile(strconv.Itoa(i) + ".xml")
		//files = append(files, (strconv.Itoa(i) + ".xml"))
	}

	defer DeleteAll()
	average := [2]float64{0, 0}
	mxx := DataValue{}
	mnn := DataValue{"", "", 999999}
	for i := 0; i < 90; i++ {

		curs := ParseWinnersXml(("r" + strconv.Itoa(i) + ".xml"))

		for i := 0; i < len(curs.Valute); i++ {
			var err error
			// определение среднего курса рубля
			numba := 0.0
			numba2 := 0.0
			if numba, err = strconv.ParseFloat(strings.Replace(curs.Valute[i].Value, ",", ".", -1), 64); err != nil {
				fmt.Println("endif")
			}
			if numba2, err = strconv.ParseFloat(strings.Replace(curs.Valute[i].Nominal, ",", ".", -1), 64); err != nil {
				fmt.Println("endif")
			}
			average[0] += float64(numba) / float64(numba2) //добавляем стоимость только за единицу любой валюты
			average[1] += 1
			// определение самой дорогой валюты
			if float64(numba)/float64(numba2) > mxx.Value {
				mxx.Value = numba / numba2
				mxx.Name = curs.Valute[i].Name
				mxx.Date = curs.Date
			}
			// определение самой дешевой валюты
			if float64(numba)/float64(numba2) < mnn.Value {
				mnn.Value = numba / numba2
				mnn.Name = curs.Valute[i].Name
				mnn.Date = curs.Date
			}
		}

	}
	//чистим консоль
	cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Println("среднее значение курса рубля:", average[0]/average[1])
	fmt.Println("_______________________________")
	fmt.Println("максимальное значение валюты:")
	fmt.Println("значение:", mxx.Value)
	fmt.Println("название:", mxx.Name)
	fmt.Println("дата:", mxx.Date)
	fmt.Println("_______________________________")
	fmt.Println("минимальное значение валюты:")
	fmt.Println("значение:", mnn.Value)
	fmt.Println("название:", mnn.Name)
	fmt.Println("дата:", mnn.Date)
}

func ParseWinnersXml(path string) ValCurs {

	xmlFile, err := os.Open(path)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)

	rss := ValCurs{}
	xml.Unmarshal(byteValue, &rss)

	return rss
}
