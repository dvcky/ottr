package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/proto"

	"os"
	pb "ottr/pb"
	"strconv"
	"strings"
)

// OTTR (otter) - an Open Train Tracker, Reimagined

type ottrMap struct {
	size   [2]int
	status [2]string
	// because creating 2-dimensional arrays using runtime-calculated variables is...jank, we are just going to have a static
	// cap that seems fairly reasonable. worst case scenario, we can recompile with different values here
	oMap [60][60]string
	// stucture: 0-25=A-Z,26-35=0-9,36-60=other(colors, names, etc.). next index is stop number on that line
	oPos [60][100]string

	nUsed [60]string
}

type ottrTrip struct {
	id   string
	line string
	head string
	stop string
}

func main() {

	f, err := os.Open("map.csv")
	if err != nil {
		log.Fatal(err)
	}

	var mapData = loadMap(f)
	var mapDataBackup ottrMap
	copier.Copy(&mapDataBackup, &mapData)

	defer f.Close()

	printMap(mapData, mapDataBackup, getFeed())

	for range time.Tick(time.Second * 15) {
		printMap(mapData, mapDataBackup, getFeed())
	}
}

func getFeed() []ottrTrip {
	// create client
	client := http.Client{}

	// configure request
	req, _ := http.NewRequest("GET", "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs-ace", nil)
	req.Header = http.Header{"x-api-key": {"PUT YOUR API KEY HERE"}}

	// send, close, and read request
	res, _ := client.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	// create feed and send it
	feed := pb.FeedMessage{}
	proto.Unmarshal(body, &feed)

	var ottrFeed []ottrTrip

	for _, entity := range feed.Entity {
		tripUpdate := entity.GetTripUpdate()
		if tripUpdate != nil {
			var tempTrip ottrTrip

			var tripInfo = parseId(tripUpdate.GetTrip().GetTripId())
			tempTrip.id = tripInfo[0]
			tempTrip.line = tripInfo[1]
			tempTrip.head = tripInfo[2]

			if len(tripUpdate.GetStopTimeUpdate()) != 0 {
				tempTrip.stop = tripUpdate.GetStopTimeUpdate()[0].GetStopId()
			}

			ottrFeed = append(ottrFeed, tempTrip)
		}
	}

	return ottrFeed
}

func printMap(oMap ottrMap, oBack ottrMap, oData []ottrTrip) {

	for _, oTrip := range oData {
		if oTrip.stop != "" {

			out, _ := strconv.Atoi(strings.TrimLeft(strings.TrimRight(oTrip.stop, oTrip.head), oTrip.line))

			var test = oMap.oPos[hashLine(oMap, oTrip.line)][out]
			if test != "" {
				var testSplit = strings.Split(test, ",")
				cR, _ := strconv.Atoi(testSplit[0])
				cC, _ := strconv.Atoi(testSplit[1])
				oMap.oMap[cR][cC] = oMap.status[1]
			}
		}
	}
	for i := 0; i < 100; i++ {
		fmt.Println()
	}
	for row := 0; row < oMap.size[0]; row++ {
		for col := 0; col < oMap.size[1]; col++ {
			if oMap.oMap[row][col] != "" {
				fmt.Print(oMap.oMap[row][col] + " ")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
	copier.Copy(&oMap, &oBack)
}

func parseId(tripId string) []string {
	var parseArray [3]string

	var firstSplit = strings.Split(tripId, "_")
	var secondSplit = strings.Split(firstSplit[1], ".")

	parseArray[0] = firstSplit[0]
	parseArray[1] = secondSplit[0]
	parseArray[2] = secondSplit[len(secondSplit)-1]

	return parseArray[:]
}

func loadMap(mapFile *os.File) ottrMap {

	var tempMap ottrMap

	csvFile := csv.NewReader(mapFile)
	data, err := csvFile.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	tempMap.status[0] = data[0][0]                // inactive stop
	tempMap.status[1] = data[0][1]                // active stop
	tempMap.size[0], _ = strconv.Atoi(data[1][1]) // map height
	tempMap.size[1], _ = strconv.Atoi(data[1][0]) // map width

	for row := 3; row < len(data); row++ {
		for col := 0; col < len(data[row]); col++ {
			if strings.HasPrefix(data[row][col], "STOP:") {
				var noPfx = strings.TrimPrefix(data[row][col], "STOP:")
				var stopData = strings.Split(noPfx, "-")
				stopIndex, _ := strconv.Atoi(stopData[1])
				tempMap.oPos[hashLine(tempMap, stopData[0])][stopIndex] = strconv.Itoa(row-3) + "," + strconv.Itoa(col)
				tempMap.oMap[row-3][col] = data[0][0]
			} else {
				tempMap.oMap[row-3][col] = data[row][col]
			}
		}
	}

	return tempMap
}

func hashLine(oMap ottrMap, line string) int {
	for i := 0; i < len(oMap.oPos); i++ {
		if oMap.oPos[i][0] == line {
			return i
		}
	}
	for i := 0; i < len(oMap.nUsed); i++ {
		if oMap.nUsed[i] == line || oMap.nUsed[i] == "" {
			return i
		}
	}
	return -1
}
