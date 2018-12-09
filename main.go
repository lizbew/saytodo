package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	DEFAULT_CONFIG_FILE = "config.toml"
	MONGO_URL           = "mongodb://127.0.0.1:27017"
	MONGO_DB_TASK       = "task"
	MONGO_COLL_TASKS    = "tasks"
	STATUS_NEW          = "new"
	STATUS_PROGRESS     = "progress"
	STATUS_CANCEL       = "cancel"
	STATUS_COMPLETE     = "complete"
	MODEL_TASK          = "task"
	MODEL_BOOKMARK      = "bookmark"
)

var (
	configFile string
	config     serverConfig
)

type serverConfig struct {
	MongoUrl    string `toml:"mongo_url"`
	AccountInfo string `toml:"account_info"`
}

type TodoTask struct {
	//Id           string    `json:"id" bson:"_id"`
	Id           bson.ObjectId `json:"id" bson:"_id"`
	Title        string        `json:"title" bson:"title"`     // binding:"required"
	Content      string        `json:"content" bson:"content"` //binding:"required"
	Link         string        `json:"link" bson:"link"`       //binding:"required"
	Status       string        `json:"status" bson:"status"`
	UserId       string        `json:"user_id" bson:"user_id"`
	ModelType    string        `json:"-" bson:"model_type"`
	CreatedTime  time.Time     `json:"created_time" bson:"created_time"`
	ModifiedTime time.Time     `json:"modified_time" bson:"modified_time"`
}

type TaskListResponse struct {
	List  []TodoTask `json:"list"`
	Count int        `json:"count"`
}

func addNewTask(c *gin.Context) {
	var jTask TodoTask

	if err := c.ShouldBindJSON(&jTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet(gin.AuthUserKey).(string)

	t := time.Now()
	jTask.Id = bson.NewObjectId()
	jTask.CreatedTime = t
	jTask.ModifiedTime = t
	jTask.Status = STATUS_NEW
	jTask.ModelType = MODEL_TASK
	jTask.UserId = userId

	//newTask := TodoTask{"test-task", "content", "new", time.Now(), time.Now()}

	err := storage_insertOne(&jTask)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("addNewTask completed")
	}

	c.JSON(http.StatusOK, jTask)

}

func queryTaskList(c *gin.Context) {

	//task := TodoTask{}
	//iter := mgoCollection_Task.Find(nil).Iter()
	//for iter.Next(&task) {
	//	taskListResponse.List = append(taskListResponse.List, task)
	//}

	var taskListResponse TaskListResponse
	storage_selectAll(&taskListResponse)

	c.JSON(http.StatusOK, taskListResponse)
}

func queryTask(c *gin.Context) {
	id := c.Params.ByName("id")
	//objectId := bson.ObjectIdHex(id)

	userId := c.MustGet(gin.AuthUserKey).(string)

	// mgoCollection_Task := getTaskCollection()
	//err := mgoCollection_Task.Find(bson.M{"_id": objectId}).One(&task)
	//err := mgoCollection_Task.FindId(objectId).One(&task)
	task, err := storage_selectOne(id)

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "not found"})
	} else if userId != task.UserId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	} else {
		c.JSON(http.StatusOK, task)
	}

}

func updateTask(c *gin.Context) {
	id := c.Params.ByName("id")

	userId := c.MustGet(gin.AuthUserKey).(string)

	var jTask TodoTask
	var existingTask *TodoTask
	var err error

	if err := c.ShouldBindJSON(&jTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if existingTask, err = storage_selectOne(id); err != nil {
		fmt.Println("updateTask  - query task error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if existingTask.UserId != userId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	updatBson := bson.M{}
	if len(jTask.Status) > 0 {
		updatBson["status"] = jTask.Status
	}
	if len(jTask.Title) > 0 {
		updatBson["title"] = jTask.Title
	}
	if len(jTask.Content) > 0 {
		updatBson["content"] = jTask.Content
	}
	if len(jTask.Link) > 0 {
		updatBson["link"] = jTask.Link
	}
	updatBson["modified_time"] = time.Now()

	storage_updateById(id, updatBson)

	c.JSON(http.StatusOK, gin.H{"code": 1, "message": "success"})
}

func parseAccounts(accountInfo string) gin.Accounts {
	accounts := gin.Accounts{}
	if accountInfo != "" {
		parts := strings.Split(accountInfo, ":")
		if len(parts) == 2 {
			accounts[parts[0]] = parts[1]
		} else {
			accounts[parts[0]] = ""
		}
	}
	return accounts
}

func setRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// saytodo
	saytodo := r.Group("saytodo", gin.BasicAuth(parseAccounts(config.AccountInfo)))
	{
		saytodo.POST("/tasks", addNewTask)
		saytodo.GET("/tasks", queryTaskList)
		saytodo.GET("/tasks/:id", queryTask)
		saytodo.PUT("/tasks/:id", updateTask)
	}

	return r
}

// https://zhuanlan.zhihu.com/p/37844072
// ToBson ...
func ToBson(r interface{}) bson.M {
	result := make(bson.M)
	v := reflect.ValueOf(r)
	t := reflect.TypeOf(r)

	for i := 0; i < v.NumField(); i++ {
		filed := v.Field(i)
		tag := t.Field(i).Tag
		key := tag.Get("bson")
		if key == "" || key == "-" || key == "_id" {
			continue
		}
		keys := strings.Split(key, ",")
		if len(keys) > 0 {
			key = keys[0]
		}
		// TODO: 处理字段嵌套问题
		switch filed.Kind() {
		case reflect.Int, reflect.Int64:
			v := filed.Int()
			if v != 0 {
				result[key] = v
			}
		case reflect.String:
			v := filed.String()
			if v != "" {
				result[key] = v
			}
		case reflect.Bool:
			result[key] = filed.Bool()
		case reflect.Ptr:

		case reflect.Float64:
			v := filed.Float()
			if v != 0 {
				result[key] = v
			}
		case reflect.Float32:
			v := filed.Float()
			if v != 0 {
				result[key] = v
			}
		default:
		}
	}
	return result
}

func main() {
	configFile = DEFAULT_CONFIG_FILE

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		fmt.Printf("Parse config [%s] error: %s\n", configFile, err)
		return
	}

	g_mgoSession = connectMongo(config.MongoUrl)
	defer closeMongo(g_mgoSession)

	r := setRouter()

	r.Run(":8080")
}
