package monitor

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"

	redis "gopkg.in/redis.v2"
)

//Report struct for holding app info
type Report struct {
	DateStarted time.Time
	Name        string
	PID         string
	IsRunning   bool
	IsTest      bool
	Redis       *redis.Client
}

func (r *Report) closeRedis() {
	if r.Redis != nil {
		r.Redis.Close()
	}
}

// Get value func
func Get(k string) string {
	r := getNewRedisConnection()

	defer r.Close()

	v := r.Get(k)

	vs, err := v.Result()

	if err != nil {
		fmt.Println("Report.Get err 2", err, k)
		return ""
	}

	return vs
}

// AddAppToList func for adding a new app to the master list of apps
func AddAppToList(name string) {
	r := getNewRedisConnection()

	defer r.Close()

	var appList []string

	apps := Get("apps")

	_ :json.Unmarshal([]byte(apps), &appList)

	if index(appList, name) == -1 {
		appList = append(appList, name)
	}

    a, _ := json.Marshal(appList)

	status := r.Set("apps", string(a))

	if status.Err() != nil {
		fmt.Println("AddApp Error", status.Err().Error())
	}

	return
}

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func (r *Report) openRedis() {
	if r.Redis == nil {
		r.Redis = getNewRedisConnection()
	}
}

func (r *Report) Save() string {
	r.openRedis()

	defer r.closeRedis()

	encoded, err := json.Marshal(r)
	if err != nil {
		fmt.Println("error marshaling")
		return "error marshaling"
	}

	status := r.Redis.Set(r.Name, string(encoded))

	fmt.Println("Set New App", status)

	if status.Err() != nil {
		return status.Err().Error()
	}

	go AddAppToList(r.Name)

	return ""
}

//CreateSaveReport create and save report into apps db
func (r *Report) CreateSaveReport() error {
	encoded, err := json.Marshal(r)
	if err != nil {
		fmt.Println("error marshaling")
		return err
	}

	db, err := bolt.Open("apps.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	fmt.Println("successfully opened apps.db")

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("apps"))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(r.Name), encoded)
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

func GetReport(key []byte) {
	db, err := bolt.Open("apps.db", 0600, nil)
	if err != nil {
		fmt.Println("GetReport err 3", err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("apps"))
		v := b.Get(key)
		fmt.Printf("%sn", v)
		return nil
	})
}
