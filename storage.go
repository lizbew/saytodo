package main

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"log"
)

var (
	g_mgoSession *mgo.Session
)

func connectMongo(mongoUrl string) *mgo.Session {
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		log.Fatal("connection mongodb failed!")
		panic(err)
	}
	return session
}

func closeMongo(s *mgo.Session) {
	s.Close()
}

func getTaskCollection() *mgo.Collection {
	return g_mgoSession.DB(MONGO_DB_TASK).C(MONGO_COLL_TASKS)
}

func storage_selectOne(id string) (*TodoTask, error) {
	mgoCollection_Task := getTaskCollection()
	objectId := bson.ObjectIdHex(id)
	var task TodoTask
	if err := mgoCollection_Task.FindId(objectId).One(&task); err != nil {
		return nil, err
	}

	return &task, nil
}

func storage_insertOne(task *TodoTask) error {
	mgoCollection_Task := getTaskCollection()
	return mgoCollection_Task.Insert(task)
}

func storage_selectAll(taskListResponse *TaskListResponse) {
	mgoCollection_Task := getTaskCollection()

	count, err := mgoCollection_Task.Count()
	if err != nil {
		log.Println("query task count failed")
		count = 0
	}

	var taskList []TodoTask
	mgoCollection_Task.Find(nil).All(&taskList)
	taskListResponse.Count = count
	taskListResponse.List = taskList

}

func storage_updateById(id string, updatBson bson.M) {
	objectId := bson.ObjectIdHex(id)

	mgoCollection_Task := getTaskCollection()
	mgoCollection_Task.Update(bson.M{"_id": objectId}, bson.M{"$set": updatBson})
}
