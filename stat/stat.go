package stat

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// CsvData stores useful data from input file
type CsvData struct {
	Url            string
	ErrCode        string
	ErrDescription string
}

// Resource ...
type Resource struct {
	CsvErr string
	url    string
	err    error
}

// Stat stores analysis data
type Stat struct {
	Data map[string][]*Resource
	info map[string]map[string]int32
	m    sync.RWMutex
}

// NewStat Stat constructor
func NewStat() *Stat {
	return &Stat{
		Data: make(map[string][]*Resource),
		info: make(map[string]map[string]int32),
	}
}

// QueryResource generates HTTP request and append result to Stat
// Reads from CsvData channel
func (st *Stat) QueryResource(info chan CsvData) {
	for v := range info {
		var err error
		var k string

		req, err := http.NewRequest(http.MethodGet, v.Url, nil)
		if err != nil {
			log.Println(err)
		}

		rsp, err := http.DefaultClient.Do(req)

		// Sync map writes
		st.m.Lock()
		if rsp != nil {
			k = strconv.Itoa(rsp.StatusCode)
		} else {
			k = "no_status"
		}

		st.Data[k] = append(st.Data[k], &Resource{
			url:    v.Url,
			CsvErr: fmt.Sprintf("%s - %s", v.ErrCode, v.ErrDescription),
			err:    err,
		})
		st.m.Unlock()
	}
}

// Info generates and prints analysis data
func (st *Stat) Info() {
	for k, v1 := range st.Data {
		for _, v2 := range v1 {
			if st.info[k] == nil {
				st.info[k] = make(map[string]int32)
			}
			st.info[k][v2.CsvErr]++
		}

		fmt.Println("\n-------------------------------------------")
		fmt.Println("STATUS CODE:        ", k)
		fmt.Println("COUNT:              ", len(v1))

		fmt.Println("\nERROR BREAKDOWN:    ")
		for k2, v2 := range st.info[k] {
			fmt.Println("ORIGINAL ERR: ", k2)
			fmt.Println("COUNT:        ", v2)
		}
	}
}
